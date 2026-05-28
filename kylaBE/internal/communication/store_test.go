package communication

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to open test database")

	// Auto-migrate models
	err = db.AutoMigrate(
		&Conversation{},
		&Message{},
		&RoutingRule{},
		&SLAPolicy{},
	)
	require.NoError(t, err, "Failed to migrate test schema")

	return db
}

// TestFindByChannelRef_NilHandling verifies nil-safety when no conversation exists.
// Recommendation #2: Integration tests for FindByChannelRef nil handling
func TestFindByChannelRef_NilHandling(t *testing.T) {
	db := setupTestDB(t)
	store := NewConversationStore(db)

	orgID := uuid.New().String()
	channel := ChannelWhatsApp
	channelRef := "+15551234567"

	t.Run("returns nil without error when not found", func(t *testing.T) {
		conv, err := store.FindByChannelRef(orgID, channel, channelRef)
		assert.NoError(t, err)
		assert.Nil(t, conv)
	})

	t.Run("returns conversation when exists", func(t *testing.T) {
		// Create conversation
		created, err := store.Create(&Conversation{
			OrgID:       orgID,
			WorkspaceID: uuid.New().String(),
			Channel:     channel,
			ChannelRef:  channelRef,
			Status:      StatusOpen,
		})
		require.NoError(t, err)

		// Lookup by channel ref
		found, err := store.FindByChannelRef(orgID, channel, channelRef)
		assert.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, created.ID, found.ID)
	})

	t.Run("returns nil for different org", func(t *testing.T) {
		differentOrgID := uuid.New().String()
		conv, err := store.FindByChannelRef(differentOrgID, channel, channelRef)
		assert.NoError(t, err)
		assert.Nil(t, conv)
	})
}

// TestListConversations_ActiveOnlyFilter verifies active_only filtering.
// Recommendation #5: Add FindAll support for active_only filtering
func TestListConversations_ActiveOnlyFilter(t *testing.T) {
	db := setupTestDB(t)
	store := NewConversationStore(db)

	orgID := uuid.New().String()
	workspaceID := uuid.New().String()

	// Create test conversations
	activeConv, err := store.Create(&Conversation{
		OrgID:       orgID,
		WorkspaceID: workspaceID,
		Channel:     ChannelEmail,
		ChannelRef:  "active@test.com",
		Status:      StatusOpen,
	})
	require.NoError(t, err)

	now := time.Now()
	resolvedConv, err := store.Create(&Conversation{
		OrgID:       orgID,
		WorkspaceID: workspaceID,
		Channel:     ChannelSMS,
		ChannelRef:  "+15551111111",
		Status:      StatusResolved,
		ResolvedAt:  &now,
	})
	require.NoError(t, err)

	t.Run("returns all conversations without filter", func(t *testing.T) {
		convs, _, total, err := store.ListConversations(ListConversationsParams{
			OrgID:       orgID,
			WorkspaceID: workspaceID,
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, convs, 2)
	})

	t.Run("filters out resolved when active_only=true", func(t *testing.T) {
		convs, _, total, err := store.ListConversations(ListConversationsParams{
			OrgID:       orgID,
			WorkspaceID: workspaceID,
			ActiveOnly:  true,
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, convs, 1)
		assert.Equal(t, activeConv.ID, convs[0].ID)
	})

	t.Run("status filter works independently", func(t *testing.T) {
		convs, _, _, err := store.ListConversations(ListConversationsParams{
			OrgID:       orgID,
			WorkspaceID: workspaceID,
			Status:      StatusResolved,
		})
		assert.NoError(t, err)
		require.Len(t, convs, 1)
		assert.Equal(t, resolvedConv.ID, convs[0].ID)
	})
}

// TestSetPriority verifies dedicated priority update method.
func TestSetPriority(t *testing.T) {
	db := setupTestDB(t)
	store := NewConversationStore(db)

	orgID := uuid.New().String()
	conv, err := store.Create(&Conversation{
		OrgID:       orgID,
		WorkspaceID: uuid.New().String(),
		Channel:     ChannelWebChat,
		ChannelRef:  "session-123",
		Status:      StatusOpen,
		Priority:    "medium",
	})
	require.NoError(t, err)

	t.Run("updates priority without affecting status", func(t *testing.T) {
		updated, err := store.SetPriority(conv.ID, orgID, "urgent")
		require.NoError(t, err)
		assert.Equal(t, "urgent", updated.Priority)
		assert.Equal(t, StatusOpen, updated.Status)
	})

	t.Run("returns error for non-existent conversation", func(t *testing.T) {
		_, err := store.SetPriority(uuid.New().String(), orgID, "high")
		assert.Error(t, err)
	})
}

// TestConversationCreation verifies basic CRUD.
func TestConversationCreation(t *testing.T) {
	db := setupTestDB(t)
	store := NewConversationStore(db)

	t.Run("creates conversation with auto-generated ID", func(t *testing.T) {
		conv, err := store.Create(&Conversation{
			OrgID:       uuid.New().String(),
			WorkspaceID: uuid.New().String(),
			Channel:     ChannelEmail,
			ChannelRef:  "test@example.com",
			Status:      StatusOpen,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, conv.ID)
		assert.False(t, conv.CreatedAt.IsZero())
	})

	t.Run("respects provided ID", func(t *testing.T) {
		customID := uuid.New().String()
		conv, err := store.Create(&Conversation{
			ID:          customID,
			OrgID:       uuid.New().String(),
			WorkspaceID: uuid.New().String(),
			Channel:     ChannelSMS,
			ChannelRef:  "+15552222222",
			Status:      StatusOpen,
		})
		require.NoError(t, err)
		assert.Equal(t, customID, conv.ID)
	})
}
