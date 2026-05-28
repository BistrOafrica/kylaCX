package crm

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CRMStore is the database layer for CRM pipelines and stages.
type CRMStore struct {
	db *gorm.DB
}

// NewCRMStore constructs a CRMStore.
func NewCRMStore(db *gorm.DB) *CRMStore {
	return &CRMStore{db: db}
}

// ── Pipelines ─────────────────────────────────────────────────────────────────

func (s *CRMStore) CreatePipeline(p *Pipeline) (*Pipeline, error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	if err := s.db.Create(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

func (s *CRMStore) FindPipelineByID(id, orgID string) (*Pipeline, error) {
	var p Pipeline
	if err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *CRMStore) ListPipelines(orgID, workspaceID string) ([]*Pipeline, error) {
	q := s.db.Where("org_id = ?", orgID)
	if workspaceID != "" {
		q = q.Where("workspace_id = ?", workspaceID)
	}
	var pipelines []*Pipeline
	if err := q.Order("created_at ASC").Find(&pipelines).Error; err != nil {
		return nil, err
	}
	return pipelines, nil
}

func (s *CRMStore) UpdatePipeline(id, orgID string, updates map[string]interface{}) (*Pipeline, error) {
	updates["updated_at"] = time.Now()
	if err := s.db.Model(&Pipeline{}).Where("id = ? AND org_id = ?", id, orgID).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.FindPipelineByID(id, orgID)
}

func (s *CRMStore) DeletePipeline(id, orgID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("pipeline_id = ? AND org_id = ?", id, orgID).
			Delete(&PipelineStage{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND org_id = ?", id, orgID).Delete(&Pipeline{}).Error
	})
}

// PipelineStageCount returns how many stages belong to a pipeline.
func (s *CRMStore) PipelineStageCount(pipelineID, orgID string) (int, error) {
	var count int64
	err := s.db.Model(&PipelineStage{}).Where("pipeline_id = ? AND org_id = ?", pipelineID, orgID).Count(&count).Error
	return int(count), err
}

// ── Stages ────────────────────────────────────────────────────────────────────

func (s *CRMStore) CreateStage(st *PipelineStage) (*PipelineStage, error) {
	if st.ID == "" {
		st.ID = uuid.New().String()
	}
	// Assign next index if not set
	if st.Index == 0 {
		var maxIdx int
		s.db.Raw(`SELECT COALESCE(MAX("index"), -1) FROM crm_pipeline_stages WHERE pipeline_id = ? AND org_id = ?`,
			st.PipelineID, st.OrgID).Scan(&maxIdx)
		st.Index = maxIdx + 1
	}
	now := time.Now()
	st.CreatedAt = now
	st.UpdatedAt = now
	if err := s.db.Create(st).Error; err != nil {
		return nil, err
	}
	return st, nil
}

func (s *CRMStore) FindStageByID(id, orgID string) (*PipelineStage, error) {
	var st PipelineStage
	if err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&st).Error; err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *CRMStore) ListStages(pipelineID, orgID string) ([]*PipelineStage, error) {
	var stages []*PipelineStage
	if err := s.db.Where("pipeline_id = ? AND org_id = ?", pipelineID, orgID).
		Order(`"index" ASC`).Find(&stages).Error; err != nil {
		return nil, err
	}
	return stages, nil
}

func (s *CRMStore) UpdateStage(id, orgID string, updates map[string]interface{}) (*PipelineStage, error) {
	updates["updated_at"] = time.Now()
	if err := s.db.Model(&PipelineStage{}).Where("id = ? AND org_id = ?", id, orgID).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.FindStageByID(id, orgID)
}

func (s *CRMStore) DeleteStage(id, orgID string) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&PipelineStage{}).Error
}

// ReorderStageIndices applies the new ordering (stageIDs is the desired order).
func (s *CRMStore) ReorderStageIndices(pipelineID, orgID string, stageIDs []string) error {
	now := time.Now()
	return s.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range stageIDs {
			if err := tx.Model(&PipelineStage{}).
				Where("id = ? AND pipeline_id = ? AND org_id = ?", id, pipelineID, orgID).
				Updates(map[string]interface{}{"index": i, "updated_at": now}).Error; err != nil {
				return fmt.Errorf("reorder stage %s: %w", id, err)
			}
		}
		return nil
	})
}

// ── Deal operations ───────────────────────────────────────────────────────────

// DealRow is a minimal projection of an Object record for board display.
type DealRow struct {
	ID   string
	Data []byte
}

// ListDealsByStage returns deal objects whose JSONB data contains stage_id = stageID.
func (s *CRMStore) ListDealsByStage(orgID, workspaceID, stageID string, limit int) ([]DealRow, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id, data FROM objects
		WHERE org_id = ? AND type_slug = 'deal' AND data->>'stage_id' = ?`
	args := []interface{}{orgID, stageID}
	if workspaceID != "" {
		query += ` AND workspace_id = ?`
		args = append(args, workspaceID)
	}
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT %d`, limit)

	rows, err := s.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deals []DealRow
	for rows.Next() {
		var d DealRow
		if err := rows.Scan(&d.ID, &d.Data); err != nil {
			return nil, err
		}
		deals = append(deals, d)
	}
	return deals, nil
}

// PatchDealStage updates the stage_id in a deal object's JSONB data.
func (s *CRMStore) PatchDealStage(dealObjectID, orgID, stageID string) error {
	var current struct{ Data []byte }
	if err := s.db.Raw("SELECT data FROM objects WHERE id = ? AND org_id = ? AND type_slug = 'deal'",
		dealObjectID, orgID).Scan(&current).Error; err != nil {
		return err
	}
	if current.Data == nil {
		return fmt.Errorf("deal object %s not found", dealObjectID)
	}

	data := make(map[string]interface{})
	_ = json.Unmarshal(current.Data, &data)
	data["stage_id"] = stageID
	merged, _ := json.Marshal(data)

	return s.db.Exec(
		"UPDATE objects SET data = ?::jsonb, updated_at = ? WHERE id = ? AND org_id = ?",
		string(merged), time.Now(), dealObjectID, orgID,
	).Error
}
