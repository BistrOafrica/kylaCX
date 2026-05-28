package apps

import (
	"time"

	authctx "kyla-be/internal/authctx"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type App struct {
	gorm.Model
	ID                  uuid.UUID      `gorm:"primaryKey;type:uuid;not null"`
	Token               string         `gorm:"not null"`
	Name                string         `gorm:"not null"`
	Description         string         `gorm:"not null"`
	Secret              string         `gorm:"not null"`
	Status              string         `gorm:"not null"`
	SerialNumber        string         `gorm:"type:varchar(100)"`
	PermissionCodeNames pq.StringArray `gorm:"type:text[]; not null"`
	CreatedBy           string         `gorm:"not null"`
	UpdatedBy           string
	ApprovedBy          string
	ApprovedAt          time.Time         `gorm:"default:null"`
	RejectedBy          string            `gorm:"default:null"`
	RejectedAt          time.Time         `gorm:"default:null"`
	IsTemplate          bool              `gorm:"default:false"`
	OwnerType           authctx.OwnerType `gorm:"not null; default:0"`
	OwnerId             string            `gorm:"not null; default:USERS"`
	WorkspaceID         *string           `gorm:"type:uuid;index" json:"workspace_id,omitempty"` // Optional workspace scope
}
