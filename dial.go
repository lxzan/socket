package socket

import (
	"bytes"
	"context"
	"errors"
	"net"
	"time"
)

type Client struct {
	baseClient
	Addr        string
	onHandshake chan bool
}

func newClient(conn net.Conn, opt *Option) (*Client, error) {
	var c = &Client{
		baseClient: baseClient{
			conn:       conn,
			readBuffer: bytes.NewBufferString(""),
			OnMessage:  make(chan *Message, 16),
			OnError:    make(chan error, 16),
			Option:     opt,
		},
		onHandshake: make(chan bool),
	}

	if opt.CryptoAlgo != CryptoAlgo_NoCrypto {
		if opt.PublicKey == "" {
			return nil, errors.New("public key not set")
		}
		if opt.CryptoAlgo == CryptoAlgo_RsaAes {
			rsa, err := NewRsaCrypto(opt.PublicKey, "")
			if err != nil {
				return nil, err
			} else {
				c.asymmetric = rsa
			}
		}
	}
	return c, nil
}

func (this *Client) sendHandshake(ctx context.Context) error {
	var key = []byte(Alphabet.Generate(16, int64(this.Option.Salt)))
	encryptKey, err := this.asymmetric.Encrypt(key)
	if err != nil {
		return err
	}

	this.Send(HandshakeMessage, &Message{Body: encryptKey})
	for {
		select {
		case <-this.onHandshake:
			encoder, err := NewAesCrypto(key)
			if err != nil {
				return err
			}
			this.aes = encoder
			return nil
		case <-ctx.Done():
			return ERR_Timeout.Wrap("handshake timeout")
		}
	}
}

func (this *Client) Reconnect() (*Client, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return Dial(ctx, this.Addr, this.Option)
}

func Dial(ctx context.Context, addr string, opt *Option) (*Client, error) {
	if opt == nil {
		opt = &Option{}
	}
	opt.initialize()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	client, err := newClient(conn, opt)
	if err != nil {
		return nil, err
	}

	client.Addr = addr
	go client.read(func(msg *Message, err error) {
		if err != nil {
			client.OnError <- err
			return
		}

		switch msg.Header.MessageType {
		case HandshakeMessage:
			client.onHandshake <- true
		case PingMessage:
			if err := client.Send(PongMessage, nil); err != nil {
				return
			}
			if err := client.conn.SetReadDeadline(time.Now().Add(client.Option.HeartbeatTimeout)); err != nil {
				return
			}
		case BinaryMessage, TextMessage:
			client.OnMessage <- msg
		}
	})

	if client.Option.CryptoAlgo != CryptoAlgo_NoCrypto {
		if err := client.sendHandshake(ctx); err != nil {
			return nil, err
		}
	}

	return client, nil
}
