package communication

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConversationStore is the DB layer for the conversations table.
type ConversationStore struct {
	db *gorm.DB
}

// NewConversationStore constructs a ConversationStore.
func NewConversationStore(db *gorm.DB) *ConversationStore {
	return &ConversationStore{db: db}
}

// Create inserts a new Conversation. If ID is empty a UUID is generated.
func (s *ConversationStore) Create(c *Conversation) (*Conversation, error) {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
	if err := s.db.Create(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

// FindByID returns a Conversation by primary key scoped to org.
func (s *ConversationStore) FindByID(id, orgID string) (*Conversation, error) {
	var c Conversation
	if err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// FindByChannelRef looks up a conversation by (org, channel, channel_ref).
// Returns nil, nil when no row is found.
func (s *ConversationStore) FindByChannelRef(orgID, channel, channelRef string) (*Conversation, error) {
	var c Conversation
	err := s.db.Where("org_id = ? AND channel = ? AND channel_ref = ?", orgID, channel, channelRef).
		First(&c).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ListConversationsParams bundles filters for ListConversations.
type ListConversationsParams struct {
	OrgID       string
	WorkspaceID string
	Status      string
	Channel     string
	AssignedTo  string
	ActiveOnly  bool   // Filter out resolved/closed conversations
	PageSize    int
	PageToken   string // last seen ID (keyset)
}

// ListConversations returns a paginated inbox list.
func (s *ConversationStore) ListConversations(p ListConversationsParams) ([]*Conversation, string, int64, error) {
	if p.PageSize <= 0 || p.PageSize > 200 {
		p.PageSize = 50
	}

	q := s.db.Model(&Conversation{}).Where("org_id = ?", p.OrgID)
	if p.WorkspaceID != "" {
		q = q.Where("workspace_id = ?", p.WorkspaceID)
	}
	if p.Status != "" {
		q = q.Where("status = ?", p.Status)
	}
	if p.Channel != "" {
		q = q.Where("channel = ?", p.Channel)
	}
	if p.AssignedTo != "" {
		q = q.Where("assigned_to = ?", p.AssignedTo)
	}
	// Recommendation #5: Add FindAll support for active_only filtering
	if p.ActiveOnly {
		q = q.Where("status NOT IN (?)", []string{StatusResolved})
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, "", 0, err
	}

	if p.PageToken != "" {
		q = q.Where("id > ?", p.PageToken)
	}

	var convs []*Conversation
	if err := q.Order("created_at DESC").Limit(p.PageSize + 1).Find(&convs).Error; err != nil {
		return nil, "", total, err
	}

	nextToken := ""
	if len(convs) > p.PageSize {
		nextToken = convs[p.PageSize-1].ID
		convs = convs[:p.PageSize]
	}
	return convs, nextToken, total, nil
}

// SetPriority updates only the priority field on a conversation.
func (s *ConversationStore) SetPriority(id, orgID, priority string) (*Conversation, error) {
	c, err := s.FindByID(id, orgID)
	if err != nil {
		return nil, err
	}
	c.Priority = priority
	c.UpdatedAt = time.Now()
	if err := s.db.Model(c).Where("id = ? AND org_id = ?", id, orgID).Updates(map[string]interface{}{
		"priority":   priority,
		"updated_at": c.UpdatedAt,
	}).Error; err != nil {
		return nil, err
	}
	return c, nil
}

// Update saves all mutable fields on an existing Conversation.
func (s *ConversationStore) Update(c *Conversation) (*Conversation, error) {
	c.UpdatedAt = time.Now()
	if err := s.db.Model(c).Where("id = ? AND org_id = ?", c.ID, c.OrgID).
		Updates(map[string]interface{}{
			"assigned_to":   c.AssignedTo,
			"team_id":       c.TeamID,
			"status":        c.Status,
			"priority":      c.Priority,
			"subject":       c.Subject,
			"sla_deadline":  c.SLADeadline,
			"snoozed_until": c.SnoozedUntil,
			"resolved_at":   c.ResolvedAt,
			"meta":          c.Meta,
			"updated_at":    c.UpdatedAt,
		}).Error; err != nil {
		return nil, err
	}
	return c, nil
}

// AssignTo sets assigned_to (and optionally team_id) on a conversation.
func (s *ConversationStore) AssignTo(id, orgID, userID, teamID string) (*Conversation, error) {
	c, err := s.FindByID(id, orgID)
	if err != nil {
		return nil, err
	}
	c.AssignedTo = userID
	if teamID != "" {
		c.TeamID = teamID
	}
	return s.Update(c)
}

// SetStatus updates a conversation's status plus any associated timestamp.
func (s *ConversationStore) SetStatus(id, orgID, status string, extra map[string]interface{}) (*Conversation, error) {
	c, err := s.FindByID(id, orgID)
	if err != nil {
		return nil, err
	}

	c.Status = status
	c.UpdatedAt = time.Now()

	updates := map[string]interface{}{
		"status":     status,
		"updated_at": c.UpdatedAt,
	}
	for k, v := range extra {
		updates[k] = v
		switch k {
		case "resolved_at":
			t := v.(time.Time)
			c.ResolvedAt = &t
		case "snoozed_until":
			t := v.(time.Time)
			c.SnoozedUntil = &t
		}
	}

	if err := s.db.Model(c).Where("id = ? AND org_id = ?", id, orgID).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("set status: %w", err)
	}
	return c, nil
}
