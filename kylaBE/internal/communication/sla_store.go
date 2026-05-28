package communication

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SLAPolicy defines service-level agreement targets.
type SLAPolicy struct {
	ID          string          `gorm:"type:uuid;primaryKey" json:"id"`
	OrgID       string          `gorm:"type:uuid;not null;index" json:"org_id"`
	WorkspaceID string          `gorm:"type:uuid;not null;index" json:"workspace_id"`
	Name        string          `gorm:"not null" json:"name"`
	Conditions  json.RawMessage `gorm:"type:jsonb" json:"conditions"` // [{field,op,value}]
	Metrics     json.RawMessage `gorm:"type:jsonb" json:"metrics"`    // {first_response_hours:1, resolution_hours:8}
	IsDefault   bool            `gorm:"default:false" json:"is_default"`
	IsActive    bool            `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// BeforeCreate hook.
func (p *SLAPolicy) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	if p.Conditions == nil {
		p.Conditions = []byte("[]")
	}
	if p.Metrics == nil {
		p.Metrics = []byte("{}")
	}
	return nil
}

// SLARecord tracks SLA compliance for a conversation.
type SLARecord struct {
	ID                    string     `gorm:"type:uuid;primaryKey" json:"id"`
	ConversationID        string     `gorm:"type:uuid;not null;uniqueIndex" json:"conversation_id"`
	PolicyID              string     `gorm:"type:uuid;not null" json:"policy_id"`
	OrgID                 string     `gorm:"type:uuid;not null;index" json:"org_id"`
	StartedAt             time.Time  `json:"started_at"`
	FirstResponseDeadline *time.Time `json:"first_response_deadline,omitempty"`
	FirstRespondedAt      *time.Time `json:"first_responded_at,omitempty"`
	ResolutionDeadline    *time.Time `json:"resolution_deadline,omitempty"`
	ResolvedAt            *time.Time `json:"resolved_at,omitempty"`
	FirstResponseBreached bool       `gorm:"default:false" json:"first_response_breached"`
	ResolutionBreached    bool       `gorm:"default:false" json:"resolution_breached"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// BeforeCreate hook.
func (r *SLARecord) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

// SLAStore manages SLA policies and records.
type SLAStore struct {
	db *gorm.DB
}

// NewSLAStore constructs an SLAStore.
func NewSLAStore(db *gorm.DB) *SLAStore {
	return &SLAStore{db: db}
}

// ── Policies ──

// CreatePolicy creates a new SLA policy.
func (s *SLAStore) CreatePolicy(policy *SLAPolicy) (*SLAPolicy, error) {
	if err := s.db.Create(policy).Error; err != nil {
		return nil, err
	}
	return policy, nil
}

// FindPolicyByID retrieves a policy by ID.
func (s *SLAStore) FindPolicyByID(id, orgID string) (*SLAPolicy, error) {
	var policy SLAPolicy
	err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&policy).Error
	return &policy, err
}

// FindActivePolicies retrieves all active policies for an org/workspace.
func (s *SLAStore) FindActivePolicies(orgID, workspaceID string) ([]*SLAPolicy, error) {
	var policies []*SLAPolicy
	err := s.db.Where("org_id = ? AND workspace_id = ? AND is_active = ?", orgID, workspaceID, true).
		Find(&policies).Error
	return policies, err
}

// UpdatePolicy updates a policy.
func (s *SLAStore) UpdatePolicy(policy *SLAPolicy) error {
	return s.db.Save(policy).Error
}

// DeletePolicy deletes a policy.
func (s *SLAStore) DeletePolicy(id, orgID string) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&SLAPolicy{}).Error
}

// ── Records ──

// CreateRecord creates an SLA record.
func (s *SLAStore) CreateRecord(record *SLARecord) (*SLARecord, error) {
	if err := s.db.Create(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}

// FindRecordByConversationID retrieves an SLA record by conversation ID.
func (s *SLAStore) FindRecordByConversationID(conversationID string) (*SLARecord, error) {
	var record SLARecord
	err := s.db.Where("conversation_id = ?", conversationID).First(&record).Error
	return &record, err
}

// UpdateRecord updates an SLA record.
func (s *SLAStore) UpdateRecord(record *SLARecord) error {
	return s.db.Save(record).Error
}

// FindBreachingRecords finds records that have breached but not yet been flagged.
func (s *SLAStore) FindBreachingRecords() ([]*SLARecord, error) {
	var records []*SLARecord
	now := time.Now()

	err := s.db.Where(
		"(first_response_deadline < ? AND first_responded_at IS NULL AND first_response_breached = false) OR "+
			"(resolution_deadline < ? AND resolved_at IS NULL AND resolution_breached = false)",
		now, now,
	).Find(&records).Error

	return records, err
}
