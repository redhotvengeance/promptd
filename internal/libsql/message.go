package libsql

import (
	"context"

	"github.com/redhotvengeance/promptd/internal/libsql/data/sql"
	"github.com/redhotvengeance/promptd/internal/promptd"
)

type MessageService struct {
	queries *sql.Queries
}

func NewMessageService(queries *sql.Queries) *MessageService {
	return &MessageService{
		queries: queries,
	}
}

func (m *MessageService) modelToStruct(model sql.Message) promptd.Message {
	return promptd.Message{
		ID: model.ID,
		ThreadID: model.ThreadID,
		Role: promptd.Role(model.Role),
		Content: model.Content,
		CreatedAt: model.CreatedAt,
	}
}

func (m *MessageService) ListMessages(ctx context.Context, threadID string) ([]promptd.Message, error) {
	dbMessages, err := m.queries.ListMessages(ctx, threadID)
	if err != nil {
		return nil, err
	}

	messages := make([]promptd.Message, 0)
	for _, dm := range dbMessages {
		messages = append(messages, m.modelToStruct(dm))
	}

	return messages, nil
}

func (m *MessageService) CreateMessage(ctx context.Context, message promptd.Message) error {
	params := sql.CreateMessageParams{
		ID: message.ID,
		ThreadID: message.ThreadID,
		Role: string(message.Role),
		Content: message.Content,
	}

	if err := m.queries.CreateMessage(ctx, params); err != nil {
		return err
	}

	return nil
}

func (m *MessageService) DeleteMessage(ctx context.Context, id string) error {
	if err := m.queries.DeleteMessage(ctx, id); err != nil {
		return err
	}

	return nil
}
