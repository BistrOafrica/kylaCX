package activities

import (
	"context"
	"net/http"

	"kyla-be/internal/communication"
	"kyla-be/internal/objectcore"
)

// Notifier is the minimal interface the SendNotificationActivity needs.
// Defined here so the activity doesn't import a notification package that
// doesn't exist yet — the worker can pass a no-op or a real implementation
// once internal/notification ships.
type Notifier interface {
	Send(ctx context.Context, orgID, userID, kind, title, body string, payload map[string]interface{}) error
}

// AIClassifier is the minimal interface RunAISkillActivity needs.
// Wired up in step 7 when internal/ai lands.
type AIClassifier interface {
	Classify(ctx context.Context, orgID, text string, labels []string) (label string, confidence float64, err error)
	Summarize(ctx context.Context, orgID, text string) (summary string, err error)
	GenerateReply(ctx context.Context, orgID string, history []string, prompt string) (reply string, err error)
}

// SLATimerStarter is the subset of the communication SLA engine the
// set_sla activity needs. Defined here so the activity doesn't pull in the
// engine's full periodic-scan machinery.
type SLATimerStarter interface {
	StartTimer(conv *communication.Conversation) error
}

// Deps is the dependency bundle injected into all activities at worker
// registration. Each activity struct embeds (or holds a reference to) Deps so
// it can talk to the rest of the platform without reaching into globals.
//
// Activities are registered on the Temporal worker as method receivers on the
// per-activity structs (e.g. UpdateObjectActivity), which lets Temporal use the
// struct's method name as the activity name.
type Deps struct {
	ObjectStore       *objectcore.ObjectCoreStore
	AdapterRegistry   *communication.AdapterRegistry
	ConversationStore *communication.ConversationStore
	MessageStore      *communication.MessageStore
	SLATimer          SLATimerStarter
	HTTPClient        *http.Client
	Notifier          Notifier
	AI                AIClassifier
}
