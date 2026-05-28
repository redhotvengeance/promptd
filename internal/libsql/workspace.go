package libsql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	queries "github.com/redhotvengeance/promptd/internal/libsql/data/sql"
	"github.com/redhotvengeance/promptd/internal/promptd"
)

func vectorToString(vec []float32) (string, error) {
	bytes, err := json.Marshal(vec)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

type WorkspaceStore struct {
	db      *sql.DB
	queries *queries.Queries
}

func NewWorkspaceStore(db *sql.DB, queries *queries.Queries) *WorkspaceStore {
	return &WorkspaceStore{
		db:      db,
		queries: queries,
	}
}

func (w *WorkspaceStore) DeleteWorkspace(ctx context.Context, workspacePath string) error {
	return w.queries.DeleteWorkspace(ctx, workspacePath)
}

func (w *WorkspaceStore) InsertChunks(ctx context.Context, chunks []promptd.Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := w.queries.WithTx(tx)

	for _, chunk := range chunks {
		vecStr, err := vectorToString(chunk.Embedding)
		if err != nil {
			log.Fatalf("Failed to stringify vector: %v", err)
		}

		if err = qtx.InsertChunk(ctx, queries.InsertChunkParams{
			ID:            chunk.ID,
			WorkspacePath: chunk.WorkspacePath,
			Filepath:      chunk.FilePath,
			Content:       chunk.Content,
			Vector32:      vecStr,
		}); err != nil {
			return fmt.Errorf("failed to insert chunk %s: %w", chunk.ID, err)
		}
	}

	return tx.Commit()
}

func (w *WorkspaceStore) SearchChunks(ctx context.Context, workspacePath string, queryVector []float32, limit int) ([]promptd.Chunk, error) {
	vecStr, err := vectorToString(queryVector)
	if err != nil {
		log.Fatalf("Failed to stringify query vector: %v", err)
	}

	rows, err := w.queries.SearchChunks(ctx, queries.SearchChunksParams{
		WorkspacePath: workspacePath,
		Vector32:      vecStr,
		Limit:         int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	var chunks []promptd.Chunk
	for _, row := range rows {
		chunks = append(chunks, promptd.Chunk{
			ID:            row.ID,
			WorkspacePath: workspacePath,
			FilePath:      row.Filepath,
			Content:       row.Content,
		})
	}

	return chunks, nil
}
