package main

import (
	"fmt"
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
			coll.SetConfig("maxitems", 100)
			coll.SetConfig("maxage", 0)
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
		HelpMessage: "Get all items including timestamps in the collection in ascending order",
		Parameters: []CommandParameter{
			CommandParameter{
				Name:     "name",
				Type:     "collection",
				Required: true,
			},
		},
		Callback: func(params []interface{}, conn net.Conn, coll *Collection) {
			files := coll.GetDataFiles()
			fmt.Fprintf(conn, "*%d\r\n", len(files))
			for _, file := range files {
				sec, nsec := splitTimestamp(filepath.Base(file))
				fmt.Fprintf(conn, "*3\r\n:%d\r\n:%d\r\n", sec, nsec)
				streamFile(file, conn)
			}
		},
	}

	Commands["gett"] = Command{
		Name:        "gett",
		HelpMessage: "Get all timestamps in the collection in ascending order",
		Parameters: []CommandParameter{
			CommandParameter{
				Name:     "name",
				Type:     "collection",
				Required: true,
			},
		},
		Callback: func(params []interface{}, conn net.Conn, coll *Collection) {
			files := coll.GetDataFiles()
			fmt.Fprintf(conn, "*%d\r\n", len(files))
			for _, file := range files {
				sec, nsec := splitTimestamp(filepath.Base(file))
				fmt.Fprintf(conn, "*2\r\n:%d\r\n:%d\r\n", sec, nsec)
			}
		},
	}

	Commands["getd"] = Command{
		Name:        "getd",
		HelpMessage: "Get all items' data in the collection in ascending order",
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
		HelpMessage: "Get the latest item in the collection",
		Parameters: []CommandParameter{
			CommandParameter{
				Name:     "name",
				Type:     "collection",
				Required: true,
			},
		},
		Callback: func(params []interface{}, conn net.Conn, coll *Collection) {
			files := coll.GetDataFiles()
			file := files[len(files)-1]
			sec, nsec := splitTimestamp(filepath.Base(file))
			fmt.Fprintf(conn, "*3\r\n:%d\r\n:%d\r\n", sec, nsec)
			streamFile(file, conn)
		},
	}

	Commands["headt"] = Command{
		Name:        "headt",
		HelpMessage: "Get the latest item's timestamp in the collection",
		Parameters: []CommandParameter{
			CommandParameter{
				Name:     "name",
				Type:     "collection",
				Required: true,
			},
		},
		Callback: func(params []interface{}, conn net.Conn, coll *Collection) {
			files := coll.GetDataFiles()
			file := files[len(files)-1]
			sec, nsec := splitTimestamp(filepath.Base(file))
			fmt.Fprintf(conn, "*2\r\n:%d\r\n:%d\r\n", sec, nsec)
		},
	}

	Commands["headd"] = Command{
		Name:        "headd",
		HelpMessage: "Get the latest item data in the collection",
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
		HelpMessage: "Get the oldest item in the collection",
		Parameters: []CommandParameter{
			CommandParameter{
				Name:     "name",
				Type:     "collection",
				Required: true,
			},
		},
		Callback: func(params []interface{}, conn net.Conn, coll *Collection) {
			files := coll.GetDataFiles()
			file := files[0]
			sec, nsec := splitTimestamp(filepath.Base(file))
			fmt.Fprintf(conn, "*3\r\n:%d\r\n:%d\r\n", sec, nsec)
			streamFile(file, conn)
		},
	}

	Commands["tailt"] = Command{
		Name:        "tailt",
		HelpMessage: "Get the oldest item's timestamp in the collection",
		Parameters: []CommandParameter{
			CommandParameter{
				Name:     "name",
				Type:     "collection",
				Required: true,
			},
		},
		Callback: func(params []interface{}, conn net.Conn, coll *Collection) {
			files := coll.GetDataFiles()
			file := files[0]
			sec, nsec := splitTimestamp(filepath.Base(file))
			fmt.Fprintf(conn, "*2\r\n:%d\r\n:%d\r\n", sec, nsec)
		},
	}

	Commands["taild"] = Command{
		Name:        "taild",
		HelpMessage: "Get the oldest item data in the collection",
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
				coll.EnforceCapacity()
				conn.Write([]byte("+OK\r\n"))
			}
		},
	}

	Commands["cget"] = Command{
		Name:        "cget",
		HelpMessage: "Get configuration value",
		Parameters: []CommandParameter{
			CommandParameter{
				Name:     "name",
				Type:     "collection",
				Required: true,
			},
			CommandParameter{
				Name:     "key",
				Type:     "string",
				Required: true,
			},
		},
		Callback: func(params []interface{}, conn net.Conn, coll *Collection) {
			value, err := coll.GetConfig(params[0].(string))
			if err != nil {
				conn.Write([]byte("-Error getting configuration value\r\n"))
			} else {
				fmt.Fprintf(conn, ":%d\r\n", value)
			}
		},
	}

	Commands["cset"] = Command{
		Name:        "cset",
		HelpMessage: "Set configuration value",
		Parameters: []CommandParameter{
			CommandParameter{
				Name:     "name",
				Type:     "collection",
				Required: true,
			},
			CommandParameter{
				Name:     "key",
				Type:     "string",
				Required: true,
			},
			CommandParameter{
				Name:     "value",
				Type:     "integer",
				Required: true,
			},
		},
		Callback: func(params []interface{}, conn net.Conn, coll *Collection) {
			if err := coll.SetConfig(params[0].(string), params[1].(int64)); err != nil {
				conn.Write([]byte("-Error getting configuration value\r\n"))
			} else {
				coll.EnforceCapacity()
				fmt.Fprint(conn, "+OK\r\n")
			}
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
