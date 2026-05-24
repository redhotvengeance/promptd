package router

import (
	"context"
	"log"

	"github.com/redhotvengeance/promptd/internal/ipc/server"
)

type Router struct {}

func NewRouter() *Router {
	return &Router{}
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
}

func (r *Router) handleEdit(ctx context.Context, req ipc.Request, client *ipc.Client) {
}

func (r *Router) handleFIM(ctx context.Context, req ipc.Request, client *ipc.Client) {
}

func (r *Router) handleTask(ctx context.Context, req ipc.Request, client *ipc.Client) {
}
