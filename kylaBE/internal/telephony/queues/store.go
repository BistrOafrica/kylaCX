package queues

import (
	"time"

	"gorm.io/gorm"
)

// Store wraps DB access for queues, members, and entries.
type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

// ── Queues ───────────────────────────────────────────────────────────────────

func (s *Store) CreateQueue(q *Queue) (*Queue, error) {
	if err := s.db.Create(q).Error; err != nil {
		return nil, err
	}
	return q, nil
}

func (s *Store) GetQueue(id, orgID string) (*Queue, error) {
	var q Queue
	err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&q).Error
	return &q, err
}

// GetQueueByIDOnly skips org scoping — used by the routing engine which is
// triggered by trusted PBX events.
func (s *Store) GetQueueByIDOnly(id string) (*Queue, error) {
	var q Queue
	err := s.db.Where("id = ?", id).First(&q).Error
	return &q, err
}

func (s *Store) ListQueues(workspaceID string, activeOnly bool) ([]*Queue, error) {
	q := s.db.Where("workspace_id = ?", workspaceID)
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	var out []*Queue
	err := q.Order("name ASC").Find(&out).Error
	return out, err
}

func (s *Store) UpdateQueue(q *Queue) (*Queue, error) {
	if err := s.db.Model(&Queue{}).Where("id = ? AND org_id = ?", q.ID, q.OrgID).Updates(map[string]interface{}{
		"name":             q.Name,
		"description":      q.Description,
		"strategy":         q.Strategy,
		"moh_path":         q.MOHPath,
		"max_wait_seconds": q.MaxWaitSeconds,
		"overflow_action":  q.OverflowAction,
		"overflow_target":  q.OverflowTarget,
		"is_active":        q.IsActive,
		"updated_at":       time.Now().UTC(),
	}).Error; err != nil {
		return nil, err
	}
	return s.GetQueue(q.ID, q.OrgID)
}

func (s *Store) DeleteQueue(id, orgID string) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&Queue{}).Error
}

// ── Members ─────────────────────────────────────────────────────────────────

func (s *Store) AddMember(m *Membership) (*Membership, error) {
	if err := s.db.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Store) RemoveMember(id, orgID string) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&Membership{}).Error
}

func (s *Store) ListMembers(queueID string) ([]*Membership, error) {
	var out []*Membership
	err := s.db.Where("queue_id = ?", queueID).Order("priority DESC, user_id ASC").Find(&out).Error
	return out, err
}

// SetMemberActive toggles the member's pause/resume flag. Returns the updated row.
func (s *Store) SetMemberActive(queueID, userID string, isActive bool) (*Membership, error) {
	if err := s.db.Model(&Membership{}).
		Where("queue_id = ? AND user_id = ?", queueID, userID).
		Updates(map[string]interface{}{
			"is_active":  isActive,
			"updated_at": time.Now().UTC(),
		}).Error; err != nil {
		return nil, err
	}
	var m Membership
	err := s.db.Where("queue_id = ? AND user_id = ?", queueID, userID).First(&m).Error
	return &m, err
}

// FindEligibleMembers returns active members ordered by the queue strategy.
// For round_robin the caller chooses the next one in priority order; for
// longest_idle the rows are pre-sorted by last_call_ended_at ASC NULLS FIRST
// so the first row is the right pick.
func (s *Store) FindEligibleMembers(queueID string, strategy Strategy) ([]*Membership, error) {
	q := s.db.Where("queue_id = ? AND is_active = ?", queueID, true)
	switch strategy {
	case StrategyLongestIdle:
		q = q.Order("priority DESC, last_call_ended_at ASC NULLS FIRST, user_id ASC")
	default: // round_robin
		q = q.Order("priority DESC, user_id ASC")
	}
	var out []*Membership
	err := q.Find(&out).Error
	return out, err
}

// MarkMemberCallEnded updates the longest_idle bookkeeping when an agent's
// call ends.
func (s *Store) MarkMemberCallEnded(queueID, userID string, at time.Time) error {
	return s.db.Model(&Membership{}).
		Where("queue_id = ? AND user_id = ?", queueID, userID).
		Updates(map[string]interface{}{
			"last_call_ended_at": &at,
			"updated_at":         time.Now().UTC(),
		}).Error
}

// ── Entries ─────────────────────────────────────────────────────────────────

func (s *Store) CreateEntry(e *Entry) (*Entry, error) {
	if err := s.db.Create(e).Error; err != nil {
		return nil, err
	}
	return e, nil
}

// GetEntryByCallID is the routing engine's lookup when a call.ended event
// arrives — we need to know which queue (if any) the call was in to mark
// the entry abandoned/connected.
func (s *Store) GetEntryByCallID(callID string) (*Entry, error) {
	var e Entry
	err := s.db.Where("call_id = ? AND ended_at IS NULL", callID).Order("entered_at DESC").First(&e).Error
	return &e, err
}

func (s *Store) ListLiveEntries(queueID, status string) ([]*Entry, error) {
	q := s.db.Where("queue_id = ?", queueID).Where("ended_at IS NULL")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var out []*Entry
	err := q.Order("priority DESC, entered_at ASC").Find(&out).Error
	return out, err
}

// AssignAgent moves the entry to "ringing" with the supplied agent. Used when
// the routing engine picks the next member.
func (s *Store) AssignAgent(entryID, agentID string) error {
	now := time.Now().UTC()
	return s.db.Model(&Entry{}).Where("id = ?", entryID).Updates(map[string]interface{}{
		"status":            string(EntryRinging),
		"assigned_agent_id": agentID,
		"assigned_at":       &now,
		"updated_at":        now,
	}).Error
}

// MarkConnected fires when the agent answers the assigned leg.
func (s *Store) MarkConnected(entryID string) error {
	return s.db.Model(&Entry{}).Where("id = ?", entryID).Updates(map[string]interface{}{
		"status":     string(EntryConnected),
		"updated_at": time.Now().UTC(),
	}).Error
}

// EndEntry finalises the entry with the given terminal status.
func (s *Store) EndEntry(entryID string, status EntryStatus, reason string) error {
	now := time.Now().UTC()
	return s.db.Model(&Entry{}).Where("id = ?", entryID).Updates(map[string]interface{}{
		"status":       string(status),
		"ended_at":     &now,
		"ended_reason": reason,
		"updated_at":   now,
	}).Error
}

// CountWaiting returns how many callers are currently waiting in a queue —
// useful for overflow checks and wallboard summaries.
func (s *Store) CountWaiting(queueID string) (int64, error) {
	var n int64
	err := s.db.Model(&Entry{}).
		Where("queue_id = ? AND status = ?", queueID, string(EntryWaiting)).
		Count(&n).Error
	return n, err
}
