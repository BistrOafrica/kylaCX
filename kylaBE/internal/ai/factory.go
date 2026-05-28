package ai

import (
	"log"
	"strings"
)

// ProviderConfig holds the LLM provider selection and per-provider creds.
// Sourced from config.EnvConfigs at main.go startup.
type ProviderConfig struct {
	Provider        string // "openai" | "anthropic" | "noop"
	OpenAIAPIKey    string
	OpenAIModel     string
	OpenAIBaseURL   string
	AnthropicAPIKey string
	AnthropicModel  string
}

// NewProvider returns a concrete LLMProvider based on cfg.Provider. Falls back
// to NoopProvider when the chosen provider is unavailable (missing key, etc.)
// so the binary always boots and workflows always complete.
func NewProvider(cfg ProviderConfig) LLMProvider {
	switch strings.ToLower(cfg.Provider) {
	case "anthropic":
		if p, err := NewAnthropicProvider(cfg.AnthropicAPIKey, cfg.AnthropicModel); err == nil {
			log.Printf("ai: using anthropic provider (model=%s)", p.model)
			return p
		} else {
			log.Printf("ai: anthropic provider unavailable, falling back to noop: %v", err)
		}
	case "noop":
		log.Println("ai: explicitly using noop provider")
		return NoopProvider{}
	default:
		// Default: openai.
		if p, err := NewOpenAIProvider(cfg.OpenAIAPIKey, cfg.OpenAIModel, cfg.OpenAIBaseURL); err == nil {
			log.Printf("ai: using openai provider (model=%s)", p.model)
			return p
		} else {
			log.Printf("ai: openai provider unavailable, falling back to noop: %v", err)
		}
	}
	return NoopProvider{}
}
