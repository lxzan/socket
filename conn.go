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

type BaseClient struct {
	conn       net.Conn
	asymmetric Crypto
	aes        Crypto
	readBuffer *bytes.Buffer
	Option     *Option
	OnMessage  chan *Message
	OnError    chan error
}

func (this *BaseClient) read(callback func(msg *Message, err error)) {
	var rl uint32 = 4
	var rlb = true

	for {
		n, err := io.CopyN(this.readBuffer, this.conn, int64(rl))
		if err != nil {
			callback(nil, ERR_ReadMessage.Wrap(err.Error()))
			return
		}
		if n != int64(rl) {
			callback(nil, ERR_ReadMessage.Wrap("network error"))
			return
		}

		packet := make([]byte, rl)
		_, err = this.readBuffer.Read(packet)
		if err != nil {
			callback(nil, ERR_ReadMessage.Wrap(err.Error()))
			return
		}

		if rlb {
			rl = binary.LittleEndian.Uint32(packet)
			rlb = false
		} else {
			rl = 4
			rlb = true
			msg, err := this.splitPacket(packet)
			if err != nil {
				callback(nil, ERR_DecodeMessage.Wrap(err.Error()))
				return
			}
			callback(msg, nil)
		}
	}
}

func (this *BaseClient) splitPacket(packet []byte) (msg *Message, err error) {
	msg = &Message{Header: &Header{}}
	var totalLength = len(packet)
	if totalLength < 6 {
		return nil, errors.New("illegal data")
	}
	if err := msg.Header.decodeProtocolHeader(packet); err != nil {
		return nil, err
	}

	msg.Body = packet[6:]
	if int(msg.Header.HeaderLength) > len(msg.Body) {
		return nil, errors.New("illegal data")
	}

	if msg.Header.CryptoAlgorithm != CryptoAlgo_NoCrypto {
		body, err := this.aes.Decrypt(msg.Body)
		if err != nil {
			return nil, err
		} else {
			msg.Body = body
		}
	}

	if body, err := uncompress(msg.Header.CompressAlgorithm, msg.Body); err != nil {
		return nil, err
	} else {
		msg.Body = body
	}

	if msg.Header.HeaderLength > 0 {
		if err := jsoniter.Unmarshal(msg.Body[:msg.Header.HeaderLength], &msg.Header.Form); err != nil {
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
func (this *BaseClient) Send(typ MessageType, msg *Message) (n int, err error) {
	if msg == nil {
		msg = &Message{}
	}
	if msg.Header == nil {
		msg.Header = &Header{Form: Form{}}
	}

	var p0 = make([]byte, 4)
	var p1 = byte(currentProtocol)
	var p2 = byte(typ)
	var p3 = byte(this.Option.CompressAlgo)
	var p4 = byte(this.Option.CryptoAlgo)
	var p5 = make([]byte, 2)
	var p6 []byte
	if len(msg.Header.Form) > 0 {
		p6, _ = jsoniter.Marshal(msg.Header.Form)
	}

	var headerLength = len(p6)
	p6 = append(p6, msg.Body...)

	if typ != TextMessage && typ != BinaryMessage {
		p3 = byte(CompressAlgo_NoCompress)
		p4 = byte(CryptoAlgo_NoCrypto)
	}
	if this.Option.CompressAlgo != CompressAlgo_NoCompress {
		if len(p6) >= this.Option.MinCompressSize {
			res, err := compress(this.Option.CompressAlgo, p6)
			if err != nil {
				return 0, err
			} else {
				p6 = res
			}
		} else {
			p3 = byte(CompressAlgo_NoCompress)
		}
	}

	if this.Option.CryptoAlgo != CryptoAlgo_NoCrypto {
		res, err := this.aes.Encrypt(p6)
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

func (this *BaseClient) SendContext(ctx context.Context, typ MessageType, msg *Message) (n int, err error) {
	var sig = make(chan bool)
	defer close(sig)

	go func() {
		n, err = this.Send(typ, msg)
		sig <- true
	}()

	for {
		select {
		case <-sig:
			return
		case <-ctx.Done():
			return 0, ERR_Timeout.Wrap("send message timeout")
		}
	}
}

type Conn struct {
	BaseClient
	PingTicker *time.Ticker
}

func newConn(conn net.Conn, opt *Option) (*Conn, error) {
	var c = &Conn{
		BaseClient: BaseClient{
			conn:       conn,
			readBuffer: bytes.NewBufferString(""),
			OnMessage:  make(chan *Message, 16),
			OnError:    make(chan error, 16),
			Option:     opt,
		},
		PingTicker: time.NewTicker(opt.PingInterval),
	}

	if opt.CryptoAlgo != CryptoAlgo_NoCrypto {
		if opt.PrivateKey == "" {
			return nil, errors.New("private key not set")
		}
		if opt.CryptoAlgo == CryptoAlgo_RsaAes {
			rsa, err := NewRsaCrypto("", opt.PrivateKey)
			if err != nil {
				return nil, err
			} else {
				c.asymmetric = rsa
			}
		}
	}
	return c, nil
}

func (this *Conn) handleHandshake(msg *Message) error {
	key, err := this.asymmetric.Decrypt(msg.Body)
	if err != nil {
		return err
	}

	encoder, err := NewAesCrypto(key)
	if err != nil {
		return err
	}
	this.aes = encoder

	_, err = this.Send(HandshakeMessage, nil)
	return err
}

func (this *Conn) Ping() {
	if _, err := this.Send(PingMessage, nil); err != nil {
		this.OnError <- err
	}
}
