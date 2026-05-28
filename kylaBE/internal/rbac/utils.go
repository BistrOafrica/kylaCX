package rbac

import (
	"kyla-be/internal/authctx"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// --- Converter functions ---

// RoleToPbRole converts a Role model to its protobuf representation.
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

// PbRoleToRole converts a protobuf Role to a Role model.
func PbRoleToRole(role *pb.Role) *Role {
	id, err := uuid.Parse(role.GetId())
	if err != nil {
		id = uuid.New()
	}
	ownerID, err := uuid.Parse(role.GetOwnerId())
	if err != nil {
		ownerID = uuid.Nil
	}
	return &Role{
		ID:                  id,
		Name:                role.Name,
		Description:         role.Description,
		PermissionCodeNames: pq.StringArray(role.Permissions),
		CreatedBy:           role.CreatedBy,
		UpdatedBy:           role.UpdatedBy,
		SerialNumber:        role.SerialNumber,
		IsDefault:           role.IsDefault,
		OwnerType:           authctx.OwnerType(role.OwnerType),
		OwnerID:             ownerID,
	}
}

// RolesToPbRoles converts a slice of Role pointers to protobuf Role pointers.
func RolesToPbRoles(roles []*Role) []*pb.Role {
	var rs []*pb.Role
	for _, role := range roles {
		rs = append(rs, RoleToPbRole(role))
	}
	return rs
}

// PbRolesToRoles converts a slice of protobuf Role pointers to Role model pointers.
func PbRolesToRoles(roles []*pb.Role) []*Role {
	var rs []*Role
	for _, role := range roles {
		rs = append(rs, PbRoleToRole(role))
	}
	return rs
}

// PermissionToPbPermission converts a k.Permission to its protobuf representation.
func PermissionToPbPermission(permission k.Permission) *pb.Permission {
	return &pb.Permission{
		Name:        permission.Name,
		Service:     permission.Service,
		Description: permission.Description,
		CodeName:    permission.CodeName,
		Console:     permission.Console,
	}
}

// PbPermissionToPermission converts a protobuf Permission to a k.Permission.
func PbPermissionToPermission(permission *pb.Permission) *k.Permission {
	return &k.Permission{
		Name:        permission.Name,
		Service:     permission.Service,
		Description: permission.Description,
		CodeName:    permission.CodeName,
		Console:     permission.Console,
	}
}

// PermissionsToPbPermissions converts a slice of k.Permission to protobuf.
func PermissionsToPbPermissions(permissions []k.Permission) []*pb.Permission {
	var perms []*pb.Permission
	for _, permission := range permissions {
		perms = append(perms, PermissionToPbPermission(permission))
	}
	return perms
}

// ConvertMapToPbPermissions converts a map of k.Permission values to a protobuf slice.
func ConvertMapToPbPermissions(permissions map[string]k.Permission) []*pb.Permission {
	var perms []*pb.Permission
	for _, permission := range permissions {
		perms = append(perms, PermissionToPbPermission(permission))
	}
	return perms
}

// --- NameExtractor helpers ---
// These minimal proxy structs carry only the fields needed for name extraction.
// Their TableName() methods ensure GORM queries the correct table.

type nameProxyUser struct {
	ID        uuid.UUID `gorm:"primarykey"`
	FirstName string
	LastName  string
}

func (nameProxyUser) TableName() string { return "users" }

type nameProxyTeam struct {
	ID   uuid.UUID `gorm:"primarykey"`
	Name string
}

func (nameProxyTeam) TableName() string { return "teams" }

type nameProxyBranch struct {
	ID   uuid.UUID `gorm:"primarykey"`
	Name string
}

func (nameProxyBranch) TableName() string { return "branches" }

type nameProxyDepartment struct {
	ID             uuid.UUID `gorm:"primarykey"`
	DepartmentName string
}

func (nameProxyDepartment) TableName() string { return "departments" }

type nameProxyOrganisation struct {
	ID               uuid.UUID `gorm:"primarykey"`
	OrganisationName string
}

func (nameProxyOrganisation) TableName() string { return "organisations" }

// Pre-built NameExtractor instances for common domain types.
var (
	UserNameExtractor = func(model interface{}) string {
		u := model.(*nameProxyUser)
		return strings.TrimSpace(u.FirstName + " " + u.LastName)
	}

	TeamNameExtractor = func(model interface{}) string {
		return model.(*nameProxyTeam).Name
	}

	BranchNameExtractor = func(model interface{}) string {
		return model.(*nameProxyBranch).Name
	}

	DepartmentNameExtractor = func(model interface{}) string {
		return model.(*nameProxyDepartment).DepartmentName
	}

	OrganisationNameExtractor = func(model interface{}) string {
		return model.(*nameProxyOrganisation).OrganisationName
	}
)

// ProxyUser returns an empty nameProxyUser for use with FindDynamicIdNameMapping.
func ProxyUser() *nameProxyUser { return &nameProxyUser{} }

// ProxyTeam returns an empty nameProxyTeam for use with FindDynamicIdNameMapping.
func ProxyTeam() *nameProxyTeam { return &nameProxyTeam{} }

// ProxyBranch returns an empty nameProxyBranch for use with FindDynamicIdNameMapping.
func ProxyBranch() *nameProxyBranch { return &nameProxyBranch{} }

// ProxyDepartment returns an empty nameProxyDepartment for use with FindDynamicIdNameMapping.
func ProxyDepartment() *nameProxyDepartment { return &nameProxyDepartment{} }

// ProxyOrganisation returns an empty nameProxyOrganisation for use with FindDynamicIdNameMapping.
func ProxyOrganisation() *nameProxyOrganisation { return &nameProxyOrganisation{} }
