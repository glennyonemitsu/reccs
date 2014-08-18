package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func checkDataPath(dir string) (bool, string) {
	var result bool
	var message string

	result = true
	message = ""

	file, err := os.Open(dir)
	defer file.Close()

	if err != nil {
		message = fmt.Sprintf("Error opening data directory: %s\n", dir)
		result = false
		return result, message
	}
	info, err := file.Stat()
	if !info.IsDir() {
		message = fmt.Sprintf("Not a directory: %s\n", dir)
		result = false
		return result, message
	}

	// TODO permission checks

	return result, message
}

func timestamp() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}
