package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

func timestamp() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
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
