package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// AnthropicProvider talks to the Anthropic Messages API. Provided as an
// alternate LLMProvider so operators can switch via the LLM_PROVIDER env var
// without touching code.
type AnthropicProvider struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewAnthropicProvider(apiKey, model string) (*AnthropicProvider, error) {
	if apiKey == "" {
		return nil, errors.New("anthropic: ANTHROPIC_API_KEY not set")
	}
	if model == "" {
		model = "claude-sonnet-4-6"
	}
	return &AnthropicProvider{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (p *AnthropicProvider) Name() string { return "anthropic" }

func (p *AnthropicProvider) Classify(ctx context.Context, text string, labels []string) (string, float64, error) {
	if text == "" || len(labels) == 0 {
		return "", 0, errors.New("classify: text and labels required")
	}
	sys := fmt.Sprintf(
		"Classify the user's text into one of these labels: %s. "+
			"Respond with ONLY this JSON: {\"label\":\"<one label>\",\"confidence\":<0..1>}.",
		strings.Join(labels, ", "),
	)
	raw, err := p.message(ctx, sys, text)
	if err != nil {
		return "", 0, err
	}
	var parsed struct {
		Label      string  `json:"label"`
		Confidence float64 `json:"confidence"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		if start := strings.Index(raw, "{"); start >= 0 {
			if end := strings.LastIndex(raw, "}"); end > start {
				_ = json.Unmarshal([]byte(raw[start:end+1]), &parsed)
			}
		}
	}
	if parsed.Label == "" {
		return labels[0], 0, fmt.Errorf("classify: could not parse model output: %q", raw)
	}
	return parsed.Label, parsed.Confidence, nil
}

func (p *AnthropicProvider) Summarize(ctx context.Context, text string, maxSentences int) (string, error) {
	if text == "" {
		return "", errors.New("summarize: text required")
	}
	sentenceHint := ""
	if maxSentences > 0 {
		sentenceHint = " in at most " + strconv.Itoa(maxSentences) + " sentences"
	}
	sys := "You are a concise summariser. Produce a summary" + sentenceHint + ". Plain text only."
	return p.message(ctx, sys, text)
}

func (p *AnthropicProvider) GenerateReply(ctx context.Context, history []string, prompt string) (string, error) {
	if prompt == "" {
		return "", errors.New("generate_reply: prompt required")
	}
	var sb strings.Builder
	for i, turn := range history {
		sb.WriteString(fmt.Sprintf("Turn %d: %s\n", i+1, turn))
	}
	sb.WriteString("\nInstruction: ")
	sb.WriteString(prompt)
	sys := "You are a customer-support assistant. Reply helpfully and concisely. Plain text only."
	return p.message(ctx, sys, sb.String())
}

// ── Messages API plumbing ────────────────────────────────────────────────────

type anthropicRequest struct {
	Model     string             `json:"model"`
	System    string             `json:"system,omitempty"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (p *AnthropicProvider) message(ctx context.Context, system, user string) (string, error) {
	body := anthropicRequest{
		Model:     p.model,
		System:    system,
		MaxTokens: 1024,
		Messages: []anthropicMessage{
			{Role: "user", Content: user},
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("anthropic: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("anthropic: build request: %w", err)
	}
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic: http: %w", err)
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("anthropic: status %d: %s", resp.StatusCode, string(respBytes))
	}
	var parsed anthropicResponse
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return "", fmt.Errorf("anthropic: decode: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("anthropic: %s: %s", parsed.Error.Type, parsed.Error.Message)
	}
	for _, c := range parsed.Content {
		if c.Type == "text" {
			return strings.TrimSpace(c.Text), nil
		}
	}
	return "", errors.New("anthropic: empty response")
}
