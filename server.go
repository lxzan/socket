package socket

import (
	"net"
	"time"
)

type Server struct {
	*Option
}

type Option struct {
	CompressAlgo                   // default gzip
	CryptoAlgo                     // default RSA-AES
	PublicKey        string        // pem file path
	PrivateKey       string        // pem file path
	MinCompressSize  int           // compress data when dataLength>=CompressMinsize
	HeartbeatTimeout time.Duration // io timeout
	PingInterval     time.Duration // ping interval
}

func (this *Option) initialize() {
	if this.CompressAlgo == 0 {
		this.CompressAlgo = CompressAlgo_Gzip
	}
	if this.CryptoAlgo == 0 {
		this.CryptoAlgo = CryptoAlgo_NoCrypto
	}
	if this.MinCompressSize == 0 {
		this.MinCompressSize = 4 * 1024
	}
	if this.HeartbeatTimeout == 0 {
		this.HeartbeatTimeout = time.Minute
	}
	if this.PingInterval == 0 {
		this.PingInterval = 5 * time.Second
	}
}

func NewServer(opt *Option) *Server {
	if opt == nil {
		opt = &Option{}
	}
	opt.initialize()
	return &Server{Option: opt}
}

func (this *Server) Run(addr string, onconnect func(client *Conn)) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		client, err := newConn(conn, this.Option)
		if err != nil {
			return err
		}

		go func() {
			defer client.PingTicker.Stop()
			onconnect(client)
		}()

		go client.read(func(msg *Message, err error) {
			if err != nil {
				client.OnError <- err
				return
			}

			switch msg.Header.MessageType {
			case HandshakeMessage:
				if err := client.handleHandshake(msg); err != nil {
					client.OnError <- err
					return
				}
			case BinaryMessage, TextMessage:
				client.OnMessage <- msg
				return
			case PongMessage:
				println(time.Now().Unix())
				if err := client.conn.SetReadDeadline(time.Now().Add(this.Option.HeartbeatTimeout)); err != nil {
					client.OnError <- err
					return
				}
			}
		})
	}
}
