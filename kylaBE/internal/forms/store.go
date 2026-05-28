package forms

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FormsStore is the database layer for forms and submissions.
type FormsStore struct {
	db *gorm.DB
}

// NewFormsStore constructs a FormsStore.
func NewFormsStore(db *gorm.DB) *FormsStore {
	return &FormsStore{db: db}
}

// ── Forms ─────────────────────────────────────────────────────────────────────

func (s *FormsStore) CreateForm(f *Form) (*Form, error) {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	if f.Fields == "" {
		f.Fields = "[]"
	}
	now := time.Now()
	f.CreatedAt = now
	f.UpdatedAt = now
	if err := s.db.Create(f).Error; err != nil {
		return nil, err
	}
	return f, nil
}

func (s *FormsStore) FindFormByID(id, orgID string) (*Form, error) {
	var f Form
	if err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (s *FormsStore) ListForms(orgID, workspaceID, filterStatus string) ([]*Form, error) {
	q := s.db.Where("org_id = ?", orgID)
	if workspaceID != "" {
		q = q.Where("workspace_id = ?", workspaceID)
	}
	if filterStatus != "" {
		q = q.Where("status = ?", filterStatus)
	}
	var forms []*Form
	if err := q.Order("created_at DESC").Find(&forms).Error; err != nil {
		return nil, err
	}
	return forms, nil
}

func (s *FormsStore) UpdateForm(id, orgID string, updates map[string]interface{}) (*Form, error) {
	updates["updated_at"] = time.Now()
	if err := s.db.Model(&Form{}).Where("id = ? AND org_id = ?", id, orgID).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.FindFormByID(id, orgID)
}

func (s *FormsStore) DeleteForm(id, orgID string) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&Form{}).Error
}

// ── Submissions ───────────────────────────────────────────────────────────────

func (s *FormsStore) CreateSubmission(sub *FormSubmission) (*FormSubmission, error) {
	if sub.ID == "" {
		sub.ID = uuid.New().String()
	}
	sub.CreatedAt = time.Now()
	return sub, s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(sub).Error; err != nil {
			return err
		}
		// Increment submission_count on the form
		return tx.Model(&Form{}).Where("id = ? AND org_id = ?", sub.FormID, sub.OrgID).
			UpdateColumn("submission_count", gorm.Expr("submission_count + 1")).Error
	})
}

// ListSubmissionsParams bundles paginated listing options for submissions.
type ListSubmissionsParams struct {
	OrgID     string
	FormID    string
	PageSize  int
	PageToken string
}

func (s *FormsStore) ListSubmissions(p ListSubmissionsParams) ([]*FormSubmission, string, int64, error) {
	if p.PageSize <= 0 || p.PageSize > 200 {
		p.PageSize = 50
	}
	q := s.db.Where("form_id = ? AND org_id = ?", p.FormID, p.OrgID)

	var total int64
	q.Model(&FormSubmission{}).Count(&total)

	if p.PageToken != "" {
		if ts, id, ok := decodeSubmissionCursor(p.PageToken); ok {
			q = q.Where("(created_at < ?) OR (created_at = ? AND id < ?)", ts, ts, id)
		} else {
			var pivot FormSubmission
			if err := s.db.Select("id", "created_at").Where("id = ? AND form_id = ? AND org_id = ?", p.PageToken, p.FormID, p.OrgID).First(&pivot).Error; err == nil {
				q = q.Where("(created_at < ?) OR (created_at = ? AND id < ?)", pivot.CreatedAt, pivot.CreatedAt, pivot.ID)
			}
		}
	}
	q = q.Order("created_at DESC, id DESC").Limit(p.PageSize + 1)

	var subs []*FormSubmission
	if err := q.Find(&subs).Error; err != nil {
		return nil, "", total, err
	}
	nextToken := ""
	if len(subs) > p.PageSize {
		nextToken = encodeSubmissionCursor(subs[p.PageSize-1].CreatedAt, subs[p.PageSize-1].ID)
		subs = subs[:p.PageSize]
	}
	return subs, nextToken, total, nil
}

// CreateObjectRecord creates an Object Core record for a form submission.
// Returns the created object ID or empty string on failure (non-fatal).
func (s *FormsStore) CreateObjectRecord(orgID, workspaceID, data string, actorID string) string {
	id := uuid.New().String()
	err := s.db.Exec(
		`INSERT INTO objects (id, org_id, workspace_id, type_slug, data, created_by, created_at, updated_at)
		 VALUES (?, ?, ?, 'form_submission', ?::jsonb, ?, now(), now())`,
		id, orgID, workspaceID, data, actorID,
	).Error
	if err != nil {
		fmt.Printf("[forms] create object record failed: %v\n", err)
		return ""
	}
	return id
}

func encodeSubmissionCursor(ts time.Time, id string) string {
	return ts.UTC().Format(time.RFC3339Nano) + "|" + id
}

func decodeSubmissionCursor(token string) (time.Time, string, bool) {
	parts := strings.SplitN(token, "|", 2)
	if len(parts) != 2 {
		return time.Time{}, "", false
	}
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, "", false
	}
	if parts[1] == "" {
		return time.Time{}, "", false
	}
	return ts, parts[1], true
}
