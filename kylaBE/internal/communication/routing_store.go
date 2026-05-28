package communication

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoutingRule defines an automatic assignment rule.
type RoutingRule struct {
	ID          string          `gorm:"type:uuid;primaryKey" json:"id"`
	OrgID       string          `gorm:"type:uuid;not null;index" json:"org_id"`
	WorkspaceID string          `gorm:"type:uuid;not null;index" json:"workspace_id"`
	Name        string          `gorm:"not null" json:"name"`
	Priority    int             `gorm:"not null;default:0;index" json:"priority"`       // higher = evaluated first
	Conditions  json.RawMessage `gorm:"type:jsonb" json:"conditions"`                   // [{field,op,value}]
	Actions     json.RawMessage `gorm:"type:jsonb" json:"actions"`                      // [{type,target_id}]
	Strategy    string          `gorm:"not null;default:'round_robin'" json:"strategy"` // round_robin | skill_based | direct
	IsActive    bool            `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// BeforeCreate hook.
func (r *RoutingRule) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.Conditions == nil {
		r.Conditions = []byte("[]")
	}
	if r.Actions == nil {
		r.Actions = []byte("[]")
	}
	return nil
}

// RoutingStore manages routing rules.
type RoutingStore struct {
	db *gorm.DB
}

// NewRoutingStore constructs a RoutingStore.
func NewRoutingStore(db *gorm.DB) *RoutingStore {
	return &RoutingStore{db: db}
}

// Create creates a new routing rule.
func (s *RoutingStore) Create(rule *RoutingRule) (*RoutingRule, error) {
	if err := s.db.Create(rule).Error; err != nil {
		return nil, err
	}
	return rule, nil
}

// FindByID retrieves a routing rule by ID.
func (s *RoutingStore) FindByID(id, orgID string) (*RoutingRule, error) {
	var rule RoutingRule
	err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&rule).Error
	return &rule, err
}

// FindActiveRules retrieves all active rules for an org/workspace, ordered by priority DESC.
func (s *RoutingStore) FindActiveRules(orgID, workspaceID string) ([]*RoutingRule, error) {
	var rules []*RoutingRule
	err := s.db.Where("org_id = ? AND workspace_id = ? AND is_active = ?", orgID, workspaceID, true).
		Order("priority DESC, created_at ASC").
		Find(&rules).Error
	return rules, err
}

// Update updates a routing rule.
func (s *RoutingStore) Update(rule *RoutingRule) error {
	return s.db.Save(rule).Error
}

// Delete deletes a routing rule.
func (s *RoutingStore) Delete(id, orgID string) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&RoutingRule{}).Error
}
