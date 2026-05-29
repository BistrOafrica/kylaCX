package ivr

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Store wraps DB access for IVR flows, DID mappings, and run history.
type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

// ── Flows ────────────────────────────────────────────────────────────────────

func (s *Store) CreateFlow(f *Flow) (*Flow, error) {
	if err := s.db.Create(f).Error; err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Store) GetFlow(id, orgID string) (*Flow, error) {
	var f Flow
	err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&f).Error
	return &f, err
}

// GetFlowByIDOnly skips org scoping — used by the executor on inbound calls
// where the trigger path (DID lookup) already proved org ownership.
func (s *Store) GetFlowByIDOnly(id string) (*Flow, error) {
	var f Flow
	err := s.db.Where("id = ?", id).First(&f).Error
	return &f, err
}

func (s *Store) ListFlows(workspaceID string, activeOnly bool) ([]*Flow, error) {
	q := s.db.Where("workspace_id = ?", workspaceID)
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	var out []*Flow
	err := q.Order("name ASC").Find(&out).Error
	return out, err
}

// UpdateFlow saves writable fields and bumps the version. Definition changes
// are common so the version increment is cheap insurance for migration logic
// later.
func (s *Store) UpdateFlow(f *Flow) (*Flow, error) {
	if err := s.db.Model(&Flow{}).Where("id = ? AND org_id = ?", f.ID, f.OrgID).Updates(map[string]interface{}{
		"name":        f.Name,
		"description": f.Description,
		"definition":  f.Definition,
		"is_active":   f.IsActive,
		"version":     gorm.Expr("version + 1"),
		"updated_at":  time.Now().UTC(),
	}).Error; err != nil {
		return nil, err
	}
	return s.GetFlow(f.ID, f.OrgID)
}

func (s *Store) DeleteFlow(id, orgID string) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&Flow{}).Error
}

// ── DID mappings ────────────────────────────────────────────────────────────

func (s *Store) CreateDIDMapping(m *DIDMapping) (*DIDMapping, error) {
	if err := s.db.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// FindFlowIDForDID resolves an inbound DID to the IVR flow that should handle
// it. Returns ("", gorm.ErrRecordNotFound) when no mapping exists — caller
// treats this as "no IVR, fall through to default inbound handling".
func (s *Store) FindFlowIDForDID(did string) (flowID, orgID, workspaceID string, err error) {
	var m DIDMapping
	if err := s.db.Where("did = ?", did).First(&m).Error; err != nil {
		return "", "", "", err
	}
	return m.FlowID, m.OrgID, m.WorkspaceID, nil
}

func (s *Store) ListDIDMappings(workspaceID string) ([]*DIDMapping, error) {
	var out []*DIDMapping
	err := s.db.Where("workspace_id = ?", workspaceID).Order("did ASC").Find(&out).Error
	return out, err
}

func (s *Store) DeleteDIDMapping(id, orgID string) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&DIDMapping{}).Error
}

// ── Runs ────────────────────────────────────────────────────────────────────

func (s *Store) CreateRun(r *Run) (*Run, error) {
	if r.VisitedNodes == nil {
		r.VisitedNodes = json.RawMessage(`[]`)
	}
	if err := s.db.Create(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Store) GetRun(id, orgID string) (*Run, error) {
	var r Run
	err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&r).Error
	return &r, err
}

// GetRunByCallID is the executor's lookup path during ESL event handling.
// One run per call is expected; if multiple exist (replay scenarios) the
// latest one wins.
func (s *Store) GetRunByCallID(callID string) (*Run, error) {
	var r Run
	err := s.db.Where("call_id = ?", callID).Order("started_at DESC").First(&r).Error
	return &r, err
}

type ListRunsParams struct {
	FlowID    string
	CallID    string
	PageSize  int
	PageToken string
}

func (s *Store) ListRuns(p ListRunsParams) ([]*Run, int64, error) {
	q := s.db.Model(&Run{})
	if p.FlowID != "" {
		q = q.Where("flow_id = ?", p.FlowID)
	}
	if p.CallID != "" {
		q = q.Where("call_id = ?", p.CallID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if p.PageSize <= 0 || p.PageSize > 200 {
		p.PageSize = 50
	}
	var out []*Run
	if err := q.Order("started_at DESC").Limit(p.PageSize).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// AdvanceRun moves the run to a new node, appending the previous node to the
// breadcrumb trail. input is the user-supplied data captured at the previous
// node (e.g. DTMF digit for a menu).
func (s *Store) AdvanceRun(runID, nextNodeID, input string) error {
	var r Run
	if err := s.db.Where("id = ?", runID).First(&r).Error; err != nil {
		return err
	}
	var steps []RunStep
	if len(r.VisitedNodes) > 0 {
		_ = json.Unmarshal(r.VisitedNodes, &steps)
	}
	steps = append(steps, RunStep{
		NodeID:    r.CurrentNodeID,
		EnteredAt: time.Now().UTC(),
		Input:     input,
	})
	stepsJSON, err := json.Marshal(steps)
	if err != nil {
		return err
	}
	return s.db.Model(&Run{}).Where("id = ?", runID).Updates(map[string]interface{}{
		"current_node_id": nextNodeID,
		"visited_nodes":   json.RawMessage(stepsJSON),
		"updated_at":      time.Now().UTC(),
	}).Error
}

// EndRun finalises the run with the supplied terminal status and reason.
func (s *Store) EndRun(runID string, status RunStatus, reason string) error {
	now := time.Now().UTC()
	return s.db.Model(&Run{}).Where("id = ?", runID).Updates(map[string]interface{}{
		"status":     string(status),
		"end_reason": reason,
		"ended_at":   &now,
		"updated_at": now,
	}).Error
}

// ── helpers ──────────────────────────────────────────────────────────────────

// DecodeDefinition parses Flow.Definition JSONB into the typed struct. Returns
// an error if the JSONB is malformed; an empty definition is not an error.
func DecodeDefinition(raw json.RawMessage) (Definition, error) {
	var d Definition
	if len(raw) == 0 {
		return d, nil
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		return d, err
	}
	if d.StartNodeID == "" {
		return d, errors.New("definition: start_node_id is required")
	}
	return d, nil
}
