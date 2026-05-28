package organisation

import (
	"fmt"
	"time"

	"kyla-be/internal/agentops"
	"kyla-be/internal/authctx"
	"kyla-be/internal/branch"
	"kyla-be/internal/rbac"
	"kyla-be/internal/user"
	"kyla-be/pkg/k"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Organisation is the top-level tenant entity.
type Organisation struct {
	gorm.Model
	ID                    uuid.UUID        `gorm:"primarykey;type:uuid;not null"`
	SerialNumber          string
	OrganisationName      string
	OrganisationBio       string
	Size                  string
	Country               string
	CountryCode           string
	Industry              string
	SubIndustry           string
	Status                string           `gorm:"default:'ACTIVE'"`
	ReferralCode          string
	ShortCode             string
	Email                 string
	Phone                 string
	CreatedBy             string
	DialerIntegrationLink string
	CrmRedirectLink       string
	RedirectFallbackLink  string
	Users                 []user.User      `gorm:"many2many:organisation_users;"`
	Branches              []branch.Branch  `gorm:"polymorphic:Owner;"`
	Roles                 []rbac.Role      `gorm:"polymorphic:Owner;"`
}

// OrgSetupArgs bundles the dependencies required by SPIN_UP_ORGANISATION.
type OrgSetupArgs struct {
	DB                *gorm.DB
	OrganisationStore *OrganisationStore
	RoleStore         *rbac.RbacStore
	UserStore         *user.UserStore
	AgentStore        *agentops.StatusStore
	Organisation      *Organisation
}

// SPIN_UP_ORGANISATION bootstraps a new organisation with a default branch,
// role set, and assigns the founding user appropriately.
func (o *Organisation) SPIN_UP_ORGANISATION(
	orgSetupArgs *OrgSetupArgs,
	u *user.User,
) (*Organisation, error) {
	org := orgSetupArgs.Organisation

	branchID := uuid.New()
	adminId := uuid.New()
	supervisorId := uuid.New()
	agentId := uuid.New()
	org.Status = k.GENERAL_STATUSES()["ACTIVE"]
	org.Branches = append(org.Branches, branch.Branch{
		ID:           branchID,
		SerialNumber: utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["branches"], branchID.String()),
		Name:         "Head Office",
		Description:  "Default Branch",
		CreatedBy:    "USERS",
		Status:       "ACTIVE",
		ParentID:     uuid.Nil,
		IsDefault:    true,
		OwnerType:    authctx.ORGANISATIONS,
		OwnerID:      org.ID,
	})
	// Branch roles saved separately since branch.Branch has no Roles slice (import cycle prevention).
	branchRoles := []rbac.Role{
		{
			ID:                  adminId,
			SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], adminId.String()),
			Name:                fmt.Sprintf("%s Admin", "Head Office"),
			Description:         "Branch Admin Role",
			PermissionCodeNames: k.ADMIN_PERMISSIONS(),
			CreatedBy:           "USERS",
			UpdatedAt:           time.Now(),
			CreatedAt:           time.Now(),
			IsDefault:           false,
			OwnerType:           authctx.BRANCHES,
			OwnerID:             branchID,
		},
		{
			ID:                  supervisorId,
			SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], supervisorId.String()),
			Name:                fmt.Sprintf("%s Supervisor", "Head Office"),
			Description:         "Branch Supervisor Role",
			PermissionCodeNames: k.SUPERVISOR_PERMISSIONS(),
			CreatedBy:           "USERS",
			UpdatedAt:           time.Now(),
			CreatedAt:           time.Now(),
			IsDefault:           false,
			OwnerType:           authctx.BRANCHES,
			OwnerID:             branchID,
		},
		{
			ID:                  agentId,
			SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], agentId.String()),
			Name:                fmt.Sprintf("%s Agent", "Head Office"),
			Description:         "Branch Agent Role",
			PermissionCodeNames: k.AGENT_PERMISSIONS(),
			CreatedBy:           "USERS",
			UpdatedAt:           time.Now(),
			CreatedAt:           time.Now(),
			IsDefault:           false,
			OwnerType:           authctx.BRANCHES,
			OwnerID:             branchID,
		},
	}
	org.Users = append(org.Users, *u)

	roleID := uuid.New()
	org.Roles = append(org.Roles, rbac.Role{
		ID:                  roleID,
		SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], roleID.String()),
		Name:                "Basic Role",
		Description:         "Basic Role for organisation users",
		PermissionCodeNames: k.BASIC_PERMISSIONS(),
		CreatedBy:           "USERS",
		UpdatedAt:           time.Now(),
		CreatedAt:           time.Now(),
		IsDefault:           true,
		OwnerType:           authctx.ORGANISATIONS,
		OwnerID:             org.ID,
	})

	roles := []rbac.Role{}
	roles = append(roles, org.Roles...)
	roles = append(roles, branchRoles...)

	newOrg, err := orgSetupArgs.OrganisationStore.Save(org)
	if err != nil {
		return nil, err
	}

	// Save branch roles explicitly (branch.Branch has no Roles slice).
	for i := range branchRoles {
		if _, err := orgSetupArgs.RoleStore.SaveRole(&branchRoles[i]); err != nil {
			return nil, fmt.Errorf("failed to save branch role: %v", err)
		}
	}

	// Assign roles and org to user (branch many2many handled via raw DB association
	// since user.User no longer carries a Branches slice).
	u.Roles = append(u.Roles, roles...)
	u.CurrentBranchID = branchID
	u.CurrentOrganisationID = org.ID
	u.ADD_BASIC_ROLE(orgSetupArgs.RoleStore)

	if _, err := orgSetupArgs.UserStore.Save(u, false); err != nil {
		return nil, err
	}

	// Associate user with the default branch via join table.
	if orgSetupArgs.DB != nil {
		if err := orgSetupArgs.DB.Exec(
			"INSERT INTO user_branches (user_id, branch_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
			u.ID, branchID,
		).Error; err != nil {
			return nil, fmt.Errorf("failed to associate user with default branch: %v", err)
		}
	}

	return newOrg, nil
}

// ADD_BASIC_ROLE appends a default Basic Role to the organisation's role list.
func (o *Organisation) ADD_BASIC_ROLE() {
	roleID := uuid.New()
	o.Roles = append(o.Roles, rbac.Role{
		ID:                  roleID,
		SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], roleID.String()),
		Name:                "Basic Role",
		Description:         "Basic Role for organisation users",
		PermissionCodeNames: k.BASIC_PERMISSIONS(),
		CreatedBy:           "USERS",
		UpdatedAt:           time.Now(),
		CreatedAt:           time.Now(),
		IsDefault:           true,
		OwnerType:           authctx.ORGANISATIONS,
		OwnerID:             o.ID,
	})
}
