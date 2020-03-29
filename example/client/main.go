package main

import (
	"context"
	"fmt"
	"github.com/lxzan/socket"
)

func main() {
	println("start...")

	client, err := socket.Dial(context.Background(), "127.0.0.1:9090", &socket.Option{
		CryptoAlgo: socket.CryptoAlgo_RsaAes,
		PublicKey:  "example/cert/pub.pem",
	})
	if err != nil {
		println(err.Error())
		return
	}

	go func() {
		var str string
		for {
			fmt.Scanf("%s", &str)
			if _, err = client.Send(socket.TextMessage, &socket.Message{Body: []byte(str)}); err != nil {
				client.OnError <- err
				return
			}
		}
	}()

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
