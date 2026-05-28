package libsql

import (
	"context"

	"github.com/redhotvengeance/promptd/internal/libsql/data/sql"
	"github.com/redhotvengeance/promptd/internal/promptd"
)

type ThreadStore struct {
	queries *sql.Queries
}

func NewThreadStore(queries *sql.Queries) *ThreadStore {
	return &ThreadStore{
		queries: queries,
	}
}

func (t *ThreadStore) modelToStruct(model sql.Thread) promptd.Thread {
	return promptd.Thread{
		ID: model.ID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func (t *ThreadStore) GetThread(ctx context.Context, id string) (*promptd.Thread, error) {
	dbThread, err := t.queries.GetThread(ctx, id)
	if err != nil {
		return nil, err
	}

	thread := t.modelToStruct(dbThread)

	return &thread, nil
}

func (t *ThreadStore) ListThreads(ctx context.Context) ([]promptd.Thread, error) {
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

func (t *ThreadStore) CreateThread(ctx context.Context, id string) error {
	if err := t.queries.CreateThread(ctx, id); err != nil {
		return err
	}

	return nil
}

func (t *ThreadStore) DeleteThread(ctx context.Context, id string) error {
	if err := t.queries.DeleteThread(ctx, id); err != nil {
		return err
	}

	return nil
}
