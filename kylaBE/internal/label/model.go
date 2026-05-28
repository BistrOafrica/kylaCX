package label

import (
	"kyla-be/internal/authctx"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Label struct {
	gorm.Model
	ID        uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	Name      string
	CreatedBy string
	OwnerID   uuid.UUID         `gorm:"not null; type:uuid; index"`
	OwnerType authctx.OwnerType `gorm:"not null; default:USERS; index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
