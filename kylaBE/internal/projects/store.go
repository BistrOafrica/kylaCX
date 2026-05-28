package projects

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Store handles persistence for projects.
type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(p *Project) (*Project, error) {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now
	if p.Status == "" {
		p.Status = "active"
	}
	if p.Visibility == "" {
		p.Visibility = "private"
	}
	if err := s.db.Create(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) Read(orgID, id string) (*Project, error) {
	var p Project
	if err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) Update(orgID string, in *Project) (*Project, error) {
	updates := map[string]interface{}{
		"updated_at": time.Now().UTC(),
	}
	if in.Title != "" {
		updates["title"] = in.Title
	}
	if in.Status != "" {
		updates["status"] = in.Status
	}
	if in.Description != "" {
		updates["description"] = in.Description
	}
	if in.Visibility != "" {
		updates["visibility"] = in.Visibility
	}
	if err := s.db.Model(&Project{}).Where("id = ? AND org_id = ?", in.ID, orgID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.Read(orgID, in.ID)
}

func (s *Store) Delete(orgID, id string) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&Project{}).Error
}

func (s *Store) List(orgID string, page, perPage int64) ([]*Project, error) {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 || perPage > 200 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var out []*Project
	if err := s.db.Where("org_id = ?", orgID).
		Order("created_at DESC").
		Offset(int(offset)).
		Limit(int(perPage)).
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) Archive(orgID, id string) error {
	now := time.Now().UTC()
	return s.db.Model(&Project{}).
		Where("id = ? AND org_id = ?", id, orgID).
		Updates(map[string]interface{}{
			"status":      "archived",
			"archived_at": &now,
			"updated_at":  now,
		}).Error
}
