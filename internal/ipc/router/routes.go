package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redhotvengeance/promptd/internal/generation"
	"github.com/redhotvengeance/promptd/internal/ipc/server"
	"github.com/redhotvengeance/promptd/internal/promptd"
)

type Router struct {
	genService *generation.Service
	datastore  promptd.Datastore
	workspace  promptd.WorkspaceService
}

type WorkspaceParams struct {
	WorkspaceRoot string `json:"workspaceRoot"`
}

type ThreadParams struct {
	ThreadID string `json:"threadID"`
}

type ThreadListParams struct {
	Limit string `json:"limit"`
	Offet string `json:"offset"`
}

type SystemCancelParams struct {
	RequestID string `json:"requestID"`
}

func NewRouter(genService *generation.Service, datastore promptd.Datastore, workspaceService promptd.WorkspaceService) *Router {
	return &Router{
		genService: genService,
		datastore:  datastore,
		workspace:  workspaceService,
	}
}

func (r *Router) Handle(req ipc.Request, client *ipc.Client) {
	ctx := context.Background()

	switch req.Method {
	case "text/chat":
		r.handleChat(ctx, req, client)
	case "text/edit":
		r.handleEdit(ctx, req, client)
	case "text/fim":
		r.handleFIM(ctx, req, client)
	case "text/task":
		r.handleTask(ctx, req, client)
	case "workspace/register":
		r.handleWorkspaceRegister(ctx, req, client)
	case "workspace/unregister":
		r.handleWorkspaceUnregister(ctx, req, client)
	case "thread/get":
		r.handleThreadGet(ctx, req, client)
	case "thread/list":
		r.handleThreadList(ctx, req, client)
	case "thread/delete":
		r.handleThreadDelete(ctx, req, client)
	case "system/status":
		r.handleSystemStatus(req, client)
	case "system/cancel":
		r.handleSystemCancel(req, client)
	case "provider/list":
		r.handleProviderList(req, client)
	default:
		log.Printf("Unknown method: %s", req.Method)
		r.sendError(client, req.ID, -32601, "Method not found")
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

func (r *Router) handleWorkspaceRegister(ctx context.Context, req ipc.Request, client *ipc.Client) {
	var params WorkspaceParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		r.sendError(client, req.ID, -32602, "Invalid params format")

		return
	}

	if err := r.workspace.Register(ctx, params.WorkspaceRoot); err != nil {
		r.sendError(client, req.ID, -32000, err.Error())

		return
	}

	_ = client.Send(ipc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"status": "indexing",
		},
	})
}

func (r *Router) handleWorkspaceUnregister(ctx context.Context, req ipc.Request, client *ipc.Client) {
	var params WorkspaceParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		r.sendError(client, req.ID, -32602, "Invalid params format")

		return
	}

	if err := r.workspace.Unregister(ctx, params.WorkspaceRoot); err != nil {
		r.sendError(client, req.ID, -32000, err.Error())

		return
	}

	_ = client.Send(ipc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"success": true,
		},
	})
}

func (r *Router) handleThreadGet(ctx context.Context, req ipc.Request, client *ipc.Client) {
	var params ThreadParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		r.sendError(client, req.ID, -32602, "Invalid params format")

		return
	}

	messages, err := r.datastore.Messages().ListMessages(ctx, params.ThreadID)
	if err != nil {
		r.sendError(client, req.ID, -32000, err.Error())

		return
	}

	_ = client.Send(ipc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"messages": messages,
		},
	})
}

func (r *Router) handleThreadList(ctx context.Context, req ipc.Request, client *ipc.Client) {
	threads, err := r.datastore.Threads().ListThreads(ctx)
	if err != nil {
		r.sendError(client, req.ID, -32000, err.Error())
	}

	_ = client.Send(ipc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"threads": threads,
		},
	})
}

func (r *Router) handleThreadDelete(ctx context.Context, req ipc.Request, client *ipc.Client) {
	var params ThreadParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		r.sendError(client, req.ID, -32602, "Invalid params format")

		return
	}

	if err := r.datastore.Threads().DeleteThread(ctx, params.ThreadID); err != nil {
		r.sendError(client, req.ID, -32000, err.Error())

		return
	}

	_ = client.Send(ipc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"success": true,
		},
	})
}

func (r *Router) handleSystemStatus(req ipc.Request, client *ipc.Client) {
	_ = client.Send(ipc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"status":          "online",
			"activeProviders": []string{}, // TODO: Fetch from config
		},
	})
}

func (r *Router) handleSystemCancel(req ipc.Request, client *ipc.Client) {
	// TODO: Implement context cancellation

	_ = client.Send(ipc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"success": true,
		},
	})
}

func (r *Router) handleProviderList(req ipc.Request, client *ipc.Client) {
	_ = client.Send(ipc.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"providers": []string{}, // TODO: Fetch from config
		},
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
