package socket

import "net"

type Server struct {
	ReadBufSize int64
	OnError     func(err error)
	OnConnect   func(client *Client)
}

func NewServer() *Server {
	return new(Server)
}

func (this *Server) Run(addr string) error {
	if this.ReadBufSize == 0 {
		this.ReadBufSize = 2048
	}

	if this.OnConnect == nil {
		this.OnConnect = func(client *Client) {}
	}
	if this.OnError == nil {
		this.OnError = func(err error) {}
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			this.OnError(err)
			continue
		}

		client:=&Client{
			Conn:        conn,
			readBufSize: this.ReadBufSize,
		}
		client.handleMessage()
		this.OnConnect(client)
	}
}
