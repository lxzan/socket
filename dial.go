package socket

import "net"

func Dial(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn:        conn,
		readBufSize: 2048,
	}
	return client, nil
}
