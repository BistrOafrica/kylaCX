package department

import (
	"fmt"
	"kyla-be/internal/authctx"
	"kyla-be/internal/rbac"
	"kyla-be/internal/team"
	"kyla-be/pkg/k"
	"kyla-be/pkg/service"
	"kyla-be/pkg/utils"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRef is a minimal GORM proxy for the users table used to express M2M
// department membership without importing the user package (which would create
// an import cycle because the user model references Department).
type UserRef struct {
	ID uuid.UUID `gorm:"type:uuid;primarykey"`
}

// TableName ensures GORM maps UserRef to the existing users table.
func (UserRef) TableName() string { return "users" }

// Department represents a department entity.
type Department struct {
	gorm.Model
	ID             uuid.UUID         `gorm:"primarykey;type:uuid;not null"`
	SerialNumber   string
	DepartmentName string
	DepartmentBio  string
	Status         string
	CreatedBy      string
	UpdatedBy      string
	Users          []UserRef         `gorm:"many2many:user_departments;"`
	Roles          []rbac.Role       `gorm:"polymorphic:Owner;"`
	Teams          []team.Team       `gorm:"polymorphic:Owner;"`
	OwnerType      authctx.OwnerType
	OwnerID        uuid.UUID         `gorm:"not null; default:00000000-0000-0000-0000-000000000000"`
}

func (d *Department) CREATE_DEPARTMENT_ROLES(roleStore *rbac.RbacStore) {
	adminId := uuid.New()
	supervisorId := uuid.New()
	agentId := uuid.New()
	roles := []rbac.Role{
		{
			ID:                  adminId,
			Name:                fmt.Sprintf("%s Admin", d.DepartmentName),
			Description:         "Department Admin Role",
			PermissionCodeNames: k.ADMIN_PERMISSIONS(),
			CreatedBy:           "USERS",
			UpdatedAt:           time.Now(),
			CreatedAt:           time.Now(),
			IsDefault:           false,
			OwnerType:           authctx.DEPARTMENTS,
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
			OwnerType:           authctx.DEPARTMENTS,
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
			OwnerType:           authctx.DEPARTMENTS,
			OwnerID:             d.ID,
		},
	}

	for _, role := range roles {
		role.SerialNumber = utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], role.ID.String())
		d.Roles = append(d.Roles, role)
	}
}

// ADD_CREATOR_TO_DEPARTMENT adds the department creator as a member.
// Simplified from the original to avoid importing the full User type which
// would create an import cycle.
func (d *Department) ADD_CREATOR_TO_DEPARTMENT(userStore *service.UserStore) error {
	creatorID := uuid.MustParse(d.CreatedBy)
	_, err := userStore.FindByID(&creatorID)
	if err != nil {
		return err
	}
	d.Users = append(d.Users, UserRef{ID: creatorID})
	return nil
}
