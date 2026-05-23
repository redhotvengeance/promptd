package promptd

import "time"

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

type Message struct {
	ID        string
	ThreadID  string
	Role      Role
	Content   string
	CreatedAt time.Time
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

type Stream interface {
	Recv() (string, error) // returns text chunk or io.EOF
	Close() error
}
