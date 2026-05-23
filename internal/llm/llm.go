package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/redhotvengeance/promptd/internal/promptd"
)

type Provider interface {
	Chat(ctx context.Context, systemPrompt string, history []promptd.Message, model string) (promptd.Stream, error)
	FIM(ctx context.Context, prefix, suffix, model string) (string, error)
	Task(ctx context.Context, systemPrompt string, history []promptd.Message, model string) (*promptd.AgentResponse, error)
	Embed(ctx context.Context, text, model string) ([]float32, error)
}

type Manager struct {
	providers map[string]Provider
	defaults promptd.Defaults
}

func NewManager(config *promptd.Config) *Manager {
	manager := &Manager{
		providers: make(map[string]Provider),
		defaults: config.Defaults,
	}

	for k, v := range config.Providers {
		p := NewProvider(&v)

		manager.providers[k] = p
	}

	return manager
}

func NewProvider(config *promptd.Provider) Provider {
	var provider Provider

	if config.Scheme == "openai" {
		provider = newOpenAIClient(config.Endpoint, config.Key)
	}

	return provider
}

func (m *Manager) resolve(override, def string) (Provider, string, error) {
	target := override
	if target == "" {
		target = def
	}

	parts := strings.SplitN(target, ":", 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid routing string '%s', expected format 'provider:model'", target)
	}

	providerName, modelName := parts[0], parts[1]

	provider, exists := m.providers[providerName]

	if !exists {
		return nil, "", fmt.Errorf("provider '%s' is not configured", providerName)
	}

	return provider, modelName, nil
}

func (m *Manager) Chat(ctx context.Context, systemPrompt string, history[]promptd.Message, providerOverride string) (promptd.Stream, error) {
	provider, model, err := m.resolve(providerOverride, m.defaults.Chat)
	if err != nil {
		return nil, err
	}

	return provider.Chat(ctx, systemPrompt, history, model)
}

func (m *Manager) FIM(ctx context.Context, prefix, suffix, providerOverride string) (string, error) {
	provider, model, err := m.resolve(providerOverride, m.defaults.FIM)
	if err != nil {
		return "", err
	}

	return provider.FIM(ctx, prefix, suffix, model)
}

func (m *Manager) Task(ctx context.Context, systemPrompt string, history []promptd.Message, providerOverride string) (*promptd.AgentResponse, error) {
	provider, model, err := m.resolve(providerOverride, m.defaults.Task)
	if err != nil {
		return nil, err
	}

	return provider.Task(ctx, systemPrompt, history, model)
}

func (m *Manager) Embed(ctx context.Context, text string) ([]float32, error) {
	provider, model, err := m.resolve("", m.defaults.Embed)
	if err != nil {
		return nil, err
	}

	return provider.Embed(ctx, text, model)
}
