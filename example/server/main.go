package main

import (
	"github.com/lxzan/socket"
	"io/ioutil"
)

func main() {
	s := socket.NewServer(nil)

	s.OnConnect = func(client *socket.Client) {
		for {
			select {
			case msg := <-client.OnMessage:
				err := ioutil.WriteFile(`/Users/Caster/MyWork/socket/runtime/test.jpg`, msg.Body, 0755)
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
