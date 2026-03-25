package ipc

import (
	"encoding/json"
	"net"
	"sync"
)

type Client struct {
	conn net.Conn
	mu sync.Mutex
}

func NewClient(conn net.Conn) *Client {
	return &Client{conn: conn}
}

func (c *Client) Send(resp Response) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	data = append(data, '\n')

	_, err = c.conn.Write(data)
	return err
}
