package main

import "github.com/lxzan/socket"

func main() {
	s := socket.NewServer()

	s.OnConnect = func(client *socket.Client) {
		client.OnMessage = func(msg *socket.Message) {
			println(string(msg.Body))
		}

		client.OnError = func(err error) {
			println(err.Error())
		}

		client.HandleMessage()
	}

	if err := s.Run(":9090"); err != nil {
		println(err.Error())
	}
}
