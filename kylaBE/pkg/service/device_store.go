package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserDeviceInfo represents a device used by a user
type UserDeviceInfo struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID      uuid.UUID `gorm:"type:uuid;index"`
	IPAddress   string    `gorm:"index"`
	DeviceMacID string    `gorm:"index"`
	DeviceType  string
	OSType      string
	DeviceName  string
	UserAgent   string
	ClientID    string `gorm:"index"`
	LastSeenAt  time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time `gorm:"index"`
	IsTrusted   bool
	IsBrowser   bool
	IsActive    bool
}

// DeviceStore manages user device records
type DeviceStore struct {
	db *gorm.DB
}

// NewDeviceStore creates a new DeviceStore
func NewDeviceStore(db *gorm.DB) *DeviceStore {
	return &DeviceStore{db: db}
}

// FindByClientID finds a device by client ID
func (s *DeviceStore) FindByClientID(clientID string) (*UserDeviceInfo, error) {
	if clientID == "" {
		return nil, errors.New("client ID is required")
	}

	var device UserDeviceInfo
	result := s.db.Where("client_id = ?", clientID).First(&device)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil when no record found
		}
		return nil, result.Error
	}
	return &device, nil
}

// FindByIPAndUserAgent finds a device by IP address and user agent
func (s *DeviceStore) FindByIPAndUserAgent(ipAddress, userAgent string) (*UserDeviceInfo, error) {
	if ipAddress == "" || userAgent == "" {
		return nil, errors.New("IP address and user agent are required")
	}

	var device UserDeviceInfo
	result := s.db.Where("ip_address = ? AND user_agent = ?", ipAddress, userAgent).First(&device)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil when no record found
		}
		return nil, result.Error
	}
	return &device, nil
}

// // CreateDevice creates a new user device record
// func (s *DeviceStore) CreateDevice(userID uuid.UUID, clientInfo *ClientInfo) (*UserDeviceInfo, error) {
// 	device := &UserDeviceInfo{
// 		ID:          uuid.New(),
// 		UserID:      userID,
// 		IPAddress:   clientInfo.IPAddress,
// 		DeviceMacID: clientInfo.DeviceMacID,
// 		DeviceType:  clientInfo.DeviceType,
// 		OSType:      clientInfo.OSType,
// 		DeviceName:  clientInfo.DeviceName,
// 		UserAgent:   clientInfo.UserAgent,
// 		ClientID:    clientInfo.ClientID,
// 		LastSeenAt:  time.Now(),
// 		CreatedAt:   time.Now(),
// 		UpdatedAt:   time.Now(),
// 	}

// 	result := s.db.Create(device)
// 	if result.Error != nil {
// 		return nil, result.Error
// 	}
// 	return device, nil
// }

// // UpdateLastSeen updates the last seen timestamp for a device
// func (s *DeviceStore) UpdateLastSeen(deviceID uuid.UUID) error {
// 	result := s.db.Model(&UserDeviceInfo{}).Where("id = ?", deviceID).Update("last_seen_at", time.Now())
// 	return result.Error
// }
