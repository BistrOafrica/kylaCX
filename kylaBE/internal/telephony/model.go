package telephony

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CallDirection enumerates the only two legal direction values.
type CallDirection string

const (
	DirectionInbound  CallDirection = "inbound"
	DirectionOutbound CallDirection = "outbound"
)

// CallStatus tracks the call's lifecycle. The PBX is the source of truth —
// our store transitions on receipt of ESL events (CHANNEL_ANSWER, CHANNEL_HANGUP).
type CallStatus string

const (
	StatusRinging  CallStatus = "ringing"
	StatusAnswered CallStatus = "answered"
	StatusEnded    CallStatus = "ended"
	StatusFailed   CallStatus = "failed"
)

// CallDisposition is the final outcome of a call — distinct from CallStatus
// which is the in-flight state. Set when the call ends.
type CallDisposition string

const (
	DispositionCompleted CallDisposition = "completed"
	DispositionNoAnswer  CallDisposition = "no_answer"
	DispositionBusy      CallDisposition = "busy"
	DispositionFailed    CallDisposition = "failed"
	DispositionVoicemail CallDisposition = "voicemail"
)

// Call is the persisted record of a call leg.
//
// ID matches the FreeSWITCH UUID for the leg so we can correlate ESL events
// to the row without a separate mapping table. Outbound calls get their ID
// from the originate response; inbound calls get theirs from the first
// CHANNEL_CREATE event.
type Call struct {
	ID          string `gorm:"type:uuid;primaryKey" json:"id"`
	OrgID       string `gorm:"type:uuid;not null;index" json:"org_id"`
	WorkspaceID string `gorm:"type:uuid;not null;index" json:"workspace_id"`
	Direction   string `gorm:"not null;index" json:"direction"`
	Status      string `gorm:"not null;default:'ringing';index" json:"status"`

	FromNumber string `gorm:"not null;default:''" json:"from_number"`
	ToNumber   string `gorm:"not null;default:''" json:"to_number"`

	AgentID    string `gorm:"type:uuid;index" json:"agent_id,omitempty"`
	ContactID  string `gorm:"type:uuid;index" json:"contact_id,omitempty"`
	QueueID    string `gorm:"type:uuid" json:"queue_id,omitempty"`
	TrunkID    string `gorm:"type:uuid" json:"trunk_id,omitempty"`
	IvrFlowID  string `gorm:"type:uuid" json:"ivr_flow_id,omitempty"`

	ConversationID string `gorm:"type:uuid" json:"conversation_id,omitempty"`
	DealID         string `gorm:"type:uuid" json:"deal_id,omitempty"`
	TicketID       string `gorm:"type:uuid" json:"ticket_id,omitempty"`

	RecordingEnabled bool   `gorm:"not null;default:false" json:"recording_enabled"`
	RecordingURL     string `json:"recording_url,omitempty"`

	StartedAt   time.Time  `gorm:"not null;default:now()" json:"started_at"`
	AnsweredAt  *time.Time `json:"answered_at,omitempty"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
	RingSeconds int        `gorm:"not null;default:0" json:"ring_seconds"`
	TalkSeconds int        `gorm:"not null;default:0" json:"talk_seconds"`

	HangupCause string `json:"hangup_cause,omitempty"`
	Disposition string `json:"disposition,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Call) TableName() string { return "calls" }

// BeforeCreate ensures the row has an ID even when the caller forgot to set
// one — but the normal path is for callers to set ID from the PBX-assigned
// UUID, so this is a safety net rather than a default.
func (c *Call) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	return nil
}

// CallEvent is one entry in a call's event log. event_type matches what the
// ESL event handler emits (started, answered, transferred, held, resumed,
// ended, note).
type CallEvent struct {
	ID        string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CallID    string          `gorm:"type:uuid;not null;index" json:"call_id"`
	OrgID     string          `gorm:"type:uuid;not null;index" json:"org_id"`
	EventType string          `gorm:"not null" json:"event_type"`
	Detail    json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"detail,omitempty"`
	At        time.Time       `gorm:"not null;default:now();index" json:"at"`
}

func (CallEvent) TableName() string { return "call_events" }

func (e *CallEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	return nil
}

// ── SIP infrastructure ──────────────────────────────────────────────────────

// SipDomain is a SIP realm. One default per org enforced by partial unique
// index in the migration.
type SipDomain struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID     string    `gorm:"type:uuid;not null;index" json:"org_id"`
	Domain    string    `gorm:"not null" json:"domain"`
	IsDefault bool      `gorm:"not null;default:false" json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

func (SipDomain) TableName() string { return "sip_domains" }

func (d *SipDomain) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.NewString()
	}
	return nil
}

// SipExtension is one agent's SIP identity. v1 enforces one extension per
// user; multi-device support comes later.
//
// PasswordHash stores a bcrypt-style hash of the SIP password. The plaintext
// is returned exactly once at creation time (so the admin can provision it
// in the PBX out-of-band) and never persisted in cleartext.
type SipExtension struct {
	ID               string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID            string     `gorm:"type:uuid;not null;index" json:"org_id"`
	WorkspaceID      string     `gorm:"type:uuid;not null;index" json:"workspace_id"`
	UserID           string     `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	DomainID         string     `gorm:"type:uuid" json:"domain_id,omitempty"`
	Extension        string     `gorm:"not null" json:"extension"`
	DisplayName      string     `gorm:"not null;default:''" json:"display_name,omitempty"`
	PasswordHash     string     `gorm:"not null;default:''" json:"-"`
	Status           string     `gorm:"not null;default:'unregistered'" json:"status"`
	LastRegistration *time.Time `json:"last_registration,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (SipExtension) TableName() string { return "sip_extensions" }

func (e *SipExtension) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	return nil
}

// SipTrunk is an outbound trunk for PSTN connectivity. The Password column
// holds the plaintext we send to the PBX gateway profile — it's necessary
// in cleartext because FreeSWITCH needs it to authenticate the trunk.
//
// The gRPC Get/List responses zero this field before returning.
type SipTrunk struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID       string    `gorm:"type:uuid;not null;index" json:"org_id"`
	Name        string    `gorm:"not null" json:"name"`
	GatewayName string    `gorm:"not null" json:"gateway_name"`
	Provider    string    `gorm:"not null;default:'custom'" json:"provider"`
	SipServer   string    `gorm:"not null;default:''" json:"sip_server"`
	Username    string    `gorm:"not null;default:''" json:"username"`
	Password    string    `gorm:"not null;default:''" json:"-"`
	FromURI     string    `gorm:"not null;default:''" json:"from_uri"`
	IsActive    bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (SipTrunk) TableName() string { return "sip_trunks" }

func (t *SipTrunk) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	return nil
}
