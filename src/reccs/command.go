package main

import (
	"net"
	"os"
	"path/filepath"
)

/*
	Parameter boundaries do not include the command or collection

	This has one parameter if Command.CollectionRequired is true
	otherwise it is considered two.
	> ADD foobar "this item"
*/
type Command struct {
	Name        string
	Parameters  []CommandParameter
	HelpMessage string
	Callback    func([]interface{}, net.Conn, *Collection)
}

var Commands map[string]Command

func init() {
	Commands = map[string]Command{}
	Commands["create"] = Command{
		Name:        "create",
		HelpMessage: "Create a new collection",
		Parameters: []CommandParameter{
			CommandParameter{
				Name:     "name",
				Type:     "collection",
				Required: true,
			},
		},
		Callback: func(params []interface{}, conn net.Conn, coll *Collection) {
			perms := os.FileMode(0700)
			os.MkdirAll(coll.DataPath, perms)
			os.MkdirAll(coll.ConfigPath, perms)
			coll.SetConfig("maxitems", "100")
			conn.Write([]byte("+OK\r\n"))
		},
	}
	Commands["delete"] = Command{
		Name:        "delete",
		HelpMessage: "Delete an existing collection",
		Parameters: []CommandParameter{
			CommandParameter{
				Name:     "name",
				Type:     "collection",
				Required: true,
			},
		},
		Callback: func(params []interface{}, conn net.Conn, coll *Collection) {
			os.RemoveAll(coll.BasePath)
			conn.Write([]byte("+OK\r\n"))
		},
	}

	Commands["get"] = Command{
		Name:        "get",
		HelpMessage: "Get all items in collection in ascending order",
		Parameters: []CommandParameter{
			CommandParameter{
				Name:     "name",
				Type:     "collection",
				Required: true,
			},
		},
		Callback: func(params []interface{}, conn net.Conn, coll *Collection) {
			files := coll.GetDataFiles()
			streamFiles(files, conn)
		},
	}

	Commands["head"] = Command{
		Name:        "head",
		HelpMessage: "Get the latest item in collection",
		Parameters: []CommandParameter{
			CommandParameter{
				Name:     "name",
				Type:     "collection",
				Required: true,
			},
		},
		Callback: func(params []interface{}, conn net.Conn, coll *Collection) {
			files := coll.GetDataFiles()
			streamFile(files[len(files)-1], conn)
		},
	}

	Commands["tail"] = Command{
		Name:        "tail",
		HelpMessage: "Get the oldest item in collection",
		Parameters: []CommandParameter{
			CommandParameter{
				Name:     "name",
				Type:     "collection",
				Required: true,
			},
		},
		Callback: func(params []interface{}, conn net.Conn, coll *Collection) {
			files := coll.GetDataFiles()
			streamFile(files[0], conn)
		},
	}

	Commands["add"] = Command{
		Name:        "add",
		HelpMessage: "Add item in collection",
		Parameters: []CommandParameter{
			CommandParameter{
				Name:     "name",
				Type:     "collection",
				Required: true,
			},
			CommandParameter{
				Name:     "data",
				Type:     "binary",
				Required: false,
			},
		},
		Callback: func(params []interface{}, conn net.Conn, coll *Collection) {
			filename := timestamp()
			fullFilePath := filepath.Join(coll.DataPath, filename)
			file, err := os.Create(fullFilePath)
			if err != nil {
				conn.Write([]byte("-cannot add item\r\n"))
			} else {
				file.Chmod(os.FileMode(0700))
				if len(params) > 0 {
					file.Write(params[0].([]byte))
				}
				file.Close()
			}
			conn.Write([]byte("+OK\r\n"))
		},
	}

	Commands["ping"] = Command{
		Name:        "ping",
		HelpMessage: "Ping reccs for life",
		Parameters:  []CommandParameter{},
		Callback: func(params []interface{}, conn net.Conn, coll *Collection) {
			conn.Write([]byte("+PONG\r\n"))
		},
	}
}
