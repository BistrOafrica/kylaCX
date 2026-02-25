package service

import (
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
)

func ReadPermissionsFromRoles(roles []*Role) []string {
	permissionSet := make(map[string]struct{})
	var permissions []string
	for _, role := range roles {
		for _, permission := range role.PermissionCodeNames {
			if _, exists := permissionSet[permission]; !exists {
				permissions = append(permissions, permission)
				permissionSet[permission] = struct{}{}
			}
		}
	}

	return permissions
}

func CheckOpScope(metadata *RequestMetadata, opScope *OpScope) bool {
	userScopes := &Scopes{
		User:         metadata.UserID,
		Teams:        metadata.Scopes.Teams,
		Departments:  metadata.Scopes.Departments,
		Branch:       metadata.Scopes.Branch,
		Organisation: metadata.Scopes.Organisation,
	}

	switch opScope.Owner {
	case OwnerType(TEAMS):
		if utils.Includes(userScopes.Teams, opScope.ID) {
			return true
		}
	case OwnerType(USERS):
		if userScopes.User.String() == opScope.ID {
			return true
		}
	case OwnerType(BRANCHES):
		if userScopes.Branch.String() == opScope.ID {
			return true
		}
	case OwnerType(DEPARTMENTS):
		if utils.Includes(userScopes.Departments, opScope.ID) {
			return true
		}
	case OwnerType(ORGANISATIONS):
		if userScopes.Organisation.String() == opScope.ID {
			return true
		}
	default:
		return false
	}
	return false
}

func GetScopeIDs(scope *Scopes) []string {
	ids := []string{}
	ids = append(ids, scope.Teams...)
	ids = append(ids, scope.Departments...)
	if scope.User != uuid.Nil {
		ids = append(ids, scope.User.String())
	}
	if scope.Branch != uuid.Nil {
		ids = append(ids, scope.Branch.String())
	}
	if scope.Organisation != uuid.Nil {
		ids = append(ids, scope.Organisation.String())
	}
	return ids
}

func CheckIfIDInScope(scope *Scopes, id string) bool {
	ids := GetScopeIDs(scope)
	for _, i := range ids {
		if i == id {
			return true
		}
	}
	return false
}

func SessionToPbUserSession(session *UserSession) *pb.UserSession {
	if session == nil {
		return nil
	}
	return &pb.UserSession{
		Id:        session.ID.String(),
		UserId:    session.UserID.String(),
		StartTime: session.StartTime,
		EndTime:   session.EndTime,
		IsValid:   session.IsValid,
	}
}

func DeviceToPbUserDevice(device *UserDeviceInfo) *pb.UserDevice {
	if device == nil {
		return nil
	}
	return &pb.UserDevice{
		Id:          device.ID.String(),
		UserId:      device.UserID.String(),
		DeviceMacId: device.DeviceMacID,
		DeviceType:  device.DeviceType,
		OsType:      device.OSType,
		DeviceName:  device.DeviceName,
		UserAgent:   device.UserAgent,
		IsTrusted:   device.IsTrusted,
		IsBrowser:   device.IsBrowser,
		IsActive:    device.IsActive,
	}
}
