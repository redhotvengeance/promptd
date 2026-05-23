package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/redhotvengeance/promptd/internal/promptd"
)

type openAIClient struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type openAIStream struct {
	body   io.ReadCloser
	reader *bufio.Reader
}

type openAIFunction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

type openAITool struct {
	Type     string         `json:"type"`
	Function openAIFunction `json:"function"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content   string `json:"content"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
}

func newOpenAIClient(endpoint, apiKey string) *openAIClient {
	endpoint = strings.TrimSuffix(endpoint, "/")

	return &openAIClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *openAIClient) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/models", nil)
	if err != nil {
		return nil, err
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failted to list models, status: %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range result.Data {
		models = append(models, m.ID)
	}

	return models, nil
}

func (c *openAIClient) Chat(ctx context.Context, systemPrompt string, history []promptd.Message, model string) (promptd.Stream, error) {
	messages := make([]openAIMessage, 0, len(history)+1)

	if systemPrompt != "" {
		messages = append(messages, openAIMessage{Role: "system", Content: systemPrompt})
	}

	for _, msg := range history {
		messages = append(messages, openAIMessage{Role: string(msg.Role), Content: msg.Content})
	}

	payload := openAIChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("provider returned status: %d, body: %s", resp.StatusCode, string(errBody))
	}

	return &openAIStream{
		body:   resp.Body,
		reader: bufio.NewReader(resp.Body),
	}, nil
}

func (c *openAIClient) Embed(ctx context.Context, text, model string) ([]float32, error) {
	payload := map[string]any{
		"input": text,
		"model": model,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding failed with status: %d, body: %s", resp.StatusCode, string(errBody))
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("API returned empty embedding data")
	}

	return result.Data[0].Embedding, nil
}

func (c *openAIClient) FIM(ctx context.Context, prefix, suffix, model string) (string, error) {
	payload := map[string]any{
		"model":       model,
		"prompt":      prefix,
		"suffix":      suffix,
		"max_tokens":  256,
		"temperature": 0.1,
		"stream":      false,
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("FIM failed with status: %d, body %s", resp.StatusCode, string(errBody))
	}

	var result struct {
		Choices []struct {
			Text string `json:"text"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", nil
	}

	return result.Choices[0].Text, nil
}

func (c *openAIClient) Task(ctx context.Context, systemPrompt string, history []promptd.Message, model string) (*promptd.AgentResponse, error) {
	messages := make([]map[string]any, 0, len(history)+1)

	if systemPrompt != "" {
		messages = append(messages, map[string]any{"role": "system", "content": systemPrompt})
	}

	for _, msg := range history {
		messages = append(messages, map[string]any{"role": string(msg.Role), "content": msg.Content})
	}

	tools := []openAITool{
		{
			Type: "function",
			Function: openAIFunction{
				Name:        "execute_command",
				Description: "Run a shell command in the workspace.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"command": map[string]any{
							"type":        "string",
							"description": "The bash command to run.",
						},
					},
					"required": []string{"command"},
				},
			},
		},
		{
			Type: "function",
			Function: openAIFunction{
				Name:        "edit_file",
				Description: "Replace the contents of a file.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"filepath": map[string]any{"type": "string"},
						"content":  map[string]any{"type": "string"},
					},
					"required": []string{"filepath", "content"},
				},
			},
		},
	}

	payload := map[string]any{
		"model":       model,
		"messages":    messages,
		"tools":       tools,
		"temperature": 0.0,
		"stream":      false,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("agent task failed with status: %d, body: %s", resp.StatusCode, string(errBody))
	}

	var result openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("provider returned empty choices array")
	}

	msg := result.Choices[0].Message

	agentResp := &promptd.AgentResponse{
		Text:      msg.Content,
		ToolCalls: make([]promptd.ToolCall, 0, len(msg.ToolCalls)),
	}

	for _, tc := range msg.ToolCalls {
		agentResp.ToolCalls = append(agentResp.ToolCalls, promptd.ToolCall{
			ID:   tc.ID,
			Name: tc.Function.Name,
			Args: tc.Function.Arguments,
		})
	}

	return agentResp, nil
}

func (s *openAIStream) Recv() (string, error) {
	var chunk struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}

	for {
		line, err := s.reader.ReadBytes('\n')
		if err != nil {
			return "", err
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if bytes.HasPrefix(line, []byte("data: [DONE]")) {
			return "", io.EOF
		}

		if jsonPayload, found := bytes.CutPrefix(line, []byte("data: ")); found {
			chunk.Choices = nil
			if err := json.Unmarshal(jsonPayload, &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				return chunk.Choices[0].Delta.Content, nil
			}
		}
	}
}

func (s *openAIStream) Close() error {
	return s.body.Close()
}
