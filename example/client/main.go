package main

import (
	"context"
	"fmt"
	"github.com/lxzan/socket"
	"time"
)

func main() {
	println("start...")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	client, err := socket.Dial(ctx, "127.0.0.1:9090", &socket.Option{
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
			if err = client.Send(socket.TextMessage, &socket.Message{Body: []byte(str)}); err != nil {
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
