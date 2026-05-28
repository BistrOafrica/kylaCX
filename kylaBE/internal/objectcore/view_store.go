package objectcore

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ViewStore handles persistence of SavedView records.
type ViewStore struct {
	db *gorm.DB
}

// NewViewStore constructs a ViewStore backed by the given DB.
func NewViewStore(db *gorm.DB) *ViewStore {
	return &ViewStore{db: db}
}

// Create persists a new SavedView.
func (s *ViewStore) Create(v *SavedView) (*SavedView, error) {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	now := time.Now()
	v.CreatedAt = now
	v.UpdatedAt = now
	if err := s.db.Create(v).Error; err != nil {
		return nil, err
	}
	return v, nil
}

// FindByID returns a SavedView by primary key, scoped to a workspace.
func (s *ViewStore) FindByID(id, workspaceID string) (*SavedView, error) {
	var sv SavedView
	if err := s.db.Where("id = ? AND workspace_id = ?", id, workspaceID).First(&sv).Error; err != nil {
		return nil, err
	}
	return &sv, nil
}

// ListByWorkspace returns all views for a workspace, optionally filtered by type slug.
func (s *ViewStore) ListByWorkspace(workspaceID, typeSlug string) ([]*SavedView, error) {
	q := s.db.Where("workspace_id = ?", workspaceID)
	if typeSlug != "" {
		q = q.Where("type_slug = ?", typeSlug)
	}
	var views []*SavedView
	if err := q.Order("name ASC").Find(&views).Error; err != nil {
		return nil, err
	}
	return views, nil
}

// Update saves mutable fields on a SavedView.
func (s *ViewStore) Update(v *SavedView) (*SavedView, error) {
	v.UpdatedAt = time.Now()
	if err := s.db.Model(v).Where("id = ? AND workspace_id = ?", v.ID, v.WorkspaceID).
		Updates(map[string]interface{}{
			"name":       v.Name,
			"filters":    v.Filters,
			"sort":       v.Sort,
			"columns":    v.Columns,
			"is_shared":  v.IsShared,
			"updated_at": v.UpdatedAt,
		}).Error; err != nil {
		return nil, err
	}
	return s.FindByID(v.ID, v.WorkspaceID)
}

// Delete hard-deletes a SavedView.
func (s *ViewStore) Delete(id, workspaceID string) error {
	return s.db.Where("id = ? AND workspace_id = ?", id, workspaceID).Delete(&SavedView{}).Error
}
