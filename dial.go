package socket

import (
	"bytes"
	"context"
	"errors"
	"net"
	"time"
)

type Client struct {
	BaseClient
	onHandshake chan bool
}

func newClient(conn net.Conn, opt *Option) (*Client, error) {
	var c = &Client{
		BaseClient: BaseClient{
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

func (this *Client) handleMessage() {
	for {
		select {
		default:
			msg, err := this.read()
			if err != nil {
				this.OnError <- err
			}

			switch msg.Header.MessageType {
			case BinaryMessage, TextMessage:
				this.OnMessage <- msg
			case PingMessage:
				if _, err := this.Send(PongMessage, nil); err != nil {
					this.OnError <- err
				}
				if err := this.conn.SetReadDeadline(time.Now().Add(this.Option.HeartbeatTimeout)); err != nil {
					this.OnError <- err
				}
			}
		}
	}
}

func (this *Client) sendHandshake(ctx context.Context) error {
	var key = []byte(Alphabet.Generate(16))
	encryptKey, err := this.asymmetric.Encrypt(key)
	if err != nil {
		return err
	}

	if _, err := this.Send(HandshakeMessage, &Message{Body: encryptKey}); err != nil {
		return err
	}

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
			return errors.New("handshake timeout")
		}
	}
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

	go client.handleMessage()

	if client.Option.CryptoAlgo != CryptoAlgo_NoCrypto {
		if err := client.sendHandshake(ctx); err != nil {
			return nil, err
		}
	}

	return client, nil
}
