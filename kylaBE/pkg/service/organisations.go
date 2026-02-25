package service

import (
	"fmt"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OwnerType string

const (
	USERS         OwnerType = "USERS"
	TEAMS         OwnerType = "TEAMS"
	BRANCHES      OwnerType = "BRANCHES"
	DEPARTMENTS   OwnerType = "DEPARTMENTS"
	ORGANISATIONS OwnerType = "ORGANISATIONS"
)

func (o OwnerType) String() string {
	return string(o)
}

type OpScope struct {
	Owner OwnerType
	ID    string
}

type OrgSetupArgs struct {
	OrganisationStore *OrganisationStore
	BranchStore       *BranchStore
	RoleStore         *RbacStore
	UserStore         *UserStore
	AgentStore        *StatusStore
	Organisation      *Organisation
	PbOrg             *pb.Organisation
	EmailService      *utils.ResendService
}

type Organisation struct {
	gorm.Model
	ID                    uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	SerialNumber          string
	OrganisationName      string
	OrganisationBio       string
	Size                  string
	Country               string
	CountryCode           string
	Industry              string
	SubIndustry           string
	Status                string `gorm:"default:'ACTIVE'"`
	ReferralCode          string
	ShortCode             string
	Email                 string
	Phone                 string
	CreatedBy             string
	DialerIntegrationLink string
	CrmRedirectLink       string
	RedirectFallbackLink  string
	Users                 []User       `gorm:"many2many:organisation_users;"`
	Departments           []Department `gorm:"polymorphic:Owner;"`
	Branches              []Branch     `gorm:"polymorphic:Owner;"`
	Teams                 []Team       `gorm:"polymorphic:Owner;"`
	Roles                 []Role       `gorm:"polymorphic:Owner;"`
}

func (o *Organisation) SPIN_UP_ORGANISATION(
	orgSetupArgs *OrgSetupArgs,
	user *User,
) (*Organisation, error) {

	org := orgSetupArgs.Organisation

	branchID := uuid.New()
	adminId := uuid.New()
	supervisorId := uuid.New()
	agentId := uuid.New()
	org.Status = k.GENERAL_STATUSES()["ACTIVE"]
	org.Branches = append(org.Branches, Branch{
		ID:           branchID,
		SerialNumber: utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["branches"], branchID.String()),
		Name:         "Head Office",
		Description:  "Default Branch",
		CreatedBy:    "USERS",
		Status:       "ACTIVE",
		ParentID:     uuid.Nil,
		IsDefault:    true,
		OwnerType:    OwnerType(ORGANISATIONS),
		OwnerID:      org.ID,
		Users:        []User{*user},
		Roles: []Role{
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
				OwnerType:           OwnerType(BRANCHES),
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
				OwnerType:           OwnerType(BRANCHES),
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
				OwnerType:           OwnerType(BRANCHES),
				OwnerID:             branchID,
			},
		},
	})
	org.Users = append(org.Users, *user)
	roleID := uuid.New()
	org.Roles = append(org.Roles, Role{
		ID:                  roleID,
		SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], roleID.String()),
		Name:                "Basic Role",
		Description:         "Basic Role for organisation users",
		PermissionCodeNames: k.BASIC_PERMISSIONS(),
		CreatedBy:           "USERS",
		UpdatedAt:           time.Now(),
		CreatedAt:           time.Now(),
		IsDefault:           true,
		OwnerType:           OwnerType(ORGANISATIONS),
		OwnerID:             org.ID,
	})

	roles := []Role{}
	roles = append(roles, org.Roles...)
	roles = append(roles, org.Branches[0].Roles...)
	newOrg, err := orgSetupArgs.OrganisationStore.Save(org)
	if err != nil {
		return nil, err
	}

	user.Roles = append(user.Roles, roles...)
	user.Branches = append(user.Branches, org.Branches...)
	user.Organisations = append(user.Organisations, *org)
	user.CurrentBranchID = branchID
	user.CurrentOrganisationID = org.ID
	user.ADD_BASIC_ROLE(orgSetupArgs.RoleStore)
	if _, err := orgSetupArgs.UserStore.Save(user, false); err != nil {
		return nil, err
	}

	return newOrg, nil
}

func (o *Organisation) ADD_BASIC_ROLE() {
	roleID := uuid.New()
	o.Roles = append(o.Roles, Role{
		ID:                  roleID,
		SerialNumber:        utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], roleID.String()),
		Name:                "Basic Role",
		Description:         "Basic Role for organisation users",
		PermissionCodeNames: k.BASIC_PERMISSIONS(),
		CreatedBy:           "USERS",
		UpdatedAt:           time.Now(),
		CreatedAt:           time.Now(),
		IsDefault:           true,
		OwnerType:           OwnerType(ORGANISATIONS),
		OwnerID:             o.ID,
	})
}
