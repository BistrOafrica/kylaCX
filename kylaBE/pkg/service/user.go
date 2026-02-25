package service

import (
	"kyla-be/pkg/k"
	"kyla-be/pkg/utils"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User contains the user information
type User struct {
	gorm.Model
	ID                    uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	SerialNumber          string
	FirstName             string `gorm:"not null"`
	LastName              string
	Username              string `gorm:"not null; unique"`
	Email                 string `gorm:"not null; unique"`
	Phone                 string
	HashedPassword        string `gorm:"not null"`
	Status                string
	AgentStatusID         uuid.UUID `gorm:"type:uuid; not null"`
	AgentStatus           AgentStatus
	IsDefault             bool   `gorm:"default:false"`
	CreatedBy             string `gorm:"not null;default:USERS"`
	UpdatedBy             string
	EmailSignature        string
	Shifts                []Shift        `gorm:"many2many:user_shifts;"`
	CurrentBranchID       uuid.UUID      `gorm:"type:uuid;"`
	CurrentOrganisationID uuid.UUID      `gorm:"type:uuid;"`
	Roles                 []Role         `gorm:"many2many:user_roles;"`
	Branches              []Branch       `gorm:"many2many:user_branches;"`
	Organisations         []Organisation `gorm:"many2many:organisation_users;"`
	Departments           []Department   `gorm:"many2many:user_departments;"`
	Teams                 []Team         `gorm:"many2many:user_teams;"`
	OwnerType             OwnerType
	OwnerID               uuid.UUID `gorm:"type:uuid;not null; default:00000000-0000-0000-0000-000000000000"`
	FirebaseUID           string
	ReferralCode          string `gorm:"not null;type:text;default:''"`
	// MFA fields
	MFAEnabled       bool      `gorm:"default:false"`
	MFASecret        string    // TOTP secret for authenticator apps
	MFARecoveryCodes []string  `gorm:"type:text[]"`       // Recovery codes for MFA
	Passkeys         []Passkey `gorm:"foreignKey:UserID"` // Passkeys for passwordless authentication
	LastMFALogin     time.Time // Last successful MFA login
	LastLogin        time.Time
	LastMfaLogin     time.Time
	Image            string
	CallingCode      string
}

// Passkey represents a WebAuthn credential
type Passkey struct {
	gorm.Model
	ID              uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	UserID          uuid.UUID `gorm:"type:uuid;not null"`
	CredentialID    []byte    `gorm:"type:bytea;not null"`
	PublicKey       []byte    `gorm:"type:bytea;not null"`
	AttestationType string
	Transport       []string `gorm:"type:text[]"`
	Flags           uint32
	Authenticator   AuthenticatorInfo
}

// AuthenticatorInfo contains information about the authenticator device
type AuthenticatorInfo struct {
	gorm.Model
	PasskeyID    uuid.UUID `gorm:"type:uuid;not null"`
	AAGUID       []byte    `gorm:"type:bytea"`
	SignCount    uint32
	CloneWarning bool
	Attachment   string
	UserVerified bool
	UserPresent  bool
}

func (u *User) CREATE_SUPER_USER_ROLE(roleStore *RbacStore) {
	id := uuid.New()
	permissions := k.ALL_PERMISSION_CODENAMES()
	role := &Role{
		ID:                  id,
		SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], id.String()),
		Name:                "Super User",
		Description:         "Super User Role",
		PermissionCodeNames: permissions,
		CreatedBy:           "USERS",
		UpdatedAt:           time.Now(),
		CreatedAt:           time.Now(),
		IsDefault:           true,
		OwnerType:           USERS,
		OwnerID:             u.ID,
	}
	u.Roles = append(u.Roles, *role)
}

func (u *User) CREATE_NEW_USER_ROLE(roleStore *RbacStore) {
	id := uuid.New()
	permissions := k.BASIC_PERMISSIONS()
	role := &Role{
		ID:                  id,
		SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], id.String()),
		Name:                "New User",
		Description:         "New User Role",
		PermissionCodeNames: permissions,
		CreatedBy:           "USERS",
		UpdatedAt:           time.Now(),
		CreatedAt:           time.Now(),
		IsDefault:           true,
		OwnerType:           USERS,
		OwnerID:             u.ID,
	}
	u.Roles = append(u.Roles, *role)
}

func (u *User) ADD_BASIC_ROLE(roleStore *RbacStore) {
	role, err := roleStore.FindDefaultRole(u.CurrentOrganisationID)
	if err != nil {
		log.Println("Basic Role not found")
	}
	u.Roles = append(u.Roles, *role)
}

func (u *User) CREATE_AGENT_STATUS(agentStore *StatusStore) error {
	id := uuid.New()
	agentStatus := &AgentStatus{
		ID:            id,
		AgentID:       u.ID,
		SerialNumber:  utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["agent_status"], id.String()),
		StatusChanges: []StatusChange{},
		OwnerType:     OwnerType(USERS),
		OwnerId:       u.ID,
	}
	u.AgentStatusID = id
	u.AgentStatus = *agentStatus
	if err := agentStore.Save(agentStatus); err != nil {
		return err
	}
	return nil
}
