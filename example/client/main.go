package main

import (
	"context"
	"github.com/lxzan/socket"
)

func main() {
	println("start...")
	for i := 0; i < 1; i++ {
		go func() {
			client, err := socket.Dial(context.Background(), "127.0.0.1:9090", &socket.Option{
				//CryptoAlgo: socket.CryptoAlgo_RsaAes,
				//PublicKey:  "example/cert/pub.pem",
			})
			if err != nil {
				println(err.Error())
				return
			}

			for j := 0; j < 1; j++ {
				_, err = client.Send(socket.TextMessage, &socket.Message{Body: []byte("hello, ")})
				if err != nil {
					println(err.Error())
				}
			}

			for {
				select {
				case msg := <-client.OnMessage:
					println(string(msg.Body))
				case err := <-client.OnError:
					println(err.Error())
					return
				}
			}
		}()

	}

	select {}
}
