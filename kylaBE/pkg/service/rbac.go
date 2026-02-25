package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	ID                  uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	Name                string    `gorm:"not null"`
	SerialNumber        string
	Description         string
	PermissionCodeNames pq.StringArray `gorm:"type:text[]"`
	CreatedBy           string         `gorm:"not null"`
	UpdatedBy           string
	UpdatedAt           time.Time `gorm:"not null"`
	CreatedAt           time.Time `gorm:"not null"`
	IsDefault           bool
	OwnerType           OwnerType
	OwnerID             uuid.UUID `gorm:"type:uuid;not null;default:00000000-0000-0000-0000-000000000000"`
}
