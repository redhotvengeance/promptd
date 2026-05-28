package ipc

import "encoding/json"

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      string          `json:"id"`
}

type Response struct {
	JSONRPC string `json:"jsonrpc"`
	Result  any    `json:"result"`
	Error   *Error `json:"error,omitempty"`
	ID      string `json:"id"`
}

type Notification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}
