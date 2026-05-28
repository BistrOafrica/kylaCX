package ticketing
package ticketing

import (
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTicketingTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE ticket_rooms (
			id TEXT PRIMARY KEY,
			ticket_id TEXT NOT NULL,
			org_id TEXT NOT NULL,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			message_count INTEGER NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error; err != nil {
		t.Fatalf("create ticket_rooms schema: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE ticket_room_messages (
			id TEXT PRIMARY KEY,
			room_id TEXT NOT NULL,
			org_id TEXT NOT NULL,
			author_id TEXT NOT NULL,
			content TEXT NOT NULL,
			is_private BOOLEAN NOT NULL,
			created_at DATETIME
		)
	`).Error; err != nil {
		t.Fatalf("create ticket_room_messages schema: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE ticket_macros (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			workspace_id TEXT,
			name TEXT NOT NULL,
			content TEXT NOT NULL,
			actions TEXT NOT NULL,
			visibility TEXT NOT NULL,
			created_by TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error; err != nil {
		t.Fatalf("create ticket_macros schema: %v", err)
	}

	return db
}

func TestTicketingStoreMutationSmoke(t *testing.T) {
	db := setupTicketingTestDB(t)
	store := NewTicketingStore(db)

	room, err := store.CreateRoom(&TicketRoom{
		TicketID: "ticket1",
		OrgID:    "org1",
		Name:     "Internal Notes",
		Type:     "internal",
	})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	_, err = store.AddMessage(&TicketRoomMessage{
		RoomID:    room.ID,
		OrgID:     "org1",
		AuthorID:  "user1",
		Content:   "hello",
		IsPrivate: true,
	})
	if err != nil {
		t.Fatalf("add message: %v", err)
	}

	reloadedRoom, err := store.FindRoomByID(room.ID, "org1")
	if err != nil {
		t.Fatalf("find room: %v", err)
	}
	if reloadedRoom.MessageCount != 1 {
		t.Fatalf("expected message_count=1, got %d", reloadedRoom.MessageCount)
	}

	macro, err := store.CreateMacro(&Macro{
		OrgID:       "org1",
		WorkspaceID: "ws1",
		Name:        "Close ticket",
		Content:     "Closing this now",
		Actions:     `[{"field":"status","value":"closed"}]`,
		Visibility:  "private",
		CreatedBy:   "user1",
	})
	if err != nil {
		t.Fatalf("create macro: %v", err)
	}

	updatedMacro, err := store.UpdateMacro(macro.ID, "org1", map[string]interface{}{
		"name": "Close ticket now",
	})
	if err != nil {
		t.Fatalf("update macro: %v", err)
	}
	if updatedMacro.Name != "Close ticket now" {
		t.Fatalf("expected updated macro name, got %q", updatedMacro.Name)
	}

	if err := store.DeleteMacro(macro.ID, "org1"); err != nil {
		t.Fatalf("delete macro: %v", err)
	}
	if _, err := store.FindMacroByID(macro.ID, "org1"); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected macro not found after delete, got: %v", err)
	}
}
