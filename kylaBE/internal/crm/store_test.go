package crm

import (
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCRMTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE crm_pipelines (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			workspace_id TEXT,
			name TEXT NOT NULL,
			description TEXT,
			type TEXT NOT NULL,
			color TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error; err != nil {
		t.Fatalf("create crm_pipelines schema: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE crm_pipeline_stages (
			id TEXT PRIMARY KEY,
			pipeline_id TEXT NOT NULL,
			org_id TEXT NOT NULL,
			name TEXT NOT NULL,
			color TEXT,
			"index" INTEGER NOT NULL,
			probability INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error; err != nil {
		t.Fatalf("create crm_pipeline_stages schema: %v", err)
	}

	return db
}

func TestCRMStoreMutationSmoke(t *testing.T) {
	db := setupCRMTestDB(t)
	store := NewCRMStore(db)

	pipeline, err := store.CreatePipeline(&Pipeline{
		OrgID:       "org1",
		WorkspaceID: "ws1",
		Name:        "Sales Pipeline",
		Description: "initial",
		Type:        "sales",
	})
	if err != nil {
		t.Fatalf("create pipeline: %v", err)
	}

	updatedPipeline, err := store.UpdatePipeline(pipeline.ID, "org1", map[string]interface{}{
		"name": "Sales Pipeline V2",
	})
	if err != nil {
		t.Fatalf("update pipeline: %v", err)
	}
	if updatedPipeline.Name != "Sales Pipeline V2" {
		t.Fatalf("expected updated pipeline name, got %q", updatedPipeline.Name)
	}

	stageA, err := store.CreateStage(&PipelineStage{
		PipelineID: pipeline.ID,
		OrgID:      "org1",
		Name:       "New",
	})
	if err != nil {
		t.Fatalf("create stage A: %v", err)
	}
	stageB, err := store.CreateStage(&PipelineStage{
		PipelineID: pipeline.ID,
		OrgID:      "org1",
		Name:       "Qualified",
	})
	if err != nil {
		t.Fatalf("create stage B: %v", err)
	}

	updatedStage, err := store.UpdateStage(stageA.ID, "org1", map[string]interface{}{
		"name":        "New Lead",
		"probability": 10,
	})
	if err != nil {
		t.Fatalf("update stage: %v", err)
	}
	if updatedStage.Name != "New Lead" || updatedStage.Probability != 10 {
		t.Fatalf("unexpected updated stage: %#v", updatedStage)
	}

	if err := store.ReorderStageIndices(pipeline.ID, "org1", []string{stageB.ID, stageA.ID}); err != nil {
		t.Fatalf("reorder stages: %v", err)
	}

	stages, err := store.ListStages(pipeline.ID, "org1")
	if err != nil {
		t.Fatalf("list stages: %v", err)
	}
	if len(stages) != 2 {
		t.Fatalf("expected 2 stages, got %d", len(stages))
	}
	if stages[0].ID != stageB.ID || stages[1].ID != stageA.ID {
		t.Fatalf("unexpected stage order after reorder: %s, %s", stages[0].ID, stages[1].ID)
	}

	if err := store.DeleteStage(stageA.ID, "org1"); err != nil {
		t.Fatalf("delete stage: %v", err)
	}
	if _, err := store.FindStageByID(stageA.ID, "org1"); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected stage not found after delete, got: %v", err)
	}

	if err := store.DeletePipeline(pipeline.ID, "org1"); err != nil {
		t.Fatalf("delete pipeline: %v", err)
	}
	if _, err := store.FindPipelineByID(pipeline.ID, "org1"); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected pipeline not found after delete, got: %v", err)
	}
}
