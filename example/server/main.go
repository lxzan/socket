package main

import (
	"github.com/lxzan/socket"
)

func main() {
	s := socket.NewServer(nil)

	s.OnConnect = func(client *socket.Client) {
		for {
			select {
			case msg := <-client.OnMessage:
				println(string(msg.Body))
			case err := <-client.OnError:
				println(err.Error())
				return
			}
		}
	}

	if err := s.Run(":9090"); err != nil {
		println(err.Error())
	}
}
