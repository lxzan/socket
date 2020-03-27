### Universe Socket Communication Protocol

- Server
```go
    package main
    
    import (
    	"github.com/lxzan/socket"
    )
    
    func main() {
    	s := socket.NewServer(&socket.DialOption{
    		CryptoAlgo: socket.CryptoAlgo_RsaAes,
    		PrivateKey: "example/cert/prv.pem",
    	})
    
    	s.OnConnect = func(client *socket.Client) {
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
    
    	if err := s.Run(":9090"); err != nil {
    		println(err.Error())
    	}
    }

```


- Client
```go
package main

import (
	"github.com/lxzan/socket"
)

func main() {
	client, err := socket.Dial("127.0.0.1:9090", &socket.DialOption{
		CryptoAlgo: socket.CryptoAlgo_RsaAes,
		PublicKey:  "example/cert/pub.pem",
	})
	if err != nil {
		println(err.Error())
		return
	}

	_, err = client.WriteMessage(socket.TextMessage, nil, []byte("hello world!"))
	if err != nil {
		println(err.Error())
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
}

```
