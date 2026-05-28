package generation

import (
	"context"
	"fmt"
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

type EditParams struct {
	Instruction      string              `json:"instruction"`
	StartLine        int                 `json:"startLine"`
	EndLine          int                 `json:"endLine"`
	Selection        string              `json:"selection"`
	Prefix           string              `json:"prefix"`
	Suffix           string              `json:"suffix"`
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

func (s *Service) HandleEdit(ctx context.Context, params EditParams) (<-chan string, error) {
	var builder strings.Builder

	builder.WriteString("You are a surgical code refactoring engine. Your goal is to rewrite the requested section of code based entirely on the user's instructions.\n")
	builder.WriteString("CRITICAL: Output ONLY the raw code replacement. Do not include markdown code fences (```), conversational commentary, or explanations. Start outputting code immediately.\n\n")
	builder.WriteString("<context_sandwich>\n")

	if params.Context != nil && params.Context.ActiveFilePath != "" {
		fmt.Fprintf(&builder, "File: %s\n\n", params.Context.ActiveFilePath)
	}

	fmt.Fprintf(&builder, "Prefix:\n%s\n\n", params.Prefix)
	fmt.Fprintf(&builder, "Target Selection (Lines %d-%d):\n%s\n\n", params.StartLine, params.EndLine, params.Selection)
	fmt.Fprintf(&builder, "Suffix:\n%s\n\n", params.Suffix)
	builder.WriteString("<context_sandwich>\n")

	history := []promptd.Message{{Role: promptd.RoleUser, Content: params.Instruction}}

	stream, err := s.llm.Chat(ctx, builder.String(), history, params.ProviderOverride)
	if err != nil {
		return nil, err
	}

	chunks := make(chan string)
	go func() {
		defer close(chunks)
		defer stream.Close()

		for {
			chunk, err := stream.Recv()
			if err != nil {
				break
			}

			chunks <- chunk
		}
	}()

	return chunks, nil
}

func generateID() string {
	return uuid.NewString()
}
