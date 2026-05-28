package user

import (
	"kyla-be/internal/agentops"
	"kyla-be/internal/authctx"
	"kyla-be/internal/rbac"
	"kyla-be/pkg/k"
	"kyla-be/pkg/utils"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User contains the core user information.
// Many2many back-references to Branch, Organisation, Department, Team and Shift
// are intentionally omitted to prevent import cycles; those associations are
// managed through their respective join tables in the database.
type User struct {
	gorm.Model
	ID                    uuid.UUID          `gorm:"primarykey;type:uuid;not null"`
	SerialNumber          string
	FirstName             string             `gorm:"not null"`
	LastName              string
	Username              string             `gorm:"not null; unique"`
	Email                 string             `gorm:"not null; unique"`
	Phone                 string
	HashedPassword        string             `gorm:"not null"`
	Status                string
	AgentStatusID         uuid.UUID          `gorm:"type:uuid; not null"`
	AgentStatus           agentops.AgentStatus
	IsDefault             bool               `gorm:"default:false"`
	CreatedBy             string             `gorm:"not null;default:USERS"`
	UpdatedBy             string
	EmailSignature        string
	CurrentBranchID       uuid.UUID          `gorm:"type:uuid;"`
	CurrentOrganisationID uuid.UUID          `gorm:"type:uuid;"`
	Roles                 []rbac.Role        `gorm:"many2many:user_roles;"`
	OwnerType             authctx.OwnerType
	OwnerID               uuid.UUID          `gorm:"type:uuid;not null;default:00000000-0000-0000-0000-000000000000"`
	FirebaseUID           string
	ReferralCode          string             `gorm:"not null;type:text;default:''"`
	// MFA fields
	MFAEnabled       bool      `gorm:"default:false"`
	MFASecret        string
	MFARecoveryCodes []string  `gorm:"type:text[]"`
	Passkeys         []Passkey `gorm:"foreignKey:UserID"`
	LastMFALogin     time.Time
	LastLogin        time.Time
	LastMfaLogin     time.Time
	Image            string
	CallingCode      string
}

// Passkey represents a WebAuthn credential for passwordless authentication.
type Passkey struct {
	gorm.Model
	ID              uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	UserID          uuid.UUID `gorm:"type:uuid;not null"`
	CredentialID    []byte    `gorm:"type:bytea;not null"`
	PublicKey       []byte    `gorm:"type:bytea;not null"`
	AttestationType string
	Transport       []string          `gorm:"type:text[]"`
	Flags           uint32
	Authenticator   AuthenticatorInfo
}

// AuthenticatorInfo contains information about the authenticator device.
type AuthenticatorInfo struct {
	gorm.Model
	PasskeyID    uuid.UUID `gorm:"type:uuid;not null"`
	AAGUID       []byte    `gorm:"type:bytea"`
	SignCount     uint32
	CloneWarning  bool
	Attachment    string
	UserVerified  bool
	UserPresent   bool
}

// CREATE_SUPER_USER_ROLE creates and attaches a super-user role to the user.
func (u *User) CREATE_SUPER_USER_ROLE(roleStore *rbac.RbacStore) {
	id := uuid.New()
	permissions := k.ALL_PERMISSION_CODENAMES()
	role := &rbac.Role{
		ID:                  id,
		SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], id.String()),
		Name:                "Super User",
		Description:         "Super User Role",
		PermissionCodeNames: permissions,
		CreatedBy:           "USERS",
		UpdatedAt:           time.Now(),
		CreatedAt:           time.Now(),
		IsDefault:           true,
		OwnerType:           authctx.USERS,
		OwnerID:             u.ID,
	}
	u.Roles = append(u.Roles, *role)
}

// CREATE_NEW_USER_ROLE creates and attaches a basic new-user role.
func (u *User) CREATE_NEW_USER_ROLE(roleStore *rbac.RbacStore) {
	id := uuid.New()
	permissions := k.BASIC_PERMISSIONS()
	role := &rbac.Role{
		ID:                  id,
		SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], id.String()),
		Name:                "New User",
		Description:         "New User Role",
		PermissionCodeNames: permissions,
		CreatedBy:           "USERS",
		UpdatedAt:           time.Now(),
		CreatedAt:           time.Now(),
		IsDefault:           true,
		OwnerType:           authctx.USERS,
		OwnerID:             u.ID,
	}
	u.Roles = append(u.Roles, *role)
}

// ADD_BASIC_ROLE attaches the default organisation role to the user.
func (u *User) ADD_BASIC_ROLE(roleStore *rbac.RbacStore) {
	role, err := roleStore.FindDefaultRole(u.CurrentOrganisationID)
	if err != nil {
		log.Println("Basic Role not found")
		return
	}
	u.Roles = append(u.Roles, *role)
}

// CREATE_AGENT_STATUS creates and persists an initial AgentStatus for the user.
func (u *User) CREATE_AGENT_STATUS(agentStore *agentops.StatusStore) error {
	id := uuid.New()
	agentStatus := &agentops.AgentStatus{
		ID:            id,
		AgentID:       u.ID,
		SerialNumber:  utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["agent_status"], id.String()),
		StatusChanges: []agentops.StatusChange{},
		OwnerType:     authctx.USERS,
		OwnerId:       u.ID,
	}
	u.AgentStatusID = id
	u.AgentStatus = *agentStatus
	if err := agentStore.Save(agentStatus); err != nil {
		return err
	}
	return nil
}
