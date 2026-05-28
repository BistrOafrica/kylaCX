package audit

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Store handles database persistence for AuditLog records.
type Store struct {
	db *gorm.DB
}

// NewStore returns an audit Store backed by db.
func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

// Create persists a single AuditLog record.
func (s *Store) Create(entry *AuditLog) error {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	return s.db.Create(entry).Error
}

// FindByUser returns the most recent limit audit entries for userID.
func (s *Store) FindByUser(userID string, limit int) ([]*AuditLog, error) {
	var entries []*AuditLog
	err := s.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&entries).Error
	return entries, err
}

// FindByOrg returns the most recent limit audit entries for orgID.
func (s *Store) FindByOrg(orgID string, limit int) ([]*AuditLog, error) {
	var entries []*AuditLog
	err := s.db.
		Where("org_id = ?", orgID).
		Order("created_at DESC").
		Limit(limit).
		Find(&entries).Error
	return entries, err
}
