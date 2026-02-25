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
type Branch struct {
	gorm.Model
	ID           uuid.UUID    `gorm:"primarykey;type:uuid;not null"`                                   // string id = 1;
	Name         string       `gorm:"type:text;"`                                                      // string name = 2;
	Description  string       `gorm:"type:text;"`                                                      // string description = 3;
	ParentID     uuid.UUID    `gorm:"type:uuid;default:00000000-0000-0000-0000-000000000000"`          // string parent_id = 5;
	Status       string       `gorm:"default:'ACTIVE'"`                                                // string status = 8;
	CreatedBy    string       `gorm:"type:text;"`                                                      // string created_by = 9;
	SerialNumber string       `gorm:"type:text;"`                                                      // string serial_number = 10;
	IsDefault    bool         `gorm:"default:false"`                                                   // bool is_default = 11;
	Users        []User       `gorm:"many2many:user_branches;"`                                        // repeated User users = 12;
	Roles        []Role       `gorm:"polymorphic:Owner;"`                                              // repeated Role roles = 13;
	Teams        []Team       `gorm:"polymorphic:Owner;"`                                              // repeated Team teams = 14;
	Departments  []Department `gorm:"polymorphic:Owner;"`                                              // repeated Department departments = 15;
	UpdatedBy    string       `gorm:"type:text;"`                                                      // string updated_by = 16;
	OwnerType    OwnerType    `gorm:"type:text;not null;default:0"`                                    // OwnerType ownerType = 17;
	OwnerID      uuid.UUID    `gorm:"type:uuid;not null;default:00000000-0000-0000-0000-000000000000"` // string owner_id = 18;
	Location     string       `gorm:"type:text;"`                                                      // string location = 19;
	Address      string       `gorm:"type:text;"`                                                      // string address = 19;
}

func (b *Branch) CREATE_BRANCH_ROLES(roleStore *RbacStore) error {
	adminId := uuid.New()
	supervisorId := uuid.New()
	agentId := uuid.New()
	roles := []Role{
		{
			ID:                  adminId,
			Name:                fmt.Sprintf("%s Admin", b.Name),
			Description:         "Branch Admin Role",
			PermissionCodeNames: k.ADMIN_PERMISSIONS(),
			CreatedBy:           "USERS",
			UpdatedAt:           time.Now(),
			CreatedAt:           time.Now(),
			IsDefault:           false,
			OwnerType:           BRANCHES,
			OwnerID:             b.ID,
		},
		{
			ID:                  supervisorId,
			Name:                fmt.Sprintf("%s Supervisor", b.Name),
			Description:         "Branch Supervisor Role",
			PermissionCodeNames: k.SUPERVISOR_PERMISSIONS(),
			CreatedBy:           "USERS",
			UpdatedAt:           time.Now(),
			CreatedAt:           time.Now(),
			IsDefault:           false,
			OwnerType:           BRANCHES,
			OwnerID:             b.ID,
		},
		{
			ID:                  agentId,
			Name:                fmt.Sprintf("%s Agent", b.Name),
			Description:         "Branch Agent Role",
			PermissionCodeNames: k.AGENT_PERMISSIONS(),
			CreatedBy:           "USERS",
			UpdatedAt:           time.Now(),
			CreatedAt:           time.Now(),
			IsDefault:           false,
			OwnerType:           BRANCHES,
			OwnerID:             b.ID,
		},
	}

	for _, role := range roles {
		role.SerialNumber = utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], role.ID.String())
		_, err := roleStore.SaveRole(&role)
		if err != nil {
			return err
		}
		b.Roles = append(b.Roles, role)
	}
	return nil
}
