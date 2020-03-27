package main

import (
	"github.com/lxzan/socket"
	"io/ioutil"
)

func main() {
	println("start...")
	client, err := socket.Dial("127.0.0.1:9090")
	if err != nil {
		println(err.Error())
		return
	}

	p := "/Users/Caster/Downloads/UYN`L`WOG$HXB7YA@C(}48F.jpg"
	f, _ := ioutil.ReadFile(p)
	//client.WriteMessage(socket.BinaryMessage, nil, []byte("hello"))
	_, err = client.WriteMessage(socket.BinaryMessage, nil, f)
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
