package socket

import (
	"context"
	"errors"
	"net"
	"time"
)

type DialOption struct {
	serverSide       bool
	CompressAlgo                   // default gzip
	CryptoAlgo                     // default RSA-AES
	PrivateKey       string        // pem file path
	PublicKey        string        // pem file path
	CompressMinsize  int           // compress data when dataLength>=CompressMinsize
	HeartbeatTimeout time.Duration // io timeout
	PingInterval     time.Duration // ping interval
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
		if err := sendHandshake(ctx, client); err != nil {
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

// client side
func sendHandshake(ctx context.Context, client *Client) error {
	var key = []byte(Alphabet.Generate(16))
	encryptKey, err := client.asymmetric.Encrypt(key)
	if err != nil {
		return err
	}

	if _, err := client.Send(HandshakeMessage, &Message{Body: encryptKey}); err != nil {
		return err
	}

	for {
		select {
		case <-client.onHandshake:
			encoder, err := NewAesCrypto(key)
			if err != nil {
				return err
			}
			client.aes = encoder
			return nil
		case <-ctx.Done():
			return errors.New("handshake timeout")
		}
	}
}
