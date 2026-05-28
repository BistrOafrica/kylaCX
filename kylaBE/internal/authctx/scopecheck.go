package authctx

import (
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
)

// CheckOpScope returns true if the request metadata grants access to opScope.
func CheckOpScope(metadata *RequestMetadata, opScope *OpScope) bool {
	userScopes := &Scopes{
		User:         metadata.UserID,
		Teams:        metadata.Scopes.Teams,
		Departments:  metadata.Scopes.Departments,
		Branch:       metadata.Scopes.Branch,
		Organisation: metadata.Scopes.Organisation,
	}

	switch opScope.Owner {
	case TEAMS:
		if utils.Includes(userScopes.Teams, opScope.ID) {
			return true
		}
	case USERS:
		if userScopes.User.String() == opScope.ID {
			return true
		}
	case BRANCHES:
		if userScopes.Branch.String() == opScope.ID {
			return true
		}
	case DEPARTMENTS:
		if utils.Includes(userScopes.Departments, opScope.ID) {
			return true
		}
	case ORGANISATIONS:
		if userScopes.Organisation.String() == opScope.ID {
			return true
		}
	default:
		return false
	}
	return false
}

// GetScopeIDs returns all IDs in the given scope as a flat string slice.
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

// CheckIfIDInScope returns true if id is contained within any scope bucket.
func CheckIfIDInScope(scope *Scopes, id string) bool {
	for _, i := range GetScopeIDs(scope) {
		if i == id {
			return true
		}
	}
	return false
}
