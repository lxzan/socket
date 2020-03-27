package socket

import "net"

type Server struct {
	defaultClient *Client
}

func NewServer(opt *DialOption) *Server {
	s := new(Server)
	client, err := newServerSideClient(nil, opt)
	if err != nil {
		panic(err)
	}
	s.defaultClient = client
	return s
}

func (this *Server) Run(addr string, onconnect func(client *Client)) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		client, err := newServerSideClient(conn, this.defaultClient.Option)
		if err != nil {
			return err
		}

		go onconnect(client)

		go client.handleMessage()
	}
}
