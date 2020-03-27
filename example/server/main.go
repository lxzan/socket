package main

import (
	"fmt"
	"github.com/lxzan/socket"
	"sync/atomic"
	"time"
)

func main() {
	s := socket.NewServer(&socket.DialOption{
		CryptoAlgo: socket.CryptoAlgo_RsaAes,
		PrivateKey: "example/cert/prv.pem",
	})

	var sum int64 = 0
	var t1 int64
	s.Run(":9090", func(client *socket.Client) {
		for {
			select {
			case <-client.OnMessage:
				if t1 == 0 {
					t1 = time.Now().UnixNano()
				}

				num := atomic.AddInt64(&sum, 1)
				if num%10000 == 0 {
					var t2 = time.Now().UnixNano()
					println(fmt.Sprintf("%d, %dms", num, (t2-t1)/1000000))
				}
			case err := <-client.OnError:
				println(err.Error())
				return
			}
		}
	})
}
