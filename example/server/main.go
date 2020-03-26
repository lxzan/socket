package main

import (
	"github.com/lxzan/socket"
	"io/ioutil"
)

func main() {
	s := socket.NewServer()

	s.OnConnect = func(client *socket.Client) {
		for {
			select {
			case msg := <-client.OnMessage:
				err := ioutil.WriteFile(`C:\Users\Caster\Desktop\WorkPlace\socket\runtime\test.png`, msg.Body, 0755)
				println(&err)
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
