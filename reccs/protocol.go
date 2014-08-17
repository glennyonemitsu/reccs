package main

import (
	"fmt"
	"github.com/fzzy/radix/redis/resp"
	"net"
	"os"
	"path/filepath"
)

func handleRequest(conn net.Conn, data []byte) {
	var collection string
	var collectionDir string
	var dataDir string
	var metaDir string
	var perms os.FileMode
	var getcollectionFilenames func(path string, info os.FileInfo, err error) error
	var fileinfos []os.FileInfo

	message, err := resp.NewMessage(data)
	if err != nil {
		return
	}
	msgs, err := message.Array()
	if err != nil {
		return
	}
	command, _ := msgs[0].Str()
	if len(msgs) > 1 {
		collection, _ = msgs[1].Str()
		perms = os.FileMode(0700)
		collectionDir = filepath.Join(DataDir, collection)
		dataDir = filepath.Join(collectionDir, "data")
		metaDir = filepath.Join(collectionDir, "meta")
	} else {
		collection = ""
	}

	// CREATE GET ADD HEAD TAIL MSET MGET
	switch command {
	case "GET":
		getcollectionFilenames = func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			fileinfos = append(fileinfos, info)
			return nil
		}
		filepath.Walk(dataDir, getcollectionFilenames)

		fmt.Fprintf(conn, "*%d\r\n", len(fileinfos))
		for _, info := range fileinfos {
			remaining := info.Size()
			fmt.Fprintf(conn, "$%d\r\n", remaining)
			fullFilePath := filepath.Join(dataDir, info.Name())
			fh, _ := os.Open(fullFilePath)
			var bytes []byte
			for remaining > 0 {
				if remaining < 1024 {
					bytes = make([]byte, remaining)
					remaining = 0
				} else {
					bytes = make([]byte, 1024)
					remaining -= 1024
				}
				fh.Read(bytes)
				conn.Write(bytes)
			}
			fmt.Fprintf(conn, "\r\n")
			fh.Close()
		}

	case "CREATE":
		os.MkdirAll(dataDir, perms)
		os.MkdirAll(metaDir, perms)
	case "ADD":
		filename := timestamp()
		fmt.Println(filename)
		fullFilePath := filepath.Join(dataDir, filename)
		file, err := os.Create(fullFilePath)
		if err != nil {
			conn.Write([]byte("+ERR cannot create file\r\n"))
		} else {
			file.Chmod(perms)
			entryData, _ := msgs[2].Bytes()
			file.Write(entryData)
			file.Close()
		}

	}
	conn.Write([]byte("+OK\r\n"))

}

func isValidcollection(collection string) bool {
	return true
}
