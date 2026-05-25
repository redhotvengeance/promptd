package generation

import (
	"context"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/redhotvengeance/promptd/internal/promptd"
)

type Service struct {
	llm   promptd.LLM
	store promptd.Datastore
}

func NewService(llm promptd.LLM, store promptd.Datastore) *Service {
	return &Service{
		llm:   llm,
		store: store,
	}
}

type ChatParams struct {
	Prompt           string              `json:"prompt"`
	ThreadID         string              `json:"threadID,omitempty"`
	Context          *promptd.JITContext `json:"context,omitempty"`
	ProviderOverride string              `json:"providerOverride,omitempty"`
}

func (s *Service) HandleChat(ctx context.Context, params ChatParams) (<-chan string, error) {
	messages := s.store.Messages()
	threads := s.store.Threads()

	if _, err := threads.GetThread(ctx, params.ThreadID); err != nil {
		if err := threads.CreateThread(ctx, params.ThreadID); err != nil {
			log.Printf("Error create thread %s: %v", params.ThreadID, err)
		}
	}

	history, err := messages.ListMessages(ctx, params.ThreadID)
	if err != nil {
		log.Printf("Failed to get thread history: %v", err)
	}

	userMsg := promptd.Message{
		ID:       generateID(),
		ThreadID: params.ThreadID,
		Role:     promptd.RoleUser,
		Content:  params.Prompt,
	}

	if err := messages.CreateMessage(ctx, userMsg); err != nil {
		log.Printf("CRITICAL: Failed to save user message to DB: %v", err)
	}

	history = append(history, userMsg)

	systemPrompt := "You are a helpful coding assistant."

	stream, err := s.llm.Chat(ctx, systemPrompt, history, params.ProviderOverride)
	if err != nil {
		return nil, err
	}

	chunks := make(chan string)

	go func() {
		defer close(chunks)
		defer stream.Close()

		var builder strings.Builder

		for {
			chunk, err := stream.Recv()
			if err != nil {
				break
			}

			builder.WriteString(chunk)

			chunks <- chunk
		}

		if err = messages.CreateMessage(ctx, promptd.Message{
			ID:       generateID(),
			ThreadID: params.ThreadID,
			Role:     promptd.RoleAssistant,
			Content:  builder.String(),
		}); err != nil {
			log.Printf("CRITICAL: Failed to save assistant message to DB: %v", err)
		}
	}()

	return chunks, nil
}

func generateID() string {
	return uuid.NewString()
}
