package service

import (
	"fmt"
	"kyla-be/pkg/k"
	"kyla-be/pkg/utils"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Branch represents a branch entity
type Department struct {
	gorm.Model
	ID             uuid.UUID `gorm:"primarykey;type:uuid;not null"`
	SerialNumber   string
	DepartmentName string
	DepartmentBio  string
	Status         string
	CreatedBy      string
	UpdatedBy      string
	Users          []User `gorm:"many2many:user_departments;"`
	Roles          []Role `gorm:"polymorphic:Owner;"`
	Teams          []Team `gorm:"polymorphic:Owner;"`
	OwnerType      OwnerType
	OwnerID        uuid.UUID `gorm:"not null; default:00000000-0000-0000-0000-000000000000"`
}

func (d *Department) CREATE_DEPARTMENT_ROLES(roleStore *RbacStore) {
	adminId := uuid.New()
	supervisorId := uuid.New()
	agentId := uuid.New()
	roles := []Role{
		{
			ID:                  adminId,
			Name:                fmt.Sprintf("%s Admin", d.DepartmentName),
			Description:         "Department Admin Role",
			PermissionCodeNames: k.ADMIN_PERMISSIONS(),
			CreatedBy:           "USERS",
			UpdatedAt:           time.Now(),
			CreatedAt:           time.Now(),
			IsDefault:           false,
			OwnerType:           DEPARTMENTS,
			OwnerID:             d.ID,
		},
		{
			ID:                  supervisorId,
			Name:                fmt.Sprintf("%s Supervisor", d.DepartmentName),
			Description:         "Department Supervisor Role",
			PermissionCodeNames: k.SUPERVISOR_PERMISSIONS(),
			CreatedBy:           "USERS",
			UpdatedAt:           time.Now(),
			CreatedAt:           time.Now(),
			IsDefault:           false,
			OwnerType:           DEPARTMENTS,
			OwnerID:             d.ID,
		},
		{
			ID:                  agentId,
			Name:                fmt.Sprintf("%s Agent", d.DepartmentName),
			Description:         "Department Agent Role",
			PermissionCodeNames: k.AGENT_PERMISSIONS(),
			CreatedBy:           "USERS",
			UpdatedAt:           time.Now(),
			CreatedAt:           time.Now(),
			IsDefault:           false,
			OwnerType:           DEPARTMENTS,
			OwnerID:             d.ID,
		},
	}

	for _, role := range roles {
		role.SerialNumber = utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], role.ID.String())
		d.Roles = append(d.Roles, role)
	}
}

func (d *Department) ADD_CREATOR_TO_DEPARTMENT(userStore *UserStore) error {
	creatorID := uuid.MustParse(d.CreatedBy)
	user, err := userStore.FindByID(&creatorID)
	if err != nil {
		return err
	}
	d.Users = append(d.Users, *user)
	return nil
}
