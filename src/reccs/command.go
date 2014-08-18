package main

import (
	"net"
	"os"
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

func (c *Command) containValidParameters() bool {
	return true
}

func (c *Command) Run(msgs []*Message, conn net.Conn) {
	println("running")
	println(msg)

	var collection *Collection
	var parameters []interface{}
	var parametersMatch bool

	parametersMatch = true
	for i, msg := range msgs {
	}

	if parametersMatch {
	}
}

var Commands map[string]Command

func init() {
	Commands = map[string]Command{
		"create": Command{
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
		},
		"delete": Command{
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
		},
	}

}
