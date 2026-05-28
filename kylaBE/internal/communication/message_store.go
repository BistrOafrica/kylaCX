package communication

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageStore is the DB layer for the messages table.
type MessageStore struct {
	db *gorm.DB
}

// NewMessageStore constructs a MessageStore.
func NewMessageStore(db *gorm.DB) *MessageStore {
	return &MessageStore{db: db}
}

// Create inserts a new Message.
func (s *MessageStore) Create(m *Message) (*Message, error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	m.CreatedAt = time.Now()
	if err := s.db.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// FindByExternalID looks up a message by its provider-assigned ID.
// Returns nil, nil when not found.
func (s *MessageStore) FindByExternalID(externalID string) (*Message, error) {
	if externalID == "" {
		return nil, nil
	}
	var m Message
	err := s.db.Where("external_id = ?", externalID).First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// ListByConversation returns messages for a conversation, newest-first.
// before is an optional message ID cursor — only messages older than that ID are returned.
func (s *MessageStore) ListByConversation(conversationID string, limit int, before string) ([]*Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := s.db.Where("conversation_id = ?", conversationID)
	if before != "" {
		q = q.Where("id < ?", before)
	}
	var msgs []*Message
	if err := q.Order("created_at DESC").Limit(limit).Find(&msgs).Error; err != nil {
		return nil, err
	}
	return msgs, nil
}

// UpdateStatus sets the delivery status of a message.
func (s *MessageStore) UpdateStatus(id, status string) error {
	return s.db.Model(&Message{}).Where("id = ?", id).
		Update("status", status).Error
}
