package projects

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupProjectsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE projects (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			title TEXT NOT NULL,
			status TEXT NOT NULL,
			description TEXT,
			visibility TEXT NOT NULL,
			archived_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return db
}

func TestProjectStoreCRUD(t *testing.T) {
	db := setupProjectsTestDB(t)
	store := NewStore(db)

	created, err := store.Create(&Project{
		OrgID:       "org1",
		Title:       "Roadmap",
		Description: "Phase 4 completion",
		Visibility:  "private",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	read, err := store.Read("org1", created.ID)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if read.Title != "Roadmap" {
		t.Fatalf("expected title Roadmap, got %s", read.Title)
	}

	updated, err := store.Update("org1", &Project{ID: created.ID, Status: "active", Title: "Roadmap v2"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Title != "Roadmap v2" {
		t.Fatalf("expected updated title")
	}

	if err := store.Archive("org1", created.ID); err != nil {
		t.Fatalf("archive: %v", err)
	}
	archived, err := store.Read("org1", created.ID)
	if err != nil {
		t.Fatalf("read archived: %v", err)
	}
	if archived.Status != "archived" {
		t.Fatalf("expected archived status, got %s", archived.Status)
	}

	list, err := store.List("org1", 1, 20)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 project, got %d", len(list))
	}

	if err := store.Delete("org1", created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := store.Read("org1", created.ID); err == nil {
		t.Fatalf("expected not found after delete")
	}
}
