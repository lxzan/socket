package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)


func main() {
	listener, err := net.Listen("tcp", ":8080")
	CatchFatal(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			println(fmt.Sprintf("close: %s", err.Error()))
		} else {
			err = handleClient(conn)
			if err != nil {
				println(fmt.Sprintf("close: %s", err.Error()))
			}
		}
	}
}

func CatchFatal(err error) {
	if err != nil {
		panic(err)
	}
}

func handleClient(conn net.Conn) error {
	defer conn.Close()

	var buf = bytes.NewBufferString("")
	for {
		pack := make([]byte, 1024)
		_, err := conn.Read(pack)
		if err != nil {
			return err
		}
		pl := PackLength(pack)
		buf.Write(pack[:pl])

		var rl uint32 = 4
		var rlb = true
		for uint32(buf.Len()) >= rl {
			var p = make([]byte, rl)
			_, err = buf.Read(p)
			if err != nil {
				return err
			}

			if rlb {
				rl = binary.LittleEndian.Uint32(p)
				rlb = false
			} else {
				Read(p)
				rl = 4
				rlb = true
			}
		}
	}
}

func PackLength(msg []byte) int {
	var i int
	n := len(msg)
	for i = n - 1; i >= 0; i-- {
		if msg[i] != 0 {
			break
		}
	}
	return i + 1
}

func Read(msg []byte) {
	println("rec: " + string(msg))
}
