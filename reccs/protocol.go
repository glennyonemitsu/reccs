package main

import (
	"fmt"
	"github.com/fzzy/radix/redis/resp"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

func handleRequest(conn net.Conn, data []byte) {
	var collection string
	var collectionDir string
	var dataDir string
	var metaDir string
	var perms os.FileMode

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

	// CREATE DELETE GET ADD HEAD TAIL - collection data commands
	// MSET MGET - collection meta data
	// TNEW TOLD - timestamps
	// PING - server ping
	switch command {
	case "TNEW":
		files := getDirFiles(dataDir)
		timestamp := filepath.Base(files[len(files)-1])
		streamIntegers(splitTimestamp(timestamp), conn)
	case "TOLD":
		files := getDirFiles(dataDir)
		timestamp := filepath.Base(files[0])
		streamIntegers(splitTimestamp(timestamp), conn)
	case "PING":
		conn.Write([]byte("+PONG\r\n"))
	case "HEAD":
		files := getDirFiles(dataDir)
		file := make([]string, 1)
		file[0] = files[len(files)-1]
		streamFiles(file, conn)
	case "TAIL":
		files := getDirFiles(dataDir)
		file := make([]string, 1)
		file[0] = files[0]
		streamFiles(file, conn)
	case "GET":
		files := getDirFiles(dataDir)
		streamFiles(files, conn)
	case "CREATE":
		os.MkdirAll(dataDir, perms)
		os.MkdirAll(metaDir, perms)
		conn.Write([]byte("+OK\r\n"))
	case "DELETE":
		os.RemoveAll(collectionDir)
		conn.Write([]byte("+OK\r\n"))
	case "ADD":
		filename := timestamp()
		fullFilePath := filepath.Join(dataDir, filename)
		file, err := os.Create(fullFilePath)
		if err != nil {
			conn.Write([]byte("-cannot create file\r\n"))
		} else {
			file.Chmod(perms)
			entryData, _ := msgs[2].Bytes()
			file.Write(entryData)
			file.Close()
		}
		conn.Write([]byte("+OK\r\n"))
	default:
		conn.Write([]byte("-unrecognized command\r\n"))
	}

}

func isValidcollection(collection string) bool {
	return true
}

func streamFiles(files []string, w io.Writer) {
	var info os.FileInfo

	fmt.Fprintf(w, "*%d\r\n", len(files))
	for _, f := range files {
		fh, _ := os.Open(f)
		info, _ = os.Stat(f)
		remaining := info.Size()
		fmt.Fprintf(w, "$%d\r\n", remaining)
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
			w.Write(bytes)
		}
		fmt.Fprintf(w, "\r\n")
		fh.Close()
	}
}

func streamIntegers(ints []int64, w io.Writer) {
	fmt.Fprintf(w, "*%d\r\n", len(ints))
	for _, i := range ints {
		fmt.Fprintf(w, ":%d\r\n", i)
	}
}

func getDirFiles(dirPath string) []string {
	var files []string
	var walker func(path string, info os.FileInfo, err error) error

	walker = func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	}
	filepath.Walk(dirPath, walker)
	return files

}

func splitTimestamp(timestamp string) []int64 {
	seconds, _ := strconv.ParseInt(timestamp[0:10], 10, 64)
	nseconds, _ := strconv.ParseInt(timestamp[10:], 10, 64)
	return []int64{seconds, nseconds}
}
