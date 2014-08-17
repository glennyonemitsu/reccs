package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

const DataDir = "/srv/dw-data/"
const ReadBufferSize = 64

func timestamp() string {
	return string(time.Now().UnixNano())
}

func checkDataDir(dir string) (bool, string) {
	var result bool
	var message string

	result = true
	message = ""

	file, err := os.Open(dir)
	defer file.Close()

	if err != nil {
		message = fmt.Sprintf("Error opening data directory: %s\n", DataDir)
		result = false
		return result, message
	}
	info, err := file.Stat()
	if !info.IsDir() {
		message = fmt.Sprintf("Not a directory: %s\n", DataDir)
		result = false
		return result, message
	}

	// TODO permission checks

	return result, message
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	var data []byte
	var dataB []byte
	var readSize int
	var totalLength int

	totalLength = 0
	readSize = ReadBufferSize
	data = make([]byte, readSize)
	for {
		length, err := conn.Read(data)
		if err != nil {
			break
		}
		totalLength += length
		if length == readSize {
			readSize *= 2
			dataB = make([]byte, readSize)
			dataB = data
			data = make([]byte, readSize)
			data = dataB
		} else {
			break
		}
	}
	data = data[0:totalLength]
	handleRequest(conn, data)
}

func main() {
	if success, message := checkDataDir(DataDir); !success {
		fmt.Println(message)
		os.Exit(1)
	}

	socket, err := net.Listen("tcp", "localhost:9990")
	if err != nil {
		fmt.Println("Cannot setup socket")
		os.Exit(2)
	}
	for {
		conn, err := socket.Accept()
		if err != nil {
			// handle error
			continue
		}
		go handleConnection(conn)
	}
}
