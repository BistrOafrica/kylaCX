package ai

import "context"

// LLMProvider is the abstraction every concrete provider (OpenAI, Anthropic,
// etc.) implements. Activities and the AIService gRPC server both depend on
// this interface, never on a specific provider — switching providers is a
// config change, not a code change.
//
// Implementations are expected to be safe for concurrent use.
type LLMProvider interface {
	// Name returns a short identifier ("openai", "anthropic", "noop").
	Name() string

	// Classify returns the best-matching label from the supplied candidate
	// list together with a 0..1 confidence score. Providers that cannot
	// produce a confidence value should return 1.0 for the chosen label.
	Classify(ctx context.Context, text string, labels []string) (label string, confidence float64, err error)

	// Summarize returns a concise summary of the supplied text.
	// maxSentences is a hint; providers may exceed it.
	Summarize(ctx context.Context, text string, maxSentences int) (string, error)

	// GenerateReply produces a single reply given conversation history (oldest
	// first) and an instruction or user prompt.
	GenerateReply(ctx context.Context, history []string, prompt string) (string, error)
}
