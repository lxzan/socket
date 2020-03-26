package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"tcp-demo/helper"
)

func main() {
	conn, err := net.Dial("tcp", ":8080")
	helper.CatchFatal(err)

	for {
		var content string
		fmt.Scanf("%s", &content)
		Send(conn, []byte(content))
	}
}

func Send(conn net.Conn, data []byte) error {
	var tmp = make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(len(data)))
	tmp = append(tmp, data...)
	_, err := conn.Write(tmp)
	return err
}
