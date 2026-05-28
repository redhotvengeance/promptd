package ipc

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type Handler func(req Request, client *Client)

type Server struct {
	socketPath string
	handler    Handler
}

func NewServer(path string, handler Handler) *Server {
	return &Server{
		socketPath: path,
		handler:    handler,
	}
}

func (s *Server) Start() error {
	os.Remove(s.socketPath)

	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return err
	}
	defer listener.Close()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stopChan

		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Println("Socket close requested. Cleanly exiting.")

				break
			}

			log.Printf("Failed to accept connection: %v", err)

			continue
		}

		go s.handleConnection(conn)
	}

	os.Remove(s.socketPath)

	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	client := NewClient(conn)

	for scanner.Scan() {
		line := scanner.Bytes()

		var raw map[string]any
		if err := json.Unmarshal(line, &raw); err != nil {
			client.Send(Response{
				JSONRPC: "2.0",
				Error: &Error{
					Code:    -32700,
					Message: "Parse error",
				},
			})

			continue
		}

		if raw["method"] == nil && raw["id"] != nil {
			var resp Response
			if err := json.Unmarshal(line, &resp); err == nil {
				client.resolve(resp)
			}

			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			client.Send(Response{
				JSONRPC: "2.0",
				Error: &Error{
					Code:    -32700,
					Message: "Parse error",
				},
			})

			continue
		}

		if req.JSONRPC != "2.0" {
			client.Send(Response{
				JSONRPC: "2.0",
				Error: &Error{
					Code:    -32600,
					Message: "Invalid request: Expected `jsonrpc` version 2.0",
				},
			})

			continue
		}

		if req.Method == "" {
			client.Send(Response{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &Error{
					Code:    -32600,
					Message: "Invalid request: Missing `method`",
				},
			})

			continue
		}

		go s.handler(req, client)
	}
}
