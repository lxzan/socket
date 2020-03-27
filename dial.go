package socket

import (
	"net"
	"time"
)

type DialOption struct {
	serverSide       bool
	CompressAlgo                   // default gzip
	CryptoAlgo                     // default RSA-AES
	HandshakeTimeout time.Duration // default 5s
	PrivateKey       string        // pem file path
	PublicKey        string        // pem file path
	CompressMinsize  int           // compress data when dataLength>=CompressMinsize
}

func Dial(addr string, opt *DialOption) (*Client, error) {
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
		if err := client.sendHandshake(); err != nil {
			return nil, err
		}
	}

	return client, nil
}
