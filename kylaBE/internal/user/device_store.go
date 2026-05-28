package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserDeviceInfo represents a device used by a user.
type UserDeviceInfo struct {
	gorm.Model
	ID          uuid.UUID  `gorm:"type:uuid;primary_key"`
	UserID      uuid.UUID  `gorm:"type:uuid;index"`
	IPAddress   string     `gorm:"index"`
	DeviceMacID string     `gorm:"index"`
	DeviceType  string
	OSType      string
	DeviceName  string
	UserAgent   string
	ClientID    string     `gorm:"index"`
	LastSeenAt  time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time `gorm:"index"`
	IsTrusted   bool
	IsBrowser   bool
	IsActive    bool
}

// DeviceStore manages user device records.
type DeviceStore struct {
	db *gorm.DB
}

// NewDeviceStore creates a new DeviceStore.
func NewDeviceStore(db *gorm.DB) *DeviceStore {
	return &DeviceStore{db: db}
}

// FindByClientID finds a device by client ID.
// Returns nil, nil when no record is found.
func (s *DeviceStore) FindByClientID(clientID string) (*UserDeviceInfo, error) {
	if clientID == "" {
		return nil, errors.New("client ID is required")
	}
	var device UserDeviceInfo
	result := s.db.Where("client_id = ?", clientID).First(&device)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &device, nil
}

// FindByIPAndUserAgent finds a device by IP address and user agent.
// Returns nil, nil when no record is found.
func (s *DeviceStore) FindByIPAndUserAgent(ipAddress, userAgent string) (*UserDeviceInfo, error) {
	if ipAddress == "" || userAgent == "" {
		return nil, errors.New("IP address and user agent are required")
	}
	var device UserDeviceInfo
	result := s.db.Where("ip_address = ? AND user_agent = ?", ipAddress, userAgent).First(&device)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &device, nil
}
