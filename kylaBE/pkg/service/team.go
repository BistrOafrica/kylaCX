package service

import (
	"fmt"
	"kyla-be/pkg/k"
	"kyla-be/pkg/utils"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Team struct {
	gorm.Model
	ID           uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	SerialNumber string    `gorm:"unique"`
	Name         string    `gorm:"unique"`
	Description  string
	CreatedBy    string
	UpdatedBy    string
	Users        []User    `gorm:"many2many:user_teams;"`
	Roles        []Role    `gorm:"polymorphic:Owner;"`
	OwnerID      uuid.UUID `gorm:"type:uuid;not null; default:00000000-0000-0000-0000-000000000000"`
	OwnerType    OwnerType
}

func (t *Team) CREATE_TEAM_ROLES(roleStore *RbacStore) {
	adminId := uuid.New()
	supervisorId := uuid.New()
	agentId := uuid.New()
	roles := []Role{
		{
			ID:                  adminId,
			Name:                fmt.Sprintf("%s Admin", t.Name),
			Description:         "Team Admin Role",
			PermissionCodeNames: k.ADMIN_PERMISSIONS(),
			CreatedBy:           "USERS",
			UpdatedAt:           time.Now(),
			CreatedAt:           time.Now(),
			IsDefault:           false,
			OwnerType:           TEAMS,
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
			OwnerType:           TEAMS,
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
			OwnerType:           TEAMS,
			OwnerID:             t.ID,
		},
	}

	for _, role := range roles {
		role.SerialNumber = utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], role.ID.String())
		t.Roles = append(t.Roles, role)
	}
}

func (t *Team) ADD_CREATOR_TO_TEAM(userStore *UserStore) {
	creatorID := uuid.MustParse(t.CreatedBy)
	user, err := userStore.FindByID(&creatorID)
	if err != nil {
		log.Println("User not found")
	}
	user.Roles = append(user.Roles, t.Roles...)
	t.Users = append(t.Users, *user)
}
