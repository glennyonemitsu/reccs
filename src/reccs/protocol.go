package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

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
