package main

import (
	"fmt"
	"github.com/fzzy/radix/redis/resp"
	"net"
	"os"
	"path/filepath"
)

func handleRequest(conn net.Conn, data []byte) {
	var channel string

	message, err := resp.NewMessage(data)
	if err != nil {
		return
	}
	msgs, err := message.Array()
	if err != nil {
		return
	}
	command, _ := msgs[0].Str()
	if len(msgs) > 1 {
		channel, _ = msgs[1].Str()
	} else {
		channel = ""
	}

	// GET ADD HEAD TAIL SET CREATE
	switch command {
	case "GET":
		fmt.Println("getting")
	case "CREATE":
		perms := os.FileMode(0700)
		channelDir := filepath.Join(DataDir, channel)
		dataDir := filepath.Join(channelDir, "data")
		metaDir := filepath.Join(channelDir, "meta")
		os.MkdirAll(dataDir, perms)
		os.MkdirAll(metaDir, perms)
	}
	conn.Write([]byte("+OK\r\n"))

}

func isValidChannel(channel string) bool {
	return true
}
