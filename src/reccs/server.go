package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
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
	var dataDir string

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
		var paramMsgs []*Message
		var paramCount int
		var collection *Collection
		var parameters []interface{}
		var message *Message

		paramMsgs = msgs[1:]
		paramCount = len(paramMsgs)

		if len(paramMsgs) > len(command.Parameters) {
			fmt.Fprint(conn, "-Incorrect command parameters\r\n")
			return
		}

		for i, param := range command.Parameters {
			if param.Required && i >= paramCount {
				fmt.Fprint(conn, "-Incorrect command parameters\r\n")
				return
			}
			message = paramMsgs[i]
			switch param.Type {
			case "collection":
				value, err := message.Str()
				if err != nil || collection != nil {
					fmt.Fprint(conn, "-Incorrect command parameters\r\n")
					return
				}
				collection = CreateCollection(value, s.DataPath)
			case "string":
				value, err := message.Str()
				if err != nil {
					fmt.Fprint(conn, "-Incorrect command parameters\r\n")
					return
				}
				parameters = append(parameters, value)
			case "integer":
				value, err := message.Int()
				if err != nil {
					fmt.Fprint(conn, "-Incorrect command parameters\r\n")
					return
				}
				parameters = append(parameters, value)
			case "binary":
				value, err := message.Bytes()
				if err != nil {
					fmt.Fprint(conn, "-Incorrect command parameters\r\n")
					return
				}
				parameters = append(parameters, value)
			}
		}
		command.Callback(parameters, conn, collection)
	} else {
		conn.Write([]byte("-Unrecognized command\r\n"))
	}
	return

	// CREATE DELETE GET ADD HEAD TAIL - collection data commands
	// CSET CGET - config setter and getter
	// TSHEAD TSTAIL - timestamps
	// PING - server ping
	var command string
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
				//maxItems, _ := strconv.Atoi(value)
				//enforceMaxItems(collection, maxItems)
			}
		} else {
			conn.Write([]byte("-Config setting error\r\n"))
		}
	case "CGET":
		key, _ := msgs[2].Str()
		value, _ := getConfig(collection, key)
		fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(value), value)
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
