package apps

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WebhookStore handles persistence for Webhook records.
type WebhookStore struct {
	db *gorm.DB
}

// NewWebhookStore constructs a WebhookStore backed by the given DB.
func NewWebhookStore(db *gorm.DB) *WebhookStore {
	return &WebhookStore{db: db}
}

// Create persists a new Webhook. The caller is responsible for setting all fields.
func (s *WebhookStore) Create(w *Webhook) (*Webhook, error) {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	if err := s.db.Create(w).Error; err != nil {
		return nil, err
	}
	return w, nil
}

// FindByID returns a webhook by its primary key.
func (s *WebhookStore) FindByID(id string) (*Webhook, error) {
	var w Webhook
	if err := s.db.First(&w, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

// FindByAppID returns all webhooks owned by an app.
func (s *WebhookStore) FindByAppID(appID string) ([]*Webhook, error) {
	var ws []*Webhook
	if err := s.db.Where("app_id = ?", appID).Order("created_at DESC").Find(&ws).Error; err != nil {
		return nil, err
	}
	return ws, nil
}

// FindByOrg returns all webhooks scoped to an organisation, optionally filtered
// to a specific workspace when workspaceID is non-nil.
func (s *WebhookStore) FindByOrg(orgID string, workspaceID *string) ([]*Webhook, error) {
	var ws []*Webhook
	q := s.db.Where("org_id = ?", orgID)
	if workspaceID != nil && *workspaceID != "" {
		q = q.Where("workspace_id = ?", *workspaceID)
	}
	if err := q.Order("created_at DESC").Find(&ws).Error; err != nil {
		return nil, err
	}
	return ws, nil
}

// Update saves mutable fields (URL, Events, IsActive) on an existing webhook.
func (s *WebhookStore) Update(w *Webhook) (*Webhook, error) {
	if err := s.db.Model(w).
		Where("id = ?", w.ID).
		Updates(map[string]interface{}{
			"url":       w.URL,
			"events":    w.Events,
			"is_active": w.IsActive,
		}).Error; err != nil {
		return nil, err
	}
	// Re-fetch so UpdatedAt reflects the DB value.
	return s.FindByID(w.ID.String())
}

// Delete hard-deletes a webhook by ID.
func (s *WebhookStore) Delete(id string) error {
	return s.db.Where("id = ?", id).Delete(&Webhook{}).Error
}
