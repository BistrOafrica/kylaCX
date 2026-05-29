package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

// OpenAIProvider talks to OpenAI's Chat Completions API. Implemented as a
// direct HTTP client (no SDK dependency) so the dep footprint stays small —
// the API surface we use is two endpoints and three request shapes.
type OpenAIProvider struct {
	apiKey  string
	model   string
	baseURL string
	http    *http.Client
}

// NewOpenAIProvider constructs an OpenAI provider. baseURL is optional and
// defaults to the public API; model defaults to gpt-4o-mini when empty.
// Returns an error when apiKey is empty so callers can fall back to NoopProvider.
func NewOpenAIProvider(apiKey, model, baseURL string) (*OpenAIProvider, error) {
	if apiKey == "" {
		return nil, errors.New("openai: OPENAI_API_KEY not set")
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (p *OpenAIProvider) Name() string { return "openai" }

// ── Skill implementations ────────────────────────────────────────────────────

func (p *OpenAIProvider) Classify(ctx context.Context, text string, labels []string) (string, float64, error) {
	if text == "" || len(labels) == 0 {
		return "", 0, errors.New("classify: text and labels required")
	}
	sys := fmt.Sprintf(
		"You classify text into one of these labels: %s. "+
			"Respond with ONLY this JSON shape: {\"label\":\"<one of the labels>\",\"confidence\":<0..1>}.",
		strings.Join(labels, ", "),
	)
	raw, err := p.chat(ctx, sys, text, true)
	if err != nil {
		return "", 0, err
	}
	var parsed struct {
		Label      string  `json:"label"`
		Confidence float64 `json:"confidence"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		// Some models occasionally emit prose around the JSON. Try a substring
		// extraction before giving up.
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

func (p *OpenAIProvider) Summarize(ctx context.Context, text string, maxSentences int) (string, error) {
	if text == "" {
		return "", errors.New("summarize: text required")
	}
	sentenceHint := ""
	if maxSentences > 0 {
		sentenceHint = " in at most " + strconv.Itoa(maxSentences) + " sentences"
	}
	sys := "You are a concise summariser. Produce a summary" + sentenceHint + ". Plain text only."
	return p.chat(ctx, sys, text, false)
}

func (p *OpenAIProvider) GenerateReply(ctx context.Context, history []string, prompt string) (string, error) {
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
	return p.chat(ctx, sys, sb.String(), false)
}

// ── HTTP plumbing ────────────────────────────────────────────────────────────

type openAIChatRequest struct {
	Model          string              `json:"model"`
	Messages       []openAIChatMessage `json:"messages"`
	Temperature    float64             `json:"temperature"`
	ResponseFormat *openAIRespFmt      `json:"response_format,omitempty"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRespFmt struct {
	Type string `json:"type"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message openAIChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// chat issues a single Chat Completions call. When jsonMode is true the model
// is asked to return strict JSON via response_format=json_object.
func (p *OpenAIProvider) chat(ctx context.Context, system, user string, jsonMode bool) (string, error) {
	body := openAIChatRequest{
		Model: p.model,
		Messages: []openAIChatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0.2,
	}
	if jsonMode {
		body.ResponseFormat = &openAIRespFmt{Type: "json_object"}
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("openai: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("openai: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai: http: %w", err)
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("openai: status %d: %s", resp.StatusCode, string(respBytes))
	}
	var parsed openAIChatResponse
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return "", fmt.Errorf("openai: decode: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("openai: %s: %s", parsed.Error.Type, parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("openai: empty choices")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

// ── Whisper (audio transcription) ────────────────────────────────────────────

// TranscribeAudio implements AudioTranscriber via OpenAI's /v1/audio/transcriptions
// endpoint (whisper-1 model). The endpoint expects multipart/form-data with
// the audio file under "file" and the model name under "model".
//
// Errors from the API are returned verbatim; the caller decides whether to
// retry, log, or surface them.
func (p *OpenAIProvider) TranscribeAudio(ctx context.Context, audio []byte, mime string) (string, error) {
	if len(audio) == 0 {
		return "", errors.New("openai whisper: empty audio")
	}
	if mime == "" {
		mime = "audio/wav"
	}

	// Build multipart body.
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	if err := mw.WriteField("model", "whisper-1"); err != nil {
		return "", fmt.Errorf("multipart model: %w", err)
	}
	filename := "audio." + extFromMIME(mime)
	hdr := textproto.MIMEHeader{}
	hdr.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename=%q`, filename))
	hdr.Set("Content-Type", mime)
	part, err := mw.CreatePart(hdr)
	if err != nil {
		return "", fmt.Errorf("multipart file: %w", err)
	}
	if _, err := part.Write(audio); err != nil {
		return "", fmt.Errorf("multipart file write: %w", err)
	}
	if err := mw.Close(); err != nil {
		return "", fmt.Errorf("multipart close: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/audio/transcriptions", &body)
	if err != nil {
		return "", fmt.Errorf("whisper request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	// Whisper jobs can take 60+ seconds for long files; respect caller deadline
	// or fall back to the provider-wide 60s client default.
	resp, err := p.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("whisper http: %w", err)
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("whisper status %d: %s", resp.StatusCode, string(respBytes))
	}
	var parsed struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return "", fmt.Errorf("whisper decode: %w", err)
	}
	return strings.TrimSpace(parsed.Text), nil
}

func extFromMIME(mime string) string {
	switch mime {
	case "audio/wav", "audio/x-wav":
		return "wav"
	case "audio/mpeg":
		return "mp3"
	case "audio/m4a", "audio/mp4":
		return "m4a"
	case "audio/ogg":
		return "ogg"
	case "audio/webm":
		return "webm"
	default:
		return "wav"
	}
}
