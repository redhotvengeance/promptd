package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redhotvengeance/promptd/internal/generation"
	"github.com/redhotvengeance/promptd/internal/ipc/server"
)

type Router struct {
	genService *generation.Service
}

func NewRouter(genService *generation.Service) *Router {
	return &Router{
		genService: genService,
	}
}

func (r *Router) Handle(req ipc.Request, client *ipc.Client) {
	ctx := context.Background()

	switch req.Method {
	case "text.chat":
		r.handleChat(ctx, req, client)
	case "text.edit":
		r.handleEdit(ctx, req, client)
	case "text.fim":
		r.handleFIM(ctx, req, client)
	case "text.task":
		r.handleTask(ctx, req, client)
	default:
		log.Printf("Unknown method: %s", req.Method)
	}
}

func (r *Router) handleChat(ctx context.Context, req ipc.Request, client *ipc.Client) {
	var params generation.ChatParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		r.sendError(client, req.ID, -32602, "Invalid params format")

		return
	}

	streamChan, err := r.genService.HandleChat(ctx, params)
	if err != nil {
		r.sendError(client, req.ID, -32000, err.Error())

		return
	}

	if err = client.Send(ipc.Notification{
		JSONRPC: "2.0",
		Method:  "stream/begin",
		Params: map[string]any{
			"id": req.ID,
		},
	}); err != nil {
		log.Printf("Failed to begin stream to client: %v", err)

		return
	}

	for chunk := range streamChan {
		if err := client.Send(ipc.Notification{
			JSONRPC: "2.0",
			Method:  "stream/chunk",
			Params: map[string]any{
				"id":   req.ID,
				"text": chunk,
			},
		}); err != nil {
			log.Printf("Failed to stream chunk to client: %v", err)

			return
		}
	}

	if err = client.Send(ipc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  nil,
	}); err != nil {
		log.Printf("Failed to finish stream to client: %v", err)

		return
	}
}

func (r *Router) handleEdit(ctx context.Context, req ipc.Request, client *ipc.Client) {
	var params generation.EditParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		r.sendError(client, req.ID, -32602, "Invalid params format")

		return
	}

	streamChan, err := r.genService.HandleEdit(ctx, params)
	if err != nil {
		r.sendError(client, req.ID, -32000, err.Error())

		return
	}

	if err = client.Send(ipc.Notification{
		JSONRPC: "2.0",
		Method:  "stream/begin",
		Params: map[string]any{
			"id": req.ID,
		},
	}); err != nil {
		log.Printf("Failed to begin stream to client: %v", err)

		return
	}

	for chunk := range streamChan {
		if err := client.Send(ipc.Notification{
			JSONRPC: "2.0",
			Method:  "stream/chunk",
			Params: map[string]any{
				"id":   req.ID,
				"text": chunk,
			},
		}); err != nil {
			log.Printf("Failed to stream chunk to client: %v", err)

			return
		}
	}

	if err = client.Send(ipc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  nil,
	}); err != nil {
		log.Printf("Failed to finish stream to client: %v", err)

		return
	}
}

func (r *Router) handleFIM(ctx context.Context, req ipc.Request, client *ipc.Client) {
	var params generation.FIMParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		r.sendError(client, req.ID, -32602, "Invalid params format")

		return
	}

	completion, err := r.genService.HandleFIM(ctx, params)
	if err != nil {
		r.sendError(client, req.ID, -32000, err.Error())

		return
	}

	_ = client.Send(ipc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]string{
			"completion": completion,
		},
	})
}

func (r *Router) handleTask(ctx context.Context, req ipc.Request, client *ipc.Client) {
	var params generation.TaskParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		r.sendError(client, req.ID, -32602, "Invalid params format")

		return
	}

	params.ExecuteTool = func(ctx context.Context, name, args string) (string, error) {
		editorResp, err := client.Call(ctx, "client/toolCall", map[string]any{
			"toolName":  name,
			"arguments": args,
		})
		if err != nil {
			return "", nil
		}

		resultMap, ok := editorResp.Result.(map[string]any)
		if !ok {
			return "", fmt.Errorf("invalid tool response format from editor")
		}

		output, _ := resultMap["output"].(string)

		return output, nil
	}

	streamChan, err := r.genService.HandleTask(ctx, params)
	if err != nil {
		r.sendError(client, req.ID, -32000, err.Error())

		return
	}

	_ = client.Send(ipc.Notification{
		JSONRPC: "2.0",
		Method:  "stream/begin",
		Params: map[string]any{
			"id": req.ID,
		},
	})

	for update := range streamChan {
		payload := map[string]any{
			"id": req.ID,
		}

		if update.Status != "" {
			payload["status"] = update.Status
		}

		if update.Text != "" {
			payload["text"] = update.Text
		}

		_ = client.Send(ipc.Notification{
			JSONRPC: "2.0",
			Method:  "stream/chunk",
			Params:  payload,
		})
	}

	_ = client.Send(ipc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  nil,
	})
}

func (r *Router) sendError(client *ipc.Client, id string, code int, msg string) {
	_ = client.Send(ipc.Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &ipc.Error{
			Code:    code,
			Message: msg,
		},
	})
}
