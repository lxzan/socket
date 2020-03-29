package main

import (
	"github.com/lxzan/socket"
)

func main() {
	s := socket.NewServer(&socket.Option{
		//CryptoAlgo: socket.CryptoAlgo_RsaAes,
		//PrivateKey: "example/cert/prv.pem",
	})

	s.Run(":9090", func(client *socket.Conn) {
		for {
			select {
			case <-client.PingTicker.C:
				client.Ping()
			case msg := <-client.OnMessage:
				println(string(msg.Body))
				client.Send(socket.TextMessage, &socket.Message{Body: []byte("rec: " + string(msg.Body))})
			case err := <-client.OnError:
				println(err.Error())
				return
			}
		}
	})
}
