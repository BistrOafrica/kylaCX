// Package ivr implements the Phase 5c IVR (Interactive Voice Response) engine.
// Lives under internal/telephony/ to share the telephony package's PBXController
// abstraction without creating a circular dependency between packages.
package ivr

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FlowStatus enumerates flow lifecycle states.
type RunStatus string

const (
	RunStatusRunning   RunStatus = "running"
	RunStatusCompleted RunStatus = "completed"
	RunStatusFailed    RunStatus = "failed"
	RunStatusAbandoned RunStatus = "abandoned"
)

// NodeType enumerates the built-in IVR node types.
// Mirrors the Phase 6 automation action types in spirit — each node has a
// fixed type and a JSON config the executor reads.
type NodeType string

const (
	NodePlayAudio NodeType = "play_audio" // play a static audio file
	NodeSay       NodeType = "say"        // TTS the supplied text via mod_say
	NodeMenu      NodeType = "menu"       // play_and_get_digits; branches on the captured digit
	NodeTransfer  NodeType = "transfer"   // hand the call to an extension, queue, or external number
	NodeRecord    NodeType = "record"     // record_session to disk (with optional max duration)
	NodeHangup    NodeType = "hangup"     // terminate the call
	NodeGoto      NodeType = "goto"       // unconditional jump (rarely used; useful for loops)
)

// Flow is the persisted IVR definition.
//
// Definition holds the node graph as JSONB so the visual builder can evolve
// without DB migrations.
type Flow struct {
	ID          string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID       string `gorm:"type:uuid;not null;index" json:"org_id"`
	WorkspaceID string `gorm:"type:uuid;not null;index" json:"workspace_id"`
	Name        string `gorm:"not null" json:"name"`
	Description string `gorm:"not null;default:''" json:"description,omitempty"`

	Definition json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"definition"`

	IsActive bool `gorm:"not null;default:false;index" json:"is_active"`
	Version  int  `gorm:"not null;default:1" json:"version"`

	CreatedBy string    `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Flow) TableName() string { return "ivr_flows" }

func (f *Flow) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.NewString()
	}
	return nil
}

// Definition is the deserialised shape of Flow.Definition. Decoded once at
// the start of an IVR run; the executor walks the resulting in-memory graph.
type Definition struct {
	StartNodeID string `json:"start_node_id"`
	Nodes       []Node `json:"nodes"`
}

// Node is one step in the IVR graph. Config is intentionally a free-form map
// so adding new node types doesn't require new struct fields.
type Node struct {
	ID         string                 `json:"id"`
	Type       NodeType               `json:"type"`
	Config     map[string]interface{} `json:"config"`
	NextNodeID string                 `json:"next_node_id,omitempty"`
	Branches   map[string]string      `json:"branches,omitempty"`
}

// FindNode returns the node with the given ID, or (nil, false) if not present.
func (d Definition) FindNode(id string) (*Node, bool) {
	for i := range d.Nodes {
		if d.Nodes[i].ID == id {
			return &d.Nodes[i], true
		}
	}
	return nil, false
}

// DIDMapping routes an inbound DID to an IVR flow. The executor consults
// this table at CHANNEL_CREATE time.
type DIDMapping struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID       string    `gorm:"type:uuid;not null;index" json:"org_id"`
	WorkspaceID string    `gorm:"type:uuid;not null;index" json:"workspace_id"`
	DID         string    `gorm:"not null" json:"did"`
	FlowID      string    `gorm:"type:uuid;not null" json:"flow_id"`
	CreatedAt   time.Time `json:"created_at"`
}

func (DIDMapping) TableName() string { return "ivr_did_mappings" }

func (m *DIDMapping) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return nil
}

// Run is one IVR execution attached to a specific call. The visited_nodes
// JSONB is a breadcrumb trail the analytics layer reads.
type Run struct {
	ID            string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FlowID        string `gorm:"type:uuid;not null;index" json:"flow_id"`
	CallID        string `gorm:"type:uuid;not null;index" json:"call_id"`
	OrgID         string `gorm:"type:uuid;not null;index" json:"org_id"`
	WorkspaceID   string `gorm:"type:uuid;not null;index" json:"workspace_id"`
	Status        string `gorm:"not null;default:'running';index" json:"status"`
	CurrentNodeID string `gorm:"not null;default:''" json:"current_node_id"`

	VisitedNodes json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"visited_nodes"`

	StartedAt time.Time  `gorm:"not null;default:now()" json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	EndReason string     `json:"end_reason,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Run) TableName() string { return "ivr_runs" }

func (r *Run) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}
	return nil
}

// RunStep is one entry in Run.VisitedNodes.
type RunStep struct {
	NodeID    string    `json:"node_id"`
	EnteredAt time.Time `json:"entered_at"`
	Input     string    `json:"input,omitempty"`
}
