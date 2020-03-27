package socket

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"github.com/json-iterator/go"
	"io"
	"net"
	"time"
)

type Client struct {
	conn        net.Conn
	asymmetric  Encoder
	aes         Encoder
	compression Encoder
	Option      *DialOption
	OnMessage   chan *Message
	OnError     chan error
	onHandshake chan *Message
}

func initClient(conn net.Conn, opt *DialOption) *Client {
	if opt == nil {
		opt = &DialOption{}
	}
	if opt.CompressAlgo == 0 {
		opt.CompressAlgo = CompressAlgo_Gzip
	}
	if opt.CryptoAlgo == 0 {
		opt.CryptoAlgo = CryptoAlgo_NoCrypto
	}
	if opt.CompressMinsize == 0 {
		opt.CompressMinsize = 4 * 1024
	}
	if opt.HandshakeTimeout == time.Duration(0) {
		opt.HandshakeTimeout = 5 * time.Second
	}

	var client = &Client{
		conn:        conn,
		OnMessage:   make(chan *Message, 16),
		OnError:     make(chan error, 16),
		onHandshake: make(chan *Message),
		Option:      opt,
	}
	if opt.CompressAlgo != CompressAlgo_NoCompress {
		if opt.CompressAlgo == CompressAlgo_Gzip {
			client.compression = GzipEncoder
		}
	}
	return client
}

func newServerSideClient(conn net.Conn, opt *DialOption) (*Client, error) {
	var client = initClient(conn, opt)
	client.Option.serverSide = true

	if opt.CryptoAlgo != CryptoAlgo_NoCrypto {
		if opt.PrivateKey == "" {
			return nil, errors.New("private key not set")
		}
		if opt.CryptoAlgo == CryptoAlgo_RsaAes {
			rsa, err := NewRSA(opt.PublicKey, opt.PrivateKey)
			if err != nil {
				return nil, err
			} else {
				client.asymmetric = rsa
			}
		}
	}

	return client, nil
}

func newClientSideClient(conn net.Conn, opt *DialOption) (*Client, error) {
	var client = initClient(conn, opt)
	client.Option.serverSide = false

	if opt.CryptoAlgo != CryptoAlgo_NoCrypto {
		if opt.PublicKey == "" {
			return nil, errors.New("public key not set")
		}
		if opt.CryptoAlgo == CryptoAlgo_RsaAes {
			rsa, err := NewRSA(opt.PublicKey, opt.PrivateKey)
			if err != nil {
				return nil, err
			} else {
				client.asymmetric = rsa
			}
		}
	}

	return client, nil
}

func (this *Client) handleMessage() {
	var buf = bytes.NewBufferString("")
	var rl uint32 = 4
	var rlb = true
	for {
		n, err := io.CopyN(buf, this.conn, int64(rl))
		if err != nil {
			this.OnError <- ERR_ReadMessage.Wrap(err.Error())
			return
		}
		if n != int64(rl) {
			this.OnError <- ERR_ReadMessage.Wrap("data length error")
			return
		}

		frame := make([]byte, rl)
		_, err = buf.Read(frame)
		if err != nil {
			this.OnError <- ERR_ReadMessage.Wrap(err.Error())
			return
		}

		if rlb {
			rl = binary.LittleEndian.Uint32(frame)
			rlb = false
		} else {
			msg, err := this.decodeMessage(frame)
			if err != nil {
				this.OnError <- ERR_DecodeMessage.Wrap(err.Error())
				return
			}
			rl = 4
			rlb = true

			if msg.Header.MessageType == HandshakeMessage {
				if this.Option.serverSide {
					this.handleHandshake(msg)
				} else {
					this.onHandshake <- msg
				}
			} else {
				this.OnMessage <- msg
			}
		}
	}
}

func (this *Client) decodeMessage(data []byte) (msg *Message, err error) {
	msg = &Message{Header: Header{}}
	var totalLength = len(data)
	if totalLength < 6 {
		return nil, errors.New("received invalid data")
	}
	if err := msg.Header.decodeProtocolHeader(data); err != nil {
		return nil, err
	}

	msg.Body = data[6:]
	if msg.Header.CryptoAlgorithm != CryptoAlgo_NoCrypto {
		body, err := this.aes.Decode(msg.Body)
		if err != nil {
			return nil, err
		} else {
			msg.Body = body
		}
	}

	if msg.Header.CompressAlgorithm != CompressAlgo_NoCompress {
		body, err := this.compression.Decode(msg.Body)
		if err != nil {
			return nil, err
		} else {
			msg.Body = body
		}
	}

	if msg.Header.HeaderLength > 0 {
		if err := jsoniter.Unmarshal(msg.Body[:msg.Header.HeaderLength], &msg.Header.form); err != nil {
			return nil, err
		}
	}

	msg.Body = msg.Body[msg.Header.HeaderLength:]
	return msg, nil
}

// p0: Content Length 4B
// p1: Protocol Version 1B
// p2: Message Type 1B
// p3: Compression Algorithm 1B
// p4: Crypto Algorithm 1B
// p5: Header Length 2B
// p6: UserHeader and Body
func (this *Client) WriteMessage(typ MessageType, header Form, data []byte) (n int, err error) {
	if header == nil {
		header = Form{}
	}

	var p0 = make([]byte, 4)
	var p1 = byte(currentProtocol)
	var p2 = byte(typ)
	var p3 = byte(this.Option.CompressAlgo)
	var p4 = byte(this.Option.CryptoAlgo)
	var p5 = make([]byte, 2)
	var p6 []byte
	if len(header) > 0 {
		p6, _ = jsoniter.Marshal(header)
	}

	var headerLength = len(p6)
	p6 = append(p6, data...)
	if this.Option.CompressAlgo != CompressAlgo_NoCompress {
		if len(p6) >= this.Option.CompressMinsize {
			res, err := this.compression.Encode(p6)
			if err != nil {
				return 0, err
			} else {
				p6 = res
			}
		} else {
			p3 = byte(CompressAlgo_NoCompress)
		}
	}

	if typ == HandshakeMessage {
		p4 = byte(CryptoAlgo_NoCrypto)
	} else if this.Option.CryptoAlgo != CryptoAlgo_NoCrypto {
		res, err := this.aes.Encode(p6)
		if err != nil {
			return 0, err
		} else {
			p6 = res
		}
	}

	binary.LittleEndian.PutUint16(p5, uint16(headerLength))
	binary.LittleEndian.PutUint32(p0, uint32(6+len(p6)))

	var buf = bytes.NewBuffer(p0)
	buf.Write([]byte{p1, p2, p3, p4})
	buf.Write(p5)
	buf.Write(p6)
	return this.conn.Write(buf.Bytes())
}

// server side
func (this *Client) handleHandshake(msg *Message) error {
	key, err := this.asymmetric.Decode(msg.Body)
	if err != nil {
		return err
	}

	encoder, err := NewAES(key)
	if err != nil {
		return err
	}
	this.aes = encoder

	_, err = this.WriteMessage(HandshakeMessage, nil, nil)
	return err
}

// client side
func (this *Client) sendHandshake() error {
	var key = []byte(Alphabet.Generate(16))
	encryptKey, err := this.asymmetric.Encode(key)
	if err != nil {
		return err
	}

	if _, err := this.WriteMessage(HandshakeMessage, nil, encryptKey); err != nil {
		return err
	}

	ctx, _ := context.WithTimeout(context.Background(), this.Option.HandshakeTimeout)
	for {
		select {
		case <-this.onHandshake:
			encoder, err := NewAES(key)
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
