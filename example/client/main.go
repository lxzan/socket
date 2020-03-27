package main

import "github.com/lxzan/socket"

func main() {
	for i := 0; i < 100; i++ {
		go func() {
			client, err := socket.Dial("127.0.0.1:9090", &socket.DialOption{
				//CryptoAlgo: socket.CryptoAlgo_RsaAes,
				//PublicKey:  "example/cert/pub.pem",
			})
			if err != nil {
				println(err.Error())
				return
			}

			for j := 0; j < 10000; j++ {
				_, err = client.WriteMessage(socket.TextMessage, nil, []byte("hello world!"))
				if err != nil {
					println(err.Error())
				}
			}
		}()
	}

	select {}
}
