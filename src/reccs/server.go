package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Server struct {
	BindAddress string
	DataPath    string
	Commands    map[string]Command
}

func (s *Server) HandleConnection(conn net.Conn) {
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
	s.HandleRequest(conn, data)
}

func (s *Server) HandleRequest(conn net.Conn, data []byte) {
	var collection string
	var collectionDir string
	var dataDir string
	var configDir string
	var perms os.FileMode

	message, err := NewMessage(data)
	if err != nil {
		conn.Write([]byte("-Bad RESP message received\r\n"))
		return
	}

	msgs, err := message.Array()
	if err != nil {
		conn.Write([]byte("-Command error\r\n"))
		return
	}

	commandName, err := msgs[0].Str()
	if err != nil {
		conn.Write([]byte("-Server error\r\n"))
		return
	}

	if command, ok := s.Commands[strings.ToLower(commandName)]; ok {
		// [1:] is to cut off command name in RESP message
		command.Run(msgs[1:], conn)
	} else {
		conn.Write([]byte("-Unrecognized command\r\n"))
	}
	return

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

func (s *Server) Serve() {
	socket, err := net.Listen("tcp", s.BindAddress)
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
		go s.HandleConnection(conn)
	}
}

func (s *Server) checkDataPath() (bool, error) {
	file, err := os.Open(s.DataPath)
	defer file.Close()

	if err != nil {
		return false, err
	}
	info, err := file.Stat()
	if !info.IsDir() {
		return false, err
	}

	// TODO permission checks

	return true, nil
}

func CreateServer(BindAddress, DataPath string) (*Server, error) {
	s := new(Server)
	s.BindAddress = BindAddress
	s.DataPath = DataPath
	s.Commands = Commands
	if success, err := s.checkDataPath(); !success {
		return s, err
	} else {
		return s, nil
	}
}
