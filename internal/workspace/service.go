package workspace

import (
	"context"
	"fmt"
	"log"
	"strings"

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

func (s *Service) Register(ctx context.Context, workspacePath string) error {
	go func() {
		indexer := NewIndexer(s.llm, s.store)
		if err := indexer.IndexWorkspace(context.Background(), workspacePath); err != nil {
			log.Printf("Background indexing failed for %s: %v", workspacePath, err)
		}
	}()

	return nil
}

func (s *Service) Unregister(ctx context.Context, workspacePath string) error {
	if err := s.store.Workspaces().DeleteWorkspace(ctx, workspacePath); err != nil {
		return fmt.Errorf("failed to unregister workspace in db: %w", err)
	}

	return nil
}

func (s *Service) BuildContext(ctx context.Context, query string, jit *promptd.JITContext, workspacePath, embedModel string) (string, error) {
	queryVector, err := s.llm.Embed(ctx, query, embedModel)
	if err != nil {
		return "", fmt.Errorf("failed to embed query: %w", err)
	}

	chunks, err := s.store.Workspaces().SearchChunks(ctx, workspacePath, queryVector, 10)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve chunks: %w", err)
	}

	var builder strings.Builder
	builder.WriteString("You are an expert AI coding assistant. Use the provided workspace context to accurately answer the user's request.\n\n")

	if len(chunks) > 0 {
		builder.WriteString("<workspace_context>\n")
		builder.WriteString("Here are relevant code snippets retrieved from the user's project:\n\n")

		for _, chunk := range chunks {
			fmt.Fprintf(&builder, "<file path=\"%s\">\n%s\n</file>\n\n", chunk.FilePath, chunk.Content)
		}

		builder.WriteString("</workspace_context>\n\n")
	}

	if jit != nil {
		builder.WriteString("<editor_state>\n")
		builder.WriteString("This is the exact current state of the user's editor:\n\n")

		if len(jit.OpenBuffers) > 0 {
			fmt.Fprintf(&builder, "**Currently Open Buffers:** %s\n\n", strings.Join(jit.OpenBuffers, ", "))
		}

		if jit.ActiveFilePath != "" {
			fmt.Fprintf(&builder, "**Active File:** `%s`\n", jit.ActiveFilePath)
			fmt.Fprintf(&builder, "**Cursor Position:** `%d`\n", jit.CursorLine)
			fmt.Fprintf(&builder, "<active_file_context>\n%s\n</active_file_context>\n", jit.ActiveFileContent)
		}

		builder.WriteString("</editor_state>\n\n")
	}

	return builder.String(), nil
}
