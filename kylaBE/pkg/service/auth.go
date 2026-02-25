package service

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserDevice struct {
	gorm.Model
	ID         uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	UserID     uuid.UUID `gorm:"type:uuid;not null"`
	DeviceType string    `gorm:"type:varchar(100);not null"`
	OSType     string    `gorm:"type:varchar(100);not null"`
	DeviceName string    `gorm:"type:varchar(100);not null"`
	UserAgent  string    `gorm:"type:varchar(255);not null"`
	IsTrusted  bool      `gorm:"type:boolean;not null;default:false"`
	IsBrowser  bool      `gorm:"type:boolean;not null;default:false"`
	IsActive   bool      `gorm:"type:boolean;not null;default:true"`
}

type UserSession struct {
	gorm.Model
	ID           uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	UserID       uuid.UUID `gorm:"type:uuid;not null"`
	DeviceID     uuid.UUID `gorm:"type:uuid"`
	StartTime    string    `gorm:"type:varchar(255);not null"`
	EndTime      string    `gorm:"type:varchar(255);"`
	IsValid      bool      `gorm:"type:boolean;not null;default:true"`
	IpAddress    string    `gorm:"type:varchar(100);"`
	LastLoggedIn string    `gorm:"type:varchar(255);"`
}
