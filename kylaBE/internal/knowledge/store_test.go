package knowledge

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupKnowledgeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE kb_categories (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			workspace_id TEXT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL,
			icon TEXT,
			parent_id TEXT,
			position INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error; err != nil {
		t.Fatalf("create kb_categories schema: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE kb_articles (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			workspace_id TEXT,
			category_id TEXT,
			title TEXT NOT NULL,
			slug TEXT NOT NULL,
			content TEXT,
			excerpt TEXT,
			status TEXT NOT NULL,
			author_id TEXT,
			published_at DATETIME,
			view_count INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error; err != nil {
		t.Fatalf("create kb_articles schema: %v", err)
	}
	return db
}

func TestListArticlesCursorPagination(t *testing.T) {
	db := setupKnowledgeTestDB(t)
	store := NewKnowledgeStore(db)

	base := time.Now().UTC().Add(-time.Hour)
	rows := []*KBArticle{
		{ID: "a1", OrgID: "org1", WorkspaceID: "ws1", Title: "One", Slug: "one", Status: "published", CreatedAt: base.Add(1 * time.Minute), UpdatedAt: base.Add(1 * time.Minute)},
		{ID: "a2", OrgID: "org1", WorkspaceID: "ws1", Title: "Two", Slug: "two", Status: "published", CreatedAt: base.Add(2 * time.Minute), UpdatedAt: base.Add(2 * time.Minute)},
		{ID: "a3", OrgID: "org1", WorkspaceID: "ws1", Title: "Three", Slug: "three", Status: "published", CreatedAt: base.Add(3 * time.Minute), UpdatedAt: base.Add(3 * time.Minute)},
	}
	for _, r := range rows {
		if err := db.Create(r).Error; err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	page1, token, _, err := store.ListArticles(ListArticlesParams{
		OrgID:       "org1",
		WorkspaceID: "ws1",
		PageSize:    2,
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

	page2, token2, _, err := store.ListArticles(ListArticlesParams{
		OrgID:       "org1",
		WorkspaceID: "ws1",
		PageSize:    2,
		PageToken:   token,
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
