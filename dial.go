package socket

import (
	"context"
	"net"
	"time"
)

type DialOption struct {
	serverSide      bool
	CompressAlgo                  // default gzip
	CryptoAlgo                    // default RSA-AES
	PrivateKey      string        // pem file path
	PublicKey       string        // pem file path
	CompressMinsize int           // compress data when dataLength>=CompressMinsize
	IoTimeout       time.Duration // io timeout
}

func Dial(ctx context.Context, addr string, opt *DialOption) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	client, err := newClientSideClient(conn, opt)
	if err != nil {
		return nil, err
	}

	go client.handleMessage()

	if client.Option.CryptoAlgo != CryptoAlgo_NoCrypto {
		if err := client.sendHandshake(ctx); err != nil {
			return nil, err
		}
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			client.Send(PingMessage, nil)
		}
	}()

	return client, nil
}
