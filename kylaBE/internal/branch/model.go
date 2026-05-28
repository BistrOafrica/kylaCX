package branch

import (
	"kyla-be/internal/authctx"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Branch represents an organisational branch.
// Association slice fields (Users, Roles, Teams, Departments) are managed through
// their respective relationship records and are not declared here to avoid import cycles.
// All queries using these associations are performed through the BranchStore or raw DB.
type Branch struct {
	gorm.Model
	ID           uuid.UUID         `gorm:"primarykey;type:uuid;not null"`
	Name         string            `gorm:"type:text;"`
	Description  string            `gorm:"type:text;"`
	ParentID     uuid.UUID         `gorm:"type:uuid;default:00000000-0000-0000-0000-000000000000"`
	Status       string            `gorm:"default:'ACTIVE'"`
	CreatedBy    string            `gorm:"type:text;"`
	SerialNumber string            `gorm:"type:text;"`
	IsDefault    bool              `gorm:"default:false"`
	UpdatedBy    string            `gorm:"type:text;"`
	OwnerType    authctx.OwnerType `gorm:"type:text;not null;default:0"`
	OwnerID      uuid.UUID         `gorm:"type:uuid;not null;default:00000000-0000-0000-0000-000000000000"`
	Location     string            `gorm:"type:text;"`
	Address      string            `gorm:"type:text;"`
}
