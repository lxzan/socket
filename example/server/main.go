package main

import (
	"github.com/lxzan/socket"
)

func main() {
	s := socket.NewServer(&socket.DialOption{
		//CryptoAlgo: socket.CryptoAlgo_RsaAes,
		//PrivateKey: "example/cert/prv.pem",
	})

	s.Run(":9090", func(client *socket.Client) {
		for {
			select {
			case msg := <-client.OnMessage:
				println(string(msg.Body))
				client.Send(socket.TextMessage, &socket.Message{Body: []byte("world!")})
			case err := <-client.OnError:
				println(err.Error())
				return
			}
		}
	})
}
