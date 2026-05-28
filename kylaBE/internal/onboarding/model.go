package onboarding

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OnboardingStatus string

const (
	StatusNew       OnboardingStatus = "NEW"
	StatusPending   OnboardingStatus = "PENDING"
	StatusCompleted OnboardingStatus = "COMPLETED"
	StatusFailed    OnboardingStatus = "FAILED"
)

type Onboarding struct {
	gorm.Model
	ID            uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	Timestamp     string    `gorm:"type:varchar(255);not null"`
	Status        string    `gorm:"type:varchar(100);not null"`
	Packages      string    `gorm:"type:text;not null"`
	Products      string    `gorm:"type:text;not null"`
	NumberOfUsers int       `gorm:"type:int;not null"`
	Remarks       string    `gorm:"type:text;"`
	ContactEmail  string    `gorm:"type:varchar(255);not null"`
	ContactPhone  string    `gorm:"type:varchar(50);not null"`
	Name          string    `gorm:"type:varchar(255);not null"`
	Metadata      string    `gorm:"type:text;"`
}

type OnboardingLog struct {
	gorm.Model
	ID           uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	OnboardingID uuid.UUID `gorm:"type:uuid;not null"`
	Timestamp    string    `gorm:"type:varchar(255);not null"`
	Action       string    `gorm:"type:varchar(100);not null"`
	Details      string    `gorm:"type:text;"`
}
