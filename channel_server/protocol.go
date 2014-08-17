package main

import (
	"fmt"
	"github.com/fzzy/radix/redis/resp"
	"net"
	"os"
	"path/filepath"
)

func handleRequest(conn net.Conn, data []byte) {
	var channel string
	var channelDir string
	var dataDir string
	var metaDir string
	var perms os.FileMode
	var getChannelFilenames func(path string, info os.FileInfo, err error) error
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
		channel, _ = msgs[1].Str()
		perms = os.FileMode(0700)
		channelDir = filepath.Join(DataDir, channel)
		dataDir = filepath.Join(channelDir, "data")
		metaDir = filepath.Join(channelDir, "meta")
	} else {
		channel = ""
	}

	// CREATE GET ADD HEAD TAIL MSET MGET
	switch command {
	case "GET":
		getChannelFilenames = func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			fileinfos = append(fileinfos, info)
			return nil
		}
		filepath.Walk(dataDir, getChannelFilenames)

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

func isValidChannel(channel string) bool {
	return true
}
