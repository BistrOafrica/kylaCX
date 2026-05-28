package ticketing

import "time"

// TicketRoom is a threaded discussion space attached to a ticket.
type TicketRoom struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TicketID     string    `gorm:"type:uuid;not null;index"                       json:"ticket_id"`
	OrgID        string    `gorm:"type:uuid;not null;index"                       json:"org_id"`
	Name         string    `gorm:"not null"                                       json:"name"`
	Type         string    `gorm:"not null;default:'internal'"                    json:"type"` // "internal" | "customer_reply"
	MessageCount int       `gorm:"default:0"                                      json:"message_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (TicketRoom) TableName() string { return "ticket_rooms" }

// TicketRoomMessage is a single message in a ticket room.
type TicketRoomMessage struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RoomID    string    `gorm:"type:uuid;not null;index"                       json:"room_id"`
	OrgID     string    `gorm:"type:uuid;not null;index"                       json:"org_id"`
	AuthorID  string    `gorm:"type:uuid;not null"                             json:"author_id"`
	Content   string    `gorm:"not null;type:text"                             json:"content"`
	IsPrivate bool      `gorm:"default:false"                                  json:"is_private"`
	CreatedAt time.Time `json:"created_at"`
}

func (TicketRoomMessage) TableName() string { return "ticket_room_messages" }

// Macro is a canned response + field-patch action set for ticket handling.
type Macro struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID       string    `gorm:"type:uuid;not null;index"                       json:"org_id"`
	WorkspaceID string    `gorm:"type:uuid;index"                                json:"workspace_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	Content     string    `gorm:"not null;type:text"                             json:"content"`
	Actions     string    `gorm:"type:jsonb;not null;default:'[]'"               json:"actions"` // JSON array of field patches
	Visibility  string    `gorm:"not null;default:'private'"                     json:"visibility"` // "private"|"team"|"public"
	CreatedBy   string    `gorm:"type:uuid"                                      json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Macro) TableName() string { return "ticket_macros" }
