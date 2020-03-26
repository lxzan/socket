package socket

import "net"

func Dial(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	client := newClient(conn, nil)
	go client.handleMessage()
	return client, nil
}
