package socket

import "net"

type Server struct {
	OnConnect     func(client *Client)
	DefaultClient *Client
}

func NewServer(opt *DialOption) *Server {
	s := new(Server)
	client, err := newServerSideClient(nil, opt)
	if err != nil {
		panic(err)
	}
	s.DefaultClient = client
	return s
}

func (this *Server) Run(addr string) error {

	if this.OnConnect == nil {
		this.OnConnect = func(client *Client) {}
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		client, err := newClientSideClient(conn, this.DefaultClient.Option)
		if err != nil {
			return err
		}

		go this.OnConnect(client)
		client.handleMessage()
	}
}
