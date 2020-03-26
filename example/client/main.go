package main

import (
	"github.com/lxzan/socket"
	"time"
)

func main() {
	client, err := socket.Dial("127.0.0.1:9090")
	if err != nil {
		println(err.Error())
		return
	}

	client.OnMessage = func(msg *socket.Message) {

	}

	client.OnError = func(err error) {

	}

	go client.HandleMessage()

	client.WriteMessage(socket.TextMessage, nil, []byte("Hello"))
	time.Sleep(3 * time.Second)
}
