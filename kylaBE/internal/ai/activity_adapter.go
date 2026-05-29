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

// TranscriberAdapter wraps an LLMProvider and exposes its TranscribeAudio
// method when the provider implements it. Used by the telephony recording
// pipeline to swap providers without changing the call site.
//
// Returns a clear "not supported" error when the wrapped provider doesn't
// implement AudioTranscriber, so callers can fall through to the no-op
// transcription path (which leaves the recording with transcript=""+done).
type TranscriberAdapter struct {
	Provider LLMProvider
}

func (t *TranscriberAdapter) Name() string {
	if t == nil || t.Provider == nil {
		return "noop"
	}
	return t.Provider.Name()
}

func (t *TranscriberAdapter) TranscribeAudio(ctx context.Context, audio []byte, mime string) (string, error) {
	if t == nil || t.Provider == nil {
		return "", nil
	}
	if at, ok := t.Provider.(AudioTranscriber); ok {
		return at.TranscribeAudio(ctx, audio, mime)
	}
	return "", nil // provider doesn't support audio; treat as silent skip
}
