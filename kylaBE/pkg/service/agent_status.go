package service

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AgentStatusChange int

const (
	UNKNOWN_STATUS_TYPE AgentStatusChange = iota
	ONLINE                                = "ONLINE"
	OFFLINE                               = "OFFLINE"
	BUSY                                  = "BUSY"
	AVAILABLE                             = "AVAILABLE"
	IN_A_MEETING                          = "IN_A_MEETING"
	IN_A_CALL                             = "IN_A_CALL"
	ON_A_BREAK                            = "IN_A_BREAK"
	ON_OUTBOUND_QUEUE                     = "ON_OUTBOUND_QUEUE"
	ON_INBOUND_QUEUE                      = "ON_INBOUND_QUEUE"
	ON_WRAP_UP                            = "ON_WRAP_UP"
	ON_MONITORING_QUEUE                   = "ON_MONITORING_QUEUE"
	ON_TICKET_QUEUE                       = "ON_TICKET_QUEUE"
	CUSTOM_STATUS                         = "CUSTOM_STATUS"
)

type StatusChange struct {
	gorm.Model
	ID            uuid.UUID         `gorm:"primarykey;type:uuid;not null"`
	StatusType    AgentStatusChange `gorm:"type:varchar(255);not null;"`
	Description   string            `gorm:"type:varchar(255);not null;"`
	StartTime     time.Time         `gorm:"type:time;not null;"`
	EndTime       time.Time         `gorm:"type:time;not null;"`
	AgentStatusID uuid.UUID         `gorm:"type:uuid;on_update:CASCADE; on_delete:CASCADE"`
	OwnerType     OwnerType         `gorm:"type:text;not null;"`
	OwnerId       uuid.UUID         `gorm:"type:uuid;not null;"`
}

type AgentStatus struct {
	gorm.Model
	ID            uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	AgentID       uuid.UUID
	SerialNumber  string
	StatusChanges []StatusChange `gorm:"foreignKey:AgentStatusID"`
	OwnerType     OwnerType      `gorm:"type:text;not null;"`
	OwnerId       uuid.UUID      `gorm:"type:uuid;not null;"`
}

type CustomStatus struct {
	gorm.Model
	ID          uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	Name        string    `gorm:"type:varchar(255);not null;"`
	Description string    `gorm:"type:varchar(255);not null;"`
	StartTime   time.Time `gorm:"type:time;not null;"`
	EndTime     time.Time `gorm:"type:time;not null;"`
	OwnerType   OwnerType `gorm:"type:text;not null;"`
	OwnerId     uuid.UUID `gorm:"type:uuid;not null;"`
}
