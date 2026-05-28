package promptd

import (
	"context"
	"time"
)

type Thread struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ThreadStore interface {
	GetThread(ctx context.Context, id string) (*Thread, error)
	ListThreads(ctx context.Context) ([]Thread, error)
	CreateThread(ctx context.Context, id string) error
	DeleteThread(ctx context.Context, id string) error
}
