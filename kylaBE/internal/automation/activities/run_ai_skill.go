package activities

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"kyla-be/shared/events"
)

// RunAISkillActivity invokes an AI capability (classify / summarize /
// generate_reply) from a workflow.
//
// Expected node.Config keys:
//
//	skill   string — required; "classify" | "summarize" | "generate_reply"
//	text    string — input text (used by classify and summarize)
//	labels  []string — optional; candidate labels for classify
//	history []string — optional; prior turns for generate_reply
//	prompt  string — optional; user prompt for generate_reply
//
// Returns a string result whose interpretation depends on skill:
//   - classify       → "<label>:<confidence>"
//   - summarize      → summary text
//   - generate_reply → generated reply text
type RunAISkillActivity struct {
	Deps Deps
}

func (a *RunAISkillActivity) RunAISkill(ctx context.Context, params map[string]interface{}, event events.DomainEvent) (string, error) {
	if a.Deps.AI == nil {
		return "", errors.New("run_ai_skill: AI dependency missing — wire internal/ai in step 7")
	}
	skill, _ := params["skill"].(string)
	if skill == "" {
		return "", errors.New("run_ai_skill: skill is required")
	}

	switch skill {
	case "classify":
		text, _ := params["text"].(string)
		labels := stringSlice(params["labels"])
		if text == "" || len(labels) == 0 {
			return "", errors.New("run_ai_skill[classify]: text and labels are required")
		}
		label, conf, err := a.Deps.AI.Classify(ctx, event.OrgID, text, labels)
		if err != nil {
			return "", fmt.Errorf("run_ai_skill[classify]: %w", err)
		}
		return fmt.Sprintf("%s:%.2f", label, conf), nil

	case "summarize":
		text, _ := params["text"].(string)
		if text == "" {
			return "", errors.New("run_ai_skill[summarize]: text is required")
		}
		return a.Deps.AI.Summarize(ctx, event.OrgID, text)

	case "generate_reply":
		history := stringSlice(params["history"])
		prompt, _ := params["prompt"].(string)
		if prompt == "" {
			return "", errors.New("run_ai_skill[generate_reply]: prompt is required")
		}
		return a.Deps.AI.GenerateReply(ctx, event.OrgID, history, prompt)

	default:
		return "", fmt.Errorf("run_ai_skill: unknown skill %q", skill)
	}
}

func stringSlice(v interface{}) []string {
	if v == nil {
		return nil
	}
	if ss, ok := v.([]string); ok {
		return ss
	}
	if anys, ok := v.([]interface{}); ok {
		out := make([]string, 0, len(anys))
		for _, x := range anys {
			if s, ok := x.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
