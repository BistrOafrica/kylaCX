package service

import (
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"strings"

	"github.com/google/uuid"
)

func RoleToPbRole(role *Role) *pb.Role {
	return &pb.Role{
		Id:           role.ID.String(),
		Name:         role.Name,
		Description:  role.Description,
		Permissions:  role.PermissionCodeNames,
		CreatedBy:    role.CreatedBy,
		UpdatedBy:    role.UpdatedBy,
		UpdatedAt:    role.UpdatedAt.String(),
		CreatedAt:    role.CreatedAt.String(),
		SerialNumber: role.SerialNumber,
		OwnerType:    pb.OwnerType(pb.OwnerType_value[string(role.OwnerType)]),
		OwnerId:      role.OwnerID.String(),
	}
}

func PbRoleToRole(role *pb.Role) *Role {
	id, err := uuid.Parse(role.GetId())
	if err != nil {
		id = uuid.New()
	}
	ownerId, err := uuid.Parse(role.GetOwnerId())
	if err != nil {
		ownerId = uuid.Nil
	}
	return &Role{
		ID:                  id,
		Name:                role.Name,
		Description:         role.Description,
		PermissionCodeNames: role.Permissions,
		CreatedBy:           role.CreatedBy,
		UpdatedBy:           role.UpdatedBy,
		SerialNumber:        role.SerialNumber,
		IsDefault:           role.IsDefault,
		OwnerType:           OwnerType(role.OwnerType),
		OwnerID:             ownerId,
	}
}

func ConvertUsers(users []User) []*User {
	ptrUsers := make([]*User, len(users))
	for i, user := range users {
		ptrUsers[i] = &user
	}
	return ptrUsers
}

func PermissionToPbPermission(permission k.Permission) *pb.Permission {
	return &pb.Permission{
		Name:        permission.Name,
		Service:     permission.Service,
		Description: permission.Description,
		CodeName:    permission.CodeName,
		Console:     permission.Console,
	}
}

func PbPermissionToPermission(permission *pb.Permission) *k.Permission {
	return &k.Permission{
		Name:        permission.Name,
		Service:     permission.Service,
		Description: permission.Description,
		CodeName:    permission.CodeName,
		Console:     permission.Console,
	}
}

func PbPermissionsToPermissions(permissions []*pb.Permission) []k.Permission {
	var perms []k.Permission
	for _, permission := range permissions {
		perms = append(perms, *PbPermissionToPermission(permission))
	}
	return perms
}

func PbRolesToRoles(roles []*pb.Role) []*Role {
	var rs []*Role
	for _, role := range roles {
		rs = append(rs, PbRoleToRole(role))
	}
	return rs
}

func RolesToPbRoles(roles []*Role) []*pb.Role {
	var rs []*pb.Role
	for _, role := range roles {
		rs = append(rs, RoleToPbRole(role))
	}
	return rs
}

func PermissionsToPbPermissions(permissions []k.Permission) []*pb.Permission {
	var perms []*pb.Permission
	for _, permission := range permissions {
		perms = append(perms, PermissionToPbPermission(permission))
	}
	return perms
}

func ConvertMapToPbPermissions(permissions map[string]k.Permission) []*pb.Permission {
	var perms []*pb.Permission
	for _, permission := range permissions {
		perms = append(perms, PermissionToPbPermission(permission))
	}
	return perms
}

// NameExtractor defines how to get the display name from a model
type NameExtractor func(interface{}) string

// Extractors for different model types
var (
	UserNameExtractor = func(model interface{}) string {
		u := model.(*User)
		return strings.TrimSpace(u.FirstName + " " + u.LastName)
	}

	TeamNameExtractor = func(model interface{}) string {
		return model.(*Team).Name
	}

	BranchNameExtractor = func(model interface{}) string {
		return model.(*Branch).Name
	}

	DepartmentNameExtractor = func(model interface{}) string {
		return model.(*Department).DepartmentName
	}

	OrganisationNameExtractor = func(model interface{}) string {
		return model.(*Organisation).OrganisationName
	}
)
