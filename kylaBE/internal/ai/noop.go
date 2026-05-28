package ai

import "context"

// NoopProvider is the fallback used when no LLM credentials are configured.
// It returns deterministic placeholder responses so workflows still complete
// in dev environments without a real provider.
type NoopProvider struct{}

func (NoopProvider) Name() string { return "noop" }

func (NoopProvider) Classify(_ context.Context, _ string, labels []string) (string, float64, error) {
	if len(labels) == 0 {
		return "", 0, nil
	}
	return labels[0], 0, nil
}

func (NoopProvider) Summarize(_ context.Context, text string, _ int) (string, error) {
	if len(text) <= 120 {
		return text, nil
	}
	return text[:120] + "…", nil
}

func (NoopProvider) GenerateReply(_ context.Context, _ []string, prompt string) (string, error) {
	return "[noop AI] " + prompt, nil
}
