package sharing

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EdgeType classifies the relationship between two entities.
type EdgeType string

const (
	OWNS   EdgeType = "OWNS"
	SHARES EdgeType = "SHARES"
)

// Entity is a node in the sharing graph representing any principal or resource.
type Entity struct {
	gorm.Model
	ID              uuid.UUID    `gorm:"primaryKey;type:uuid"`
	Type            string       // "user", "org", "team", "department", "branch"
	Resources       []EntityLink `gorm:"foreignKey:FromID"`
	Shared          []EntityLink `gorm:"foreignKey:ToID"`
	OwnershipEntity bool         `gorm:"default:false"`
}

func (Entity) TableName() string {
	return "entities"
}

// EntityLink is a directed edge in the sharing graph.
type EntityLink struct {
	gorm.Model
	ID          uuid.UUID `gorm:"primaryKey;type:uuid"`
	FromID      uuid.UUID `gorm:"index;type:uuid"`
	FromType    string
	FromEntity  Entity   `gorm:"-:all"`
	ToID        uuid.UUID `gorm:"index;type:uuid"`
	ToType      string
	ToEntity    Entity   `gorm:"-:all"`
	SharedBy    string   `gorm:"index"`
	Type        EdgeType `gorm:"type:citext"` // "OWNS" or "SHARES"
	Roles       string   // JSON-encoded list of roles
	Permissions string   // JSON-encoded list of CRUD permissions
	Conditions  string   // JSON-encoded map of conditions
}

// AccessRequest represents a pending request for access to a resource.
type AccessRequest struct {
	gorm.Model
	ID             uuid.UUID `gorm:"primaryKey;type:uuid"`
	ResourceOwner  uuid.UUID `gorm:"type:uuid"`
	ResourceID     uuid.UUID `gorm:"index;type:uuid"`
	RequesterID    uuid.UUID `gorm:"index;type:uuid"`
	RequestedRoles string    // JSON-encoded list of roles
	Status         string
	Timestamp      time.Time
}
