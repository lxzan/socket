package main

import (
	"github.com/lxzan/socket"
)

func main() {
	println("start...")
	client, err := socket.Dial("127.0.0.1:9090", &socket.DialOption{
		CryptoAlgo: socket.CryptoAlgo_RsaAes,
		PublicKey:  "example/cert/pub.pem",
	})
	if err != nil {
		println(err.Error())
		return
	}

	//p := "/Users/Caster/Downloads/E7D96742A2D38033BFBE46FFF33A92B1.jpg"
	//f, _ := ioutil.ReadFile(p)
	//_, err = client.WriteMessage(socket.BinaryMessage, nil, f)

	client.WriteMessage(socket.TextMessage, nil, []byte("hello world!"))
	if err != nil {
		println(err.Error())
	}

	for {
		select {
		case msg := <-client.OnMessage:
			println(string(msg.Body))
		case err := <-client.OnError:
			println(err.Error())
		}
	}
}
