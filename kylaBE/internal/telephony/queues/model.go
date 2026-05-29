// Package queues implements the Phase 5d call queue runtime. Lives under
// internal/telephony/ to share the PBXController abstraction without a
// circular import between the telephony and queues packages.
package queues

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Strategy enumerates the supported routing algorithms.
type Strategy string

const (
	StrategyRoundRobin  Strategy = "round_robin"
	StrategyLongestIdle Strategy = "longest_idle"
)

// EntryStatus tracks where a waiting call is in the queue lifecycle.
type EntryStatus string

const (
	EntryWaiting    EntryStatus = "waiting"    // in the queue, not yet rung an agent
	EntryRinging    EntryStatus = "ringing"    // ringing an assigned agent
	EntryConnected  EntryStatus = "connected"  // agent answered; bridged
	EntryAbandoned  EntryStatus = "abandoned"  // caller hung up before answer
	EntryOverflow   EntryStatus = "overflow"   // overflowed to voicemail/transfer/hangup
	EntryTimeout    EntryStatus = "timeout"    // max_wait_seconds elapsed
)

// OverflowAction is what happens when the queue's max wait expires.
type OverflowAction string

const (
	OverflowVoicemail OverflowAction = "voicemail"
	OverflowHangup    OverflowAction = "hangup"
	OverflowTransfer  OverflowAction = "transfer"
)

// Queue is the persisted queue configuration.
type Queue struct {
	ID          string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID       string `gorm:"type:uuid;not null;index" json:"org_id"`
	WorkspaceID string `gorm:"type:uuid;not null;index" json:"workspace_id"`
	Name        string `gorm:"not null" json:"name"`
	Description string `gorm:"not null;default:''" json:"description,omitempty"`

	Strategy        string `gorm:"not null;default:'longest_idle'" json:"strategy"`
	MOHPath         string `gorm:"column:moh_path;not null;default:'local_stream://moh'" json:"moh_path"`
	MaxWaitSeconds  int    `gorm:"not null;default:600" json:"max_wait_seconds"`
	OverflowAction  string `gorm:"not null;default:'voicemail'" json:"overflow_action"`
	OverflowTarget  string `gorm:"not null;default:''" json:"overflow_target"`

	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Queue) TableName() string { return "call_queues" }

func (q *Queue) BeforeCreate(tx *gorm.DB) error {
	if q.ID == "" {
		q.ID = uuid.NewString()
	}
	return nil
}

// Membership maps an agent to a queue. is_active is the agent's self-service
// pause/resume toggle; agentops.AgentStatus governs overall availability.
type Membership struct {
	ID       string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	QueueID  string `gorm:"type:uuid;not null;index" json:"queue_id"`
	OrgID    string `gorm:"type:uuid;not null;index" json:"org_id"`
	UserID   string `gorm:"type:uuid;not null" json:"user_id"`
	Priority int    `gorm:"not null;default:0" json:"priority"`
	IsActive bool   `gorm:"not null;default:true" json:"is_active"`

	LastCallEndedAt *time.Time `json:"last_call_ended_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (Membership) TableName() string { return "call_queue_members" }

func (m *Membership) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return nil
}

// Entry is one caller currently in (or recently in) a queue. Lifecycle ends
// when the entry is connected, abandoned, overflowed, or timed out.
type Entry struct {
	ID              string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	QueueID         string `gorm:"type:uuid;not null;index" json:"queue_id"`
	CallID          string `gorm:"type:uuid;not null;index" json:"call_id"`
	OrgID           string `gorm:"type:uuid;not null;index" json:"org_id"`
	WorkspaceID     string `gorm:"type:uuid;not null;index" json:"workspace_id"`
	Priority        int    `gorm:"not null;default:0" json:"priority"`
	Status          string `gorm:"not null;default:'waiting';index" json:"status"`
	AssignedAgentID string `gorm:"type:uuid" json:"assigned_agent_id,omitempty"`

	AssignedAt   *time.Time `json:"assigned_at,omitempty"`
	EnteredAt    time.Time  `gorm:"not null;default:now()" json:"entered_at"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
	EndedReason  string     `json:"ended_reason,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (Entry) TableName() string { return "call_queue_entries" }

func (e *Entry) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	return nil
}
