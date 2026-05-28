package invitation

import (
	"kyla-be/pkg/pb"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// InvitationStatus represents the current status of an invitation
type InvitationStatus int32

const (
	InvitationStatusUnspecified InvitationStatus = iota
	InvitationStatusPending
	InvitationStatusAccepted
	InvitationStatusRejected
	InvitationStatusExpired
)

// InvitationType represents the type of invitation
type InvitationType int32

const (
	InvitationTypeUnspecified InvitationType = iota
	InvitationTypeNewUser
	InvitationTypeExistingUser
)

// Invitation represents a user invitation in the system
type Invitation struct {
	gorm.Model
	ID             uuid.UUID        `gorm:"column:id;primaryKey;type:uuid"`
	Email          string           `gorm:"column:email;not null"`
	InvitedBy      uuid.UUID        `gorm:"column:invited_by;not null"`
	Type           InvitationType   `gorm:"column:type;not null"`
	Status         InvitationStatus `gorm:"column:status;not null"`
	OrganisationID uuid.UUID        `gorm:"column:organisation_id;not null"`
	BranchID       uuid.UUID        `gorm:"column:branch_id"`
	DepartmentID   uuid.UUID        `gorm:"column:department_id"`
	TeamID         uuid.UUID        `gorm:"column:team_id"`
	UserID         uuid.UUID        `gorm:"column:user_id"`
	RoleIDs        pq.StringArray   `gorm:"column:role_ids;type:text[];not null"`
	Token          string           `gorm:"column:token;unique;not null"`
	ExpiresAt      time.Time        `gorm:"column:expires_at;not null"`
}

// TableName specifies the table name for GORM
func (Invitation) TableName() string {
	return "invitations"
}

// BeforeCreate is a GORM hook that runs before creating a new record
func (i *Invitation) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	if i.Token == "" {
		i.Token = uuid.New().String()
	}

	// If no roles are specified, assign default roles based on available entity IDs
	if len(i.RoleIDs) == 0 {
		roleIDs := []string{}

		// Map of owner types to their corresponding IDs
		ownerMap := map[string]uuid.UUID{
			"organisations": i.OrganisationID,
			"branches":      i.BranchID,
			"departments":   i.DepartmentID,
			"teams":         i.TeamID,
		}

		// Query roles for each available owner type and ID
		for ownerType, ownerID := range ownerMap {
			if ownerID != uuid.Nil {
				var roles []string
				if err := tx.Table("roles").
					Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).
					Pluck("id", &roles).Error; err == nil {
					roleIDs = append(roleIDs, roles...)
				}
			}
		}

		// Assign found roles to the invitation if any were found
		if len(roleIDs) > 0 {
			i.RoleIDs = roleIDs
		}
	}
	return nil
}

// NewInvitation creates a new invitation
func NewInvitation(
	email string,
	invitedBy uuid.UUID,
	invitationType InvitationType,
	organisationID uuid.UUID,
	branchID uuid.UUID,
	departmentID uuid.UUID,
	teamID uuid.UUID,
	userID uuid.UUID,
	roleIDs []string,
	expirationHours int32,
) *Invitation {
	now := time.Now()
	expiresAt := now.Add(time.Duration(expirationHours) * time.Hour)

	return &Invitation{
		Email:          email,
		InvitedBy:      invitedBy,
		Type:           invitationType,
		Status:         InvitationStatusPending,
		OrganisationID: organisationID,
		BranchID:       branchID,
		DepartmentID:   departmentID,
		TeamID:         teamID,
		UserID:         userID,
		RoleIDs:        roleIDs,
		ExpiresAt:      expiresAt,
	}
}

// ToProto converts the invitation to its protobuf representation
func (i *Invitation) ToProto() *pb.Invitation {
	invitation := &pb.Invitation{
		Id:             i.ID.String(),
		Email:          i.Email,
		InvitedBy:      i.InvitedBy.String(),
		Type:           pb.InvitationType(i.Type),
		Status:         pb.InvitationStatus(i.Status),
		OrganisationId: i.OrganisationID.String(),
		BranchId:       i.BranchID.String(),
		DepartmentId:   i.DepartmentID.String(),
		TeamId:         i.TeamID.String(),
		RoleIds:        i.RoleIDs,
		CreatedAt:      timestamppb.New(i.CreatedAt),
		UpdatedAt:      timestamppb.New(i.UpdatedAt),
		ExpiresAt:      timestamppb.New(i.ExpiresAt),
	}

	if i.BranchID != uuid.Nil {
		invitation.BranchId = i.BranchID.String()
	}
	if i.DepartmentID != uuid.Nil {
		invitation.DepartmentId = i.DepartmentID.String()
	}
	if i.TeamID != uuid.Nil {
		invitation.TeamId = i.TeamID.String()
	}
	if i.UserID != uuid.Nil {
		invitation.UserId = i.UserID.String()
	}

	return invitation
}

// FromProto creates an invitation from its protobuf representation
func FromProto(inv *pb.Invitation) *Invitation {
	roles := pq.StringArray{}
	if inv.RoleIds != nil {
		roles = append(roles, inv.RoleIds...)
	}
	// Ensure the ID is parsed correctly
	if inv.Id == "" {
		inv.Id = uuid.New().String()
	}
	invitedBy, err := uuid.Parse(inv.InvitedBy)
	if err != nil {
		invitedBy = uuid.Nil // Default to nil if parsing fails
	}
	orgID, err := uuid.Parse(inv.OrganisationId)
	if err != nil {
		orgID = uuid.Nil // Default to nil if parsing fails
	}

	i := &Invitation{
		ID:             uuid.MustParse(inv.Id),
		Email:          inv.Email,
		InvitedBy:      invitedBy,
		Type:           InvitationType(inv.Type),
		Status:         InvitationStatus(inv.Status),
		OrganisationID: orgID,
		RoleIDs:        roles,
		ExpiresAt:      inv.ExpiresAt.AsTime(),
	}

	if inv.BranchId != "" {
		branchID, err := uuid.Parse(inv.BranchId)
		if err != nil {
			branchID = uuid.Nil
		}
		i.BranchID = branchID
	}
	if inv.DepartmentId != "" {
		departmentID, err := uuid.Parse(inv.DepartmentId)
		if err != nil {
			departmentID = uuid.Nil
		}
		i.DepartmentID = departmentID
	}
	if inv.TeamId != "" {
		teamID, err := uuid.Parse(inv.TeamId)
		if err != nil {
			teamID = uuid.Nil
		}
		i.TeamID = teamID
	}
	if inv.UserId != "" {
		userID, err := uuid.Parse(inv.UserId)
		if err != nil {
			userID = uuid.Nil
		}
		i.UserID = userID
	}

	return i
}

// IsExpired checks if the invitation has expired
func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// Accept marks the invitation as accepted
func (i *Invitation) Accept() {
	i.Status = InvitationStatusAccepted
	i.UpdatedAt = time.Now()
}

// Reject marks the invitation as rejected
func (i *Invitation) Reject() {
	i.Status = InvitationStatusRejected
	i.UpdatedAt = time.Now()
}

// Cancel marks the invitation as cancelled
func (i *Invitation) Cancel() {
	i.Status = InvitationStatusRejected
	i.UpdatedAt = time.Now()
}

// MarkAsExpired marks the invitation as expired
func (i *Invitation) MarkAsExpired() {
	i.Status = InvitationStatusExpired
	i.UpdatedAt = time.Now()
}
