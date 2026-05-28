package audit

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog records every authorisation decision made by the auth interceptor.
type AuditLog struct {
	ID          uuid.UUID `gorm:"primarykey;type:uuid;default:gen_random_uuid()"`
	UserID      string    `gorm:"index;not null"`
	OrgID       string    `gorm:"index;not null"`
	WorkspaceID string    `gorm:"index"`
	Method      string    `gorm:"not null"` // gRPC full method
	Resource    string    // Casbin resource (obj)
	Action      string    // Casbin action (act)
	Allowed     bool      `gorm:"not null"`
	IPAddress   string
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}
