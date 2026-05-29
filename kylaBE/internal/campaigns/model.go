package campaigns

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CampaignStatus enumerates valid lifecycle states.
type CampaignStatus string

const (
	StatusDraft     CampaignStatus = "draft"
	StatusScheduled CampaignStatus = "scheduled"
	StatusRunning   CampaignStatus = "running"
	StatusPaused    CampaignStatus = "paused"
	StatusCompleted CampaignStatus = "completed"
	StatusCancelled CampaignStatus = "cancelled"
)

// RecipientStatus enumerates valid per-recipient states.
type RecipientStatus string

const (
	RecipientQueued    RecipientStatus = "queued"
	RecipientSent      RecipientStatus = "sent"
	RecipientDelivered RecipientStatus = "delivered"
	RecipientRead      RecipientStatus = "read"
	RecipientFailed    RecipientStatus = "failed"
)

// ScheduleMode enumerates supported schedule kinds.
type ScheduleMode string

const (
	ScheduleImmediate ScheduleMode = "immediate"
	ScheduleOnce      ScheduleMode = "scheduled_once"
	ScheduleRecurring ScheduleMode = "recurring"
)

// AudienceKind enumerates how an audience resolves to recipient rows.
type AudienceKind string

const (
	AudienceObjectQuery AudienceKind = "object_query"
	AudienceExplicit    AudienceKind = "explicit"
)

// Campaign is the channel-agnostic broadcast record.
//
// `Audience`, `Schedule`, and `Payload` are JSONB so the same row shape
// serves WhatsApp / SMS / email blasts. Per-channel send behaviour is in the
// activities that read `Payload`.
type Campaign struct {
	ID          string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID       string `gorm:"type:uuid;not null;index"                       json:"org_id"`
	WorkspaceID string `gorm:"type:uuid;not null;index"                       json:"workspace_id"`
	Name        string `gorm:"not null"                                       json:"name"`
	Description string `gorm:"not null;default:''"                            json:"description"`
	Channel     string `gorm:"not null;index"                                 json:"channel"`
	Status      string `gorm:"not null;default:'draft';index"                 json:"status"`

	Audience json.RawMessage `gorm:"type:jsonb;not null;default:'{}'"` // CampaignAudience
	Schedule json.RawMessage `gorm:"type:jsonb;not null;default:'{}'"` // CampaignSchedule
	Payload  json.RawMessage `gorm:"type:jsonb;not null;default:'{}'"` // channel-specific

	// Denormalised counters — see store.RecomputeStats for refresh logic.
	AudienceSize   int `gorm:"not null;default:0" json:"audience_size"`
	QueuedCount    int `gorm:"not null;default:0" json:"queued_count"`
	SentCount      int `gorm:"not null;default:0" json:"sent_count"`
	DeliveredCount int `gorm:"not null;default:0" json:"delivered_count"`
	ReadCount      int `gorm:"not null;default:0" json:"read_count"`
	FailedCount    int `gorm:"not null;default:0" json:"failed_count"`

	TemporalWorkflowID string `json:"temporal_workflow_id,omitempty"`
	TemporalScheduleID string `json:"temporal_schedule_id,omitempty"`

	CreatedBy string    `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Campaign) TableName() string { return "campaigns" }

func (c *Campaign) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	return nil
}

// CampaignAudience is the deserialised shape of Campaign.Audience.
// Stored as JSONB on the row so list/search queries against audiences stay
// indexable with JSONB operators.
type CampaignAudience struct {
	Kind      AudienceKind           `json:"kind"`
	TypeSlug  string                 `json:"type_slug,omitempty"`
	Filter    map[string]interface{} `json:"filter,omitempty"`
	ObjectIDs []string               `json:"object_ids,omitempty"`
}

// CampaignSchedule is the deserialised shape of Campaign.Schedule.
type CampaignSchedule struct {
	Mode     ScheduleMode `json:"mode"`
	StartAt  *time.Time   `json:"start_at,omitempty"`
	Cron     string       `json:"cron,omitempty"`
	Timezone string       `json:"timezone,omitempty"`
}

// CampaignRecipient is one resolved audience row.
type CampaignRecipient struct {
	ID          string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CampaignID  string `gorm:"type:uuid;not null;index"                       json:"campaign_id"`
	OrgID       string `gorm:"type:uuid;not null;index"                       json:"org_id"`
	ObjectID    string `gorm:"type:uuid;not null"                             json:"object_id"`
	ContactRef  string `gorm:"not null"                                       json:"contact_ref"`
	Status      string `gorm:"not null;default:'queued';index"                json:"status"`
	ExternalID  string `json:"external_id,omitempty"`
	Error       string `json:"error,omitempty"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (CampaignRecipient) TableName() string { return "campaign_recipients" }

func (r *CampaignRecipient) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}
	return nil
}

// WhatsAppTemplate is the local mirror of a Meta-approved template.
// Per-org uniqueness on (name, language) matches Meta's composite key — see
// the UNIQUE index in the migration.
type WhatsAppTemplate struct {
	ID             string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID          string `gorm:"type:uuid;not null;index"                       json:"org_id"`
	Name           string `gorm:"not null"                                       json:"name"`
	Language       string `gorm:"not null"                                       json:"language"`
	Category       string `gorm:"not null"                                       json:"category"`
	Status         string `gorm:"not null;index"                                 json:"status"`
	Header         string `gorm:"not null;default:''"                            json:"header,omitempty"`
	Body           string `gorm:"not null;default:''"                            json:"body,omitempty"`
	Footer         string `gorm:"not null;default:''"                            json:"footer,omitempty"`
	PhoneNumberID  string `gorm:"not null;default:''"                            json:"phone_number_id,omitempty"`
	WabaID         string `gorm:"not null;default:''"                            json:"waba_id,omitempty"`
	MetaTemplateID string `gorm:"not null;default:''"                            json:"meta_template_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (WhatsAppTemplate) TableName() string { return "whatsapp_templates" }

func (t *WhatsAppTemplate) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	return nil
}
