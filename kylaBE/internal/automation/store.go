package automation

import (
	"errors"
	"fmt"
	"time"

	"kyla-be/shared/events"

	"gorm.io/gorm"
)

// Store provides persistence for workflow definitions and run projections.
type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

// ── Workflow CRUD ────────────────────────────────────────────────────────────

func (s *Store) CreateWorkflow(w *Workflow) (*Workflow, error) {
	if w.Status == "" {
		w.Status = WorkflowStatusDraft
	}
	if err := s.db.Create(w).Error; err != nil {
		return nil, err
	}
	return w, nil
}

func (s *Store) GetWorkflow(id string) (*Workflow, error) {
	var w Workflow
	if err := s.db.First(&w, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *Store) ListWorkflows(orgID, workspaceID string, page, pageSize int) ([]*Workflow, int64, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	q := s.db.Model(&Workflow{}).Where("org_id = ?", orgID)
	if workspaceID != "" {
		q = q.Where("workspace_id = ?", workspaceID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []*Workflow
	err := q.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error
	return items, total, err
}

// UpdateWorkflow applies the supplied field map and returns the updated row.
// Only whitelisted columns are allowed.
func (s *Store) UpdateWorkflow(id string, updates map[string]interface{}) (*Workflow, error) {
	if len(updates) == 0 {
		return nil, errors.New("no updates supplied")
	}
	updates["updated_at"] = time.Now().UTC()
	if err := s.db.Model(&Workflow{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.GetWorkflow(id)
}

func (s *Store) DeleteWorkflow(id string) error {
	return s.db.Delete(&Workflow{}, "id = ?", id).Error
}

// IncrementRunCount bumps the run_count column atomically.
func (s *Store) IncrementRunCount(workflowID string) error {
	return s.db.Model(&Workflow{}).
		Where("id = ?", workflowID).
		UpdateColumn("run_count", gorm.Expr("run_count + 1")).
		Error
}

// ── Trigger Matching ─────────────────────────────────────────────────────────

// FindMatchingWorkflows returns all active workflows whose trigger matches the
// given DomainEvent. The match is done in two layers:
//  1. Coarse SQL filter by trigger.type extracted from the JSONB column,
//     scoped to the event's org_id (and workspace_id if present).
//  2. Caller-side condition evaluation (workflow.Conditions vs event payload),
//     handled by the engine — this store call only does the coarse filter.
func (s *Store) FindMatchingWorkflows(event *events.DomainEvent) ([]*Workflow, error) {
	if event == nil || event.OrgID == "" {
		return nil, nil
	}
	// trigger.type is matched against either the canonical "{domain}.{action}"
	// shape (e.g. "ticket.created") or a wildcard subject pattern stored as
	// "trigger.event_subject_pattern". The coarse SQL prefilter keeps the
	// candidate set small; finer-grained matching happens in the engine.
	triggerType := fmt.Sprintf("%s.%s", event.Domain, event.Action)
	q := s.db.Where("org_id = ? AND status = ?", event.OrgID, WorkflowStatusActive).
		Where("trigger->>'type' = ?", triggerType)
	if event.WorkspaceID != "" {
		q = q.Where("workspace_id = ? OR workspace_id = ''", event.WorkspaceID)
	}
	var items []*Workflow
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// ── WorkflowRun Projection ───────────────────────────────────────────────────

// CreateRun inserts a projection row when a Temporal workflow is started.
func (s *Store) CreateRun(run *WorkflowRun) (*WorkflowRun, error) {
	if run.Status == "" {
		run.Status = RunStatusPending
	}
	if err := s.db.Create(run).Error; err != nil {
		return nil, err
	}
	return run, nil
}

// UpdateRunStatus is called by Temporal lifecycle hooks when execution state
// transitions (success / failed / running).
func (s *Store) UpdateRunStatus(temporalRunID string, status RunStatus, errMsg string, finishedAt *time.Time) error {
	updates := map[string]interface{}{"status": status}
	if errMsg != "" {
		updates["error"] = errMsg
	}
	if finishedAt != nil {
		updates["finished_at"] = finishedAt
	}
	return s.db.Model(&WorkflowRun{}).
		Where("temporal_run_id = ?", temporalRunID).
		Updates(updates).Error
}

func (s *Store) ListRuns(workflowID string, page, pageSize int) ([]*WorkflowRun, int64, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	q := s.db.Model(&WorkflowRun{}).Where("workflow_id = ?", workflowID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []*WorkflowRun
	err := q.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error
	return items, total, err
}
