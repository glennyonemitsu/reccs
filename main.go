package main

import (
	"flag"
	"fmt"
	"github.com/glennyonemitsu/reccs/reccs"
	"os"
)

var DataPath = flag.String("datapath", "/srv/reccs-data/", "Data storage path")
var BindAddress = flag.String("bind", "localhost:9990", "IP:PORT to bind to")

func init() {
	flag.Parse()
}

func main() {
	server, err := reccs.CreateServer(*BindAddress, *DataPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		server.Serve()
	}
}
