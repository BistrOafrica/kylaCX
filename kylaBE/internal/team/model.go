package team

import (
	"fmt"
	"kyla-be/internal/authctx"
	"kyla-be/internal/rbac"
	"kyla-be/pkg/k"
	"kyla-be/pkg/service"
	"kyla-be/pkg/utils"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRef is a minimal GORM proxy for the users table used to express M2M
// team membership without importing the user package (which would create
// an import cycle because the user model references Team).
type UserRef struct {
	ID uuid.UUID `gorm:"type:uuid;primarykey"`
}

// TableName ensures GORM maps UserRef to the existing users table.
func (UserRef) TableName() string { return "users" }

// Team represents a team entity.
type Team struct {
	gorm.Model
	ID           uuid.UUID         `gorm:"primarykey;type:uuid;not null"`
	SerialNumber string            `gorm:"unique"`
	Name         string            `gorm:"unique"`
	Description  string
	CreatedBy    string
	UpdatedBy    string
	Users        []UserRef         `gorm:"many2many:user_teams;"`
	Roles        []rbac.Role       `gorm:"polymorphic:Owner;"`
	OwnerID      uuid.UUID         `gorm:"type:uuid;not null; default:00000000-0000-0000-0000-000000000000"`
	OwnerType    authctx.OwnerType
}

func (t *Team) CREATE_TEAM_ROLES(roleStore *rbac.RbacStore) {
	adminId := uuid.New()
	supervisorId := uuid.New()
	agentId := uuid.New()
	roles := []rbac.Role{
		{
			ID:                  adminId,
			Name:                fmt.Sprintf("%s Admin", t.Name),
			Description:         "Team Admin Role",
			PermissionCodeNames: k.ADMIN_PERMISSIONS(),
			CreatedBy:           "USERS",
			UpdatedAt:           time.Now(),
			CreatedAt:           time.Now(),
			IsDefault:           false,
			OwnerType:           authctx.TEAMS,
			OwnerID:             t.ID,
		},
		{
			ID:                  supervisorId,
			Name:                fmt.Sprintf("%s Supervisor", t.Name),
			Description:         "Team Supervisor Role",
			PermissionCodeNames: k.SUPERVISOR_PERMISSIONS(),
			CreatedBy:           "USERS",
			UpdatedAt:           time.Now(),
			CreatedAt:           time.Now(),
			IsDefault:           false,
			OwnerType:           authctx.TEAMS,
			OwnerID:             t.ID,
		},
		{
			ID:                  agentId,
			Name:                fmt.Sprintf("%s Agent", t.Name),
			Description:         "Team Agent Role",
			PermissionCodeNames: k.AGENT_PERMISSIONS(),
			CreatedBy:           "USERS",
			UpdatedAt:           time.Now(),
			CreatedAt:           time.Now(),
			IsDefault:           false,
			OwnerType:           authctx.TEAMS,
			OwnerID:             t.ID,
		},
	}

	for _, role := range roles {
		role.SerialNumber = utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], role.ID.String())
		t.Roles = append(t.Roles, role)
	}
}

// ADD_CREATOR_TO_TEAM adds the team creator as a member.
// Note: the original logic also pre-assigned team roles to the user object
// in memory; that side-effect is omitted here because rbac.Role and
// service.Role are now distinct types. The creator is still recorded as a
// team member via the user_teams join table.
func (t *Team) ADD_CREATOR_TO_TEAM(userStore *service.UserStore) {
	creatorID := uuid.MustParse(t.CreatedBy)
	_, err := userStore.FindByID(&creatorID)
	if err != nil {
		log.Println("User not found")
		return
	}
	t.Users = append(t.Users, UserRef{ID: creatorID})
}
