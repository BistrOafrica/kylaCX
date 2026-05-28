package communication

import (
	"encoding/json"
	"time"
)

// ── Channel / Status / Priority constants ─────────────────────────────────────

const (
	ChannelWhatsApp  = "whatsapp"
	ChannelEmail     = "email"
	ChannelSMS       = "sms"
	ChannelVoice     = "voice"
	ChannelWebChat   = "webchat"
	ChannelInstagram = "instagram"
	ChannelMessenger = "messenger"

	StatusOpen     = "open"
	StatusPending  = "pending"
	StatusResolved = "resolved"
	StatusSnoozed  = "snoozed"

	PriorityLow    = "low"
	PriorityNormal = "normal"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"

	SenderAgent   = "agent"
	SenderContact = "contact"
	SenderBot     = "bot"
	SenderSystem  = "system"

	ContentText        = "text"
	ContentImage       = "image"
	ContentAudio       = "audio"
	ContentVideo       = "video"
	ContentFile        = "file"
	ContentTemplate    = "template"
	ContentInteractive = "interactive"

	ContentTypeText = ContentText // Alias for consistency

	MsgStatusPending   = "pending"
	MsgStatusSent      = "sent"
	MsgStatusDelivered = "delivered"
	MsgStatusRead      = "read"
	MsgStatusFailed    = "failed"
	MsgStatusReceived  = "received"
)

// ── Conversation ──────────────────────────────────────────────────────────────

// Conversation is the inbox row for a multi-channel customer thread.
// Its ID is shared with the corresponding Object Core record (same UUID)
// so the conversation gets full timeline, custom fields, and relations.
type Conversation struct {
	// ID is the primary key and also equals the Object Core objects.id for this conversation.
	ID           string          `gorm:"type:uuid;primaryKey"               json:"id"`
	OrgID        string          `gorm:"type:uuid;not null;index"            json:"org_id"`
	WorkspaceID  string          `gorm:"type:uuid;not null;index"            json:"workspace_id"`
	Channel      string          `gorm:"not null;index"                     json:"channel"`
	ChannelRef   string          `gorm:"index"                              json:"channel_ref,omitempty"`
	ContactID    string          `gorm:"type:uuid;index"                    json:"contact_id,omitempty"`
	AssignedTo   string          `gorm:"type:uuid;index"                    json:"assigned_to,omitempty"`
	TeamID       string          `gorm:"type:uuid;index"                    json:"team_id,omitempty"`
	Status       string          `gorm:"not null;default:'open';index"      json:"status"`
	Priority     string          `gorm:"not null;default:'normal'"          json:"priority"`
	Subject      string          `json:"subject,omitempty"`
	SLADeadline  *time.Time      `json:"sla_deadline,omitempty"`
	SnoozedUntil *time.Time      `json:"snoozed_until,omitempty"`
	ResolvedAt   *time.Time      `json:"resolved_at,omitempty"`
	Meta         json.RawMessage `gorm:"type:jsonb;not null;default:'{}'"   json:"meta,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// ── Message ───────────────────────────────────────────────────────────────────

// Message is a single message within a Conversation.
type Message struct {
	ID             string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConversationID string          `gorm:"type:uuid;not null;index"                       json:"conversation_id"`
	SenderID       string          `gorm:"type:uuid;index"                                json:"sender_id,omitempty"`
	SenderType     string          `gorm:"not null;default:'agent'"                       json:"sender_type"`
	Channel        string          `gorm:"not null"                                       json:"channel"`
	ContentType    string          `gorm:"not null;default:'text'"                        json:"content_type"`
	Content        json.RawMessage `gorm:"type:jsonb;not null"                            json:"content"`
	Status         string          `gorm:"not null;default:'sent'"                        json:"status"`
	ExternalID     string          `gorm:"index"                                          json:"external_id,omitempty"`
	CreatedAt      time.Time       `gorm:"index"                                          json:"created_at"`
}
