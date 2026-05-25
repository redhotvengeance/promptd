package promptd

import (
	"context"
	"time"
)

type Message struct {
	ID        string
	ThreadID  string
	Role      Role
	Content   string
	CreatedAt time.Time
}

type MessageService interface {
	ListMessages(ctx context.Context, threadID string) ([]Message, error)
	CreateMessage(ctx context.Context, message Message) error
	DeleteMessage(ctx context.Context, id string) error
}
