package ai

import "context"

// ActivityAdapter wraps an LLMProvider and satisfies the
// `automation/activities.AIClassifier` interface. Letting the automation
// worker call the provider directly avoids an extra gRPC hop for in-process
// workflow execution.
//
// The activities package depends on its own interface (defined in
// activities/deps.go), not on this concrete type — this adapter just happens
// to satisfy that interface structurally.
type ActivityAdapter struct {
	Provider LLMProvider
}

// orgID is accepted to match the activities.AIClassifier signature but is
// currently unused — multi-tenant routing of LLM requests (per-org keys,
// per-org budget) will be added when the AI service grows beyond this minimum
// surface.
func (a *ActivityAdapter) Classify(ctx context.Context, _ string, text string, labels []string) (string, float64, error) {
	return a.Provider.Classify(ctx, text, labels)
}

func (a *ActivityAdapter) Summarize(ctx context.Context, _ string, text string) (string, error) {
	return a.Provider.Summarize(ctx, text, 0)
}

func (a *ActivityAdapter) GenerateReply(ctx context.Context, _ string, history []string, prompt string) (string, error) {
	return a.Provider.GenerateReply(ctx, history, prompt)
}
