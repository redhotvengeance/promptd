package libsql

import (
	"context"

	"github.com/redhotvengeance/promptd/internal/libsql/data/sql"
	"github.com/redhotvengeance/promptd/internal/promptd"
)

type ThreadService struct {
	queries *sql.Queries
}

func NewThreadService(queries *sql.Queries) *ThreadService {
	return &ThreadService{
		queries: queries,
	}
}

func (t *ThreadService) modelToStruct(model sql.Thread) promptd.Thread {
	return promptd.Thread{
		ID: model.ID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func (t *ThreadService) GetThread(ctx context.Context, id string) (*promptd.Thread, error) {
	dbThread, err := t.queries.GetThread(ctx, id)
	if err != nil {
		return nil, err
	}

	thread := t.modelToStruct(dbThread)

	return &thread, nil
}

func (t *ThreadService) ListThreads(ctx context.Context) ([]promptd.Thread, error) {
	dbThreads, err := t.queries.ListThreads(ctx)
	if err != nil {
		return nil, err
	}

	threads := make([]promptd.Thread, 0, len(dbThreads))
	for _, dbThread := range dbThreads {
		threads = append(threads, t.modelToStruct(dbThread))
	}

	return threads, nil
}

func (t *ThreadService) CreateThread(ctx context.Context, id string) error {
	if err := t.queries.CreateThread(ctx, id); err != nil {
		return err
	}

	return nil
}

func (t *ThreadService) DeleteThread(ctx context.Context, id string) error {
	if err := t.queries.DeleteThread(ctx, id); err != nil {
		return err
	}

	return nil
}
