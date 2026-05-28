package forms

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupFormsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE forms (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			workspace_id TEXT,
			name TEXT NOT NULL,
			description TEXT,
			fields TEXT NOT NULL,
			status TEXT NOT NULL,
			submit_redirect TEXT,
			submission_count INTEGER,
			created_by TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error; err != nil {
		t.Fatalf("create forms schema: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE form_submissions (
			id TEXT PRIMARY KEY,
			form_id TEXT NOT NULL,
			org_id TEXT NOT NULL,
			data TEXT NOT NULL,
			object_id TEXT,
			created_at DATETIME
		)
	`).Error; err != nil {
		t.Fatalf("create form_submissions schema: %v", err)
	}
	return db
}

func TestListSubmissionsCursorPagination(t *testing.T) {
	db := setupFormsTestDB(t)
	store := NewFormsStore(db)

	base := time.Now().UTC().Add(-time.Hour)
	rows := []*FormSubmission{
		{ID: "s1", FormID: "f1", OrgID: "org1", Data: "{}", CreatedAt: base.Add(1 * time.Minute)},
		{ID: "s2", FormID: "f1", OrgID: "org1", Data: "{}", CreatedAt: base.Add(2 * time.Minute)},
		{ID: "s3", FormID: "f1", OrgID: "org1", Data: "{}", CreatedAt: base.Add(3 * time.Minute)},
	}
	for _, r := range rows {
		if err := db.Create(r).Error; err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	page1, token, _, err := store.ListSubmissions(ListSubmissionsParams{
		OrgID:    "org1",
		FormID:   "f1",
		PageSize: 2,
	})
	if err != nil {
		t.Fatalf("list page1: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 records, got %d", len(page1))
	}
	if token == "" {
		t.Fatalf("expected non-empty next token")
	}

	page2, token2, _, err := store.ListSubmissions(ListSubmissionsParams{
		OrgID:     "org1",
		FormID:    "f1",
		PageSize:  2,
		PageToken: token,
	})
	if err != nil {
		t.Fatalf("list page2: %v", err)
	}
	if len(page2) != 1 {
		t.Fatalf("expected 1 record on page2, got %d", len(page2))
	}
	if token2 != "" {
		t.Fatalf("expected empty token after last page")
	}
	if page1[0].ID == page2[0].ID || page1[1].ID == page2[0].ID {
		t.Fatalf("expected non-overlapping pages")
	}
}
