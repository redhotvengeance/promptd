package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/google/uuid"
)

type Client struct {
	conn      net.Conn
	mu        sync.Mutex
	pendingMu sync.Mutex
	pending   map[string]chan Response
}

func NewClient(conn net.Conn) *Client {
	return &Client{
		conn:    conn,
		pending: make(map[string]chan Response),
	}
}

func (c *Client) Send(payload any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	data = append(data, '\n')

	_, err = c.conn.Write(data)
	return err
}

func (c *Client) Call(ctx context.Context, method string, params any) (*Response, error) {
	id := uuid.NewString()

	req := Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  nil,
		ID:      id,
	}

	if params != nil {
		paramBytes, _ := json.Marshal(params)
		req.Params = paramBytes
	}

	respChan := make(chan Response, 1)

	c.pendingMu.Lock()
	c.pending[id] = respChan
	c.pendingMu.Unlock()

	if err := c.Send(req); err != nil {
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()

		return nil, err
	}

	select {
	case resp := <-respChan:
		if resp.Error != nil {
			return nil, fmt.Errorf("client error %d: %s", resp.Error.Code, resp.Error.Message)
		}

		return &resp, nil
	case <-ctx.Done():
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()

		return nil, ctx.Err()
	}
}

func (c *Client) resolve(resp Response) {
	c.pendingMu.Lock()
	ch, exists := c.pending[resp.ID]
	if exists {
		delete(c.pending, resp.ID)
	}
	c.pendingMu.Unlock()

	if exists {
		ch <- resp
	}
}
