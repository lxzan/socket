package socket

import "net"

func Dial(addr string, opt *DialOption) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	client, err := newServerSideClient(conn, opt)
	if err != nil {
		return nil, err
	}

	go client.handleMessage()
	return client, nil
}
