package main

import (
	"fmt"
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
	var configDir string
	var perms os.FileMode

	message, err := NewMessage(data)
	if err != nil {
		conn.Write([]byte("-Server error\r\n"))
		return
	}
	msgs, err := message.Array()
	if err != nil {
		conn.Write([]byte("-Server error\r\n"))
		return
	}
	command, _ := msgs[0].Str()
	if len(msgs) > 1 {
		collection, _ = msgs[1].Str()
		perms = os.FileMode(0700)
		collectionDir = filepath.Join(*DataPath, collection)
		dataDir = filepath.Join(collectionDir, "data")
		configDir = filepath.Join(collectionDir, "config")
	} else {
		collection = ""
	}

	// CREATE DELETE GET ADD HEAD TAIL - collection data commands
	// CSET CGET - config setter and getter
	// TSHEAD TSTAIL - timestamps
	// PING - server ping
	switch command {
	case "TSHEAD":
		files := getDirFiles(dataDir)
		timestamp := filepath.Base(files[len(files)-1])
		streamIntegers(splitTimestamp(timestamp), conn)
	case "TSTAIL":
		files := getDirFiles(dataDir)
		timestamp := filepath.Base(files[0])
		streamIntegers(splitTimestamp(timestamp), conn)
	case "CSET":
		key, _ := msgs[2].Str()
		value, _ := msgs[3].Str()
		if results := setConfig(collection, key, value); results {
			conn.Write([]byte("+OK\r\n"))
			if key == "maxitems" {
				maxItems, _ := strconv.Atoi(value)
				enforceMaxItems(collection, maxItems)
			}
		} else {
			conn.Write([]byte("-Config setting error\r\n"))
		}
	case "CGET":
		key, _ := msgs[2].Str()
		value, _ := getConfig(collection, key)
		fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(value), value)
	case "PING":
		conn.Write([]byte("+PONG\r\n"))
	case "HEAD":
		files := getDirFiles(dataDir)
		file := files[len(files)-1]
		streamFile(file, conn)
	case "TAIL":
		files := getDirFiles(dataDir)
		file := files[0]
		streamFile(file, conn)
	case "GET":
		files := getDirFiles(dataDir)
		streamFiles(files, conn)
	case "CREATE":
		os.MkdirAll(dataDir, perms)
		os.MkdirAll(configDir, perms)
		setConfig(collection, "maxitems", "100")
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
		configMaxItems, _ := getConfig(collection, "maxitems")
		maxItems, _ := strconv.Atoi(configMaxItems)
		enforceMaxItems(collection, maxItems)
	default:
		conn.Write([]byte("-unrecognized command\r\n"))
	}

}

func streamFiles(files []string, w io.Writer) {
	fmt.Fprintf(w, "*%d\r\n", len(files))
	for _, f := range files {
		streamFile(f, w)
	}
}

func streamFile(file string, w io.Writer) {
	var bytes []byte
	var remaining int64

	fh, _ := os.Open(file)
	info, _ := fh.Stat()
	remaining = info.Size()
	fmt.Fprintf(w, "$%d\r\n", remaining)
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

func streamIntegers(ints []int64, w io.Writer) {
	fmt.Fprintf(w, "*%d\r\n", len(ints))
	for _, i := range ints {
		streamInteger(i, w)
	}
}

func streamInteger(value int64, w io.Writer) {
	fmt.Fprintf(w, ":%d\r\n", value)
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

func getConfig(collection string, key string) (string, error) {
	var data []byte
	configFile := filepath.Join(*DataPath, collection, "config", key)
	fh, err := os.Open(configFile)
	defer fh.Close()
	if err != nil {
		return "", err
	}
	stat, err := fh.Stat()
	if err != nil {
		return "", err
	}
	data = make([]byte, stat.Size())
	fh.Read(data)
	return string(data), nil
}

func setConfig(collection string, key string, value string) bool {
	configFile := filepath.Join(*DataPath, collection, "config", key)
	fh, err := os.Create(configFile)
	defer fh.Close()
	if err != nil {
		return false
	}
	_, err = fh.Write([]byte(value))
	if err != nil {
		return false
	}
	return true
}

func splitTimestamp(timestamp string) []int64 {
	seconds, _ := strconv.ParseInt(timestamp[0:10], 10, 64)
	nseconds, _ := strconv.ParseInt(timestamp[10:], 10, 64)
	return []int64{seconds, nseconds}
}

func enforceMaxItems(collection string, max int) {
	if max < 1 {
		return
	}
	dataPath := filepath.Join(*DataPath, collection, "data")
	files := getDirFiles(dataPath)
	if len(files) > max {
		oldFiles := files[0 : len(files)-max]
		for _, file := range oldFiles {
			os.Remove(file)
		}
	}
}
