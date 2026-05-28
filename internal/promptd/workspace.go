package promptd

import "context"

type Chunk struct {
	ID            string
	WorkspacePath string
	FilePath      string
	Content       string
	Embedding     []float32
}

type JITContext struct {
	ActiveFilePath    string
	ActiveFileContent string
	CursorLine        int
	OpenBuffers       []string
}

type WorkspaceService interface {
	Register(ctx context.Context, workspacePath string) error
	Unregister(ctx context.Context, workspacePath string) error
	BuildContext(ctx context.Context, query string, jit *JITContext, workspacePath, embedModel string) (string, error)
}

type WorkspaceStore interface {
	DeleteWorkspace(ctx context.Context, workspacePath string) error
	InsertChunks(ctx context.Context, chunks []Chunk) error
	SearchChunks(ctx context.Context, workspacePath string, queryVector []float32, limit int) ([]Chunk, error)
}
