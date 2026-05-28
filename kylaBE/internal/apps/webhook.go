package apps

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Webhook represents a registered HTTP endpoint that receives platform events.
// It is scoped to an App (token-bearing API client) and optionally to a workspace.
type Webhook struct {
	ID          uuid.UUID      `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AppID       uuid.UUID      `gorm:"type:uuid;index;not null"` // owning API app
	OrgID       string         `gorm:"index;not null"`           // organisation scope
	WorkspaceID *string        `gorm:"type:uuid;index"`          // optional workspace scope
	URL         string         `gorm:"not null"`                 // delivery endpoint
	Events      pq.StringArray `gorm:"type:text[];not null"`     // e.g. ["ticket.created"]
	Secret      string         `gorm:"not null"`                 // HMAC-SHA256 signing secret
	IsActive    bool           `gorm:"not null;default:true"`
	CreatedBy   string         `gorm:"not null"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
}

// ── REST request / response types ────────────────────────────────────────────

// RegisterWebhookRequest is the JSON body for POST /api/v1/webhooks.
type RegisterWebhookRequest struct {
	URL         string   `json:"url"          binding:"required,url"`
	Events      []string `json:"events"       binding:"required,min=1"`
	WorkspaceID string   `json:"workspace_id"` // optional
}

// UpdateWebhookRequest is the JSON body for PUT /api/v1/webhooks/:id.
type UpdateWebhookRequest struct {
	URL      string   `json:"url"      binding:"omitempty,url"`
	Events   []string `json:"events"   binding:"omitempty,min=1"`
	IsActive *bool    `json:"is_active"`
}

// WebhookResponse is returned by every webhook endpoint.
// The Secret field is only populated on creation.
type WebhookResponse struct {
	ID          string   `json:"id"`
	AppID       string   `json:"app_id"`
	OrgID       string   `json:"org_id"`
	WorkspaceID string   `json:"workspace_id,omitempty"`
	URL         string   `json:"url"`
	Events      []string `json:"events"`
	Secret      string   `json:"secret,omitempty"` // only on creation
	IsActive    bool     `json:"is_active"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

func webhookToResponse(w *Webhook, exposeSecret bool) WebhookResponse {
	resp := WebhookResponse{
		ID:        w.ID.String(),
		AppID:     w.AppID.String(),
		OrgID:     w.OrgID,
		URL:       w.URL,
		Events:    []string(w.Events),
		IsActive:  w.IsActive,
		CreatedAt: w.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: w.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if w.WorkspaceID != nil {
		resp.WorkspaceID = *w.WorkspaceID
	}
	if exposeSecret {
		resp.Secret = w.Secret
	}
	return resp
}
