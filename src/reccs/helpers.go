package main

import (
	"strconv"
	"time"
)

func timestamp() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}
