package promptd

import "context"

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

type Config struct {
	Defaults  Defaults
	Providers map[string]Provider
}

type Datastore interface {
	Messages() MessageStore
	Threads() ThreadStore
	Workspaces() WorkspaceStore
}

type Defaults struct {
	Chat  string
	FIM   string
	Task  string
	Embed string
}

type Provider struct {
	Scheme   string
	Endpoint string
	Key      string
}

type AgentResponse struct {
	Text       string
	ToolCalls  []ToolCall
	IsFinished bool
}

type ToolCall struct {
	ID string
	Name string
	Args string
}

type TaskUpdate struct {
	Status string
	Text   string
}

type Stream interface {
	Recv() (string, error) // returns text chunk or io.EOF
	Close() error
}

type LLM interface {
	Chat(ctx context.Context, systemPrompt string, history []Message, providerOverride string) (Stream, error)
	FIM(ctx context.Context, prefix, suffix, providerOverride string) (string, error)
	Task(ctx context.Context, systemPrompt string, history []Message, providerOverride string) (*AgentResponse, error)
	Embed(ctx context.Context, text, providerOverride string) ([]float32, error)
}
