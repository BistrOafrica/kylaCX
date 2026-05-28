package ticketing

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TicketingStore is the database layer for ticket rooms, messages, and macros.
type TicketingStore struct {
	db *gorm.DB
}

// NewTicketingStore constructs a TicketingStore.
func NewTicketingStore(db *gorm.DB) *TicketingStore {
	return &TicketingStore{db: db}
}

// ── Ticket Rooms ──────────────────────────────────────────────────────────────

func (s *TicketingStore) CreateRoom(r *TicketRoom) (*TicketRoom, error) {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	now := time.Now()
	r.CreatedAt = now
	r.UpdatedAt = now
	if err := s.db.Create(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

func (s *TicketingStore) FindRoomByID(id, orgID string) (*TicketRoom, error) {
	var r TicketRoom
	if err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&r).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *TicketingStore) ListRooms(ticketID, orgID string) ([]*TicketRoom, error) {
	var rooms []*TicketRoom
	if err := s.db.Where("ticket_id = ? AND org_id = ?", ticketID, orgID).
		Order("created_at ASC").Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

// ── Room Messages ─────────────────────────────────────────────────────────────

func (s *TicketingStore) AddMessage(m *TicketRoomMessage) (*TicketRoomMessage, error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	m.CreatedAt = time.Now()
	return m, s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(m).Error; err != nil {
			return err
		}
		// Increment message_count on the room
		return tx.Model(&TicketRoom{}).Where("id = ? AND org_id = ?", m.RoomID, m.OrgID).
			UpdateColumn("message_count", gorm.Expr("message_count + 1")).Error
	})
}

func (s *TicketingStore) ListMessages(roomID, orgID, before string, limit int) ([]*TicketRoomMessage, bool, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	q := s.db.Where("room_id = ? AND org_id = ?", roomID, orgID)
	if before != "" {
		if beforeTS, beforeID, ok := parseCursor(before); ok {
			q = q.Where("(created_at < ?) OR (created_at = ? AND id < ?)", beforeTS, beforeTS, beforeID)
		} else {
			var pivot TicketRoomMessage
			if err := s.db.Select("id", "created_at").
				Where("id = ? AND room_id = ? AND org_id = ?", before, roomID, orgID).
				First(&pivot).Error; err == nil {
				q = q.Where("(created_at < ?) OR (created_at = ? AND id < ?)", pivot.CreatedAt, pivot.CreatedAt, pivot.ID)
			}
		}
	}
	q = q.Order("created_at DESC, id DESC").Limit(limit + 1)

	var msgs []*TicketRoomMessage
	if err := q.Find(&msgs).Error; err != nil {
		return nil, false, err
	}

	hasMore := len(msgs) > limit
	if hasMore {
		msgs = msgs[:limit]
	}
	// Reverse so oldest-first
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, hasMore, nil
}

func parseCursor(token string) (time.Time, string, bool) {
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

// ── Macros ────────────────────────────────────────────────────────────────────

func (s *TicketingStore) CreateMacro(m *Macro) (*Macro, error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	if err := s.db.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

func (s *TicketingStore) FindMacroByID(id, orgID string) (*Macro, error) {
	var m Macro
	if err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *TicketingStore) ListMacros(orgID, workspaceID string) ([]*Macro, error) {
	q := s.db.Where("org_id = ?", orgID)
	if workspaceID != "" {
		q = q.Where("workspace_id = ? OR visibility = 'public'", workspaceID)
	}
	var macros []*Macro
	if err := q.Order("name ASC").Find(&macros).Error; err != nil {
		return nil, err
	}
	return macros, nil
}

func (s *TicketingStore) UpdateMacro(id, orgID string, updates map[string]interface{}) (*Macro, error) {
	updates["updated_at"] = time.Now()
	if err := s.db.Model(&Macro{}).Where("id = ? AND org_id = ?", id, orgID).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.FindMacroByID(id, orgID)
}

func (s *TicketingStore) DeleteMacro(id, orgID string) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&Macro{}).Error
}

// GetTicketData returns the JSONB data of a ticket Object Core record.
func (s *TicketingStore) GetTicketData(ticketID, orgID string) (map[string]interface{}, error) {
	var result struct{ Data []byte }
	if err := s.db.Raw("SELECT data FROM objects WHERE id = ? AND org_id = ? AND type_slug = 'ticket'",
		ticketID, orgID).Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}
	if result.Data == nil {
		return nil, fmt.Errorf("ticket %s not found", ticketID)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// PatchTicketData writes updated field data back to a ticket object.
func (s *TicketingStore) PatchTicketData(ticketID, orgID string, data map[string]interface{}) error {
	merged, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.db.Exec(
		"UPDATE objects SET data = ?::jsonb, updated_at = ? WHERE id = ? AND org_id = ?",
		string(merged), time.Now(), ticketID, orgID,
	).Error
}
