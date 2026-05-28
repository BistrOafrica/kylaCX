package user

import (
	"kyla-be/internal/agentops"
	"kyla-be/internal/authctx"
	"kyla-be/internal/rbac"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
)

// PbUserToUser converts a protobuf User to a User model.
func PbUserToUser(pbUser *pb.User) *User {
	id, err := uuid.Parse(pbUser.Id)
	if err != nil {
		id = uuid.New()
	}
	ownerID, err := uuid.Parse(pbUser.OwnerId)
	if err != nil {
		ownerID = uuid.Nil
	}
	roles := make([]rbac.Role, 0)
	for _, r := range rbac.PbRolesToRoles(pbUser.Roles) {
		roles = append(roles, *r)
	}
	return &User{
		ID:             id,
		SerialNumber:   utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["users"], id.String()),
		FirstName:      pbUser.FirstName,
		LastName:       pbUser.LastName,
		Username:       pbUser.Username,
		Email:          pbUser.Email,
		Phone:          pbUser.Phone,
		Status:         pbUser.Status,
		IsDefault:      false,
		CreatedBy:      pbUser.CreatedBy,
		Roles:          roles,
		EmailSignature: pbUser.EmailSignature,
		OwnerType:      authctx.OwnerType(pbUser.OwnerType),
		OwnerID:        ownerID,
	}
}

// UserToPbUser converts a User model to its protobuf representation.
// Branch list is returned empty; the list requires a dedicated query
// using the user_branches join table (post-refactor concern).
func UserToPbUser(user *User) *pb.User {
	roles := make([]*pb.Role, 0)
	for i := range user.Roles {
		roles = append(roles, rbac.RoleToPbRole(&user.Roles[i]))
	}

	return &pb.User{
		Id:                    user.ID.String(),
		FirstName:             user.FirstName,
		LastName:              user.LastName,
		Username:              user.Username,
		Email:                 user.Email,
		Phone:                 user.Phone,
		Roles:                 roles,
		Status:                user.Status,
		CreatedBy:             user.CreatedBy,
		AgentStatusId:         user.AgentStatusID.String(),
		AgentStatus:           AgentStatusToPbAgentStatus(&user.AgentStatus),
		EmailSignature:        user.EmailSignature,
		CurrentBranchId:       user.CurrentBranchID.String(),
		CurrentOrganisationId: user.CurrentOrganisationID.String(),
		Branches:              []*pb.Branch{},
		IsDefault:             user.IsDefault,
		OwnerType:             pb.OwnerType(pb.OwnerType_value[string(user.OwnerType)]),
		OwnerId:               user.OwnerID.String(),
		CreatedAt:             user.CreatedAt.String(),
		UpdatedAt:             user.UpdatedAt.String(),
		UpdatedBy:             user.UpdatedBy,
		LastLogin:             user.LastLogin.String(),
		LastMfaLogin:          user.LastMfaLogin.String(),
		Image:                 user.Image,
		CallingCode:           user.CallingCode,
	}
}

// UsersToPbUsers converts a slice of User pointers to protobuf User pointers.
func UsersToPbUsers(users []*User) []*pb.User {
	var us []*pb.User
	for _, user := range users {
		us = append(us, UserToPbUser(user))
	}
	return us
}

// PbUsersToUsers converts a slice of protobuf User pointers to User model values.
func PbUsersToUsers(users []*pb.User) []User {
	var us []User
	for _, user := range users {
		us = append(us, *PbUserToUser(user))
	}
	return us
}

// AgentStatusToPbAgentStatus converts an agentops.AgentStatus to its protobuf form.
func AgentStatusToPbAgentStatus(s *agentops.AgentStatus) *pb.AgentStatus {
	if s == nil {
		return &pb.AgentStatus{}
	}
	statusChanges := make([]*pb.StatusChange, 0, len(s.StatusChanges))
	for _, sc := range s.StatusChanges {
		statusChanges = append(statusChanges, &pb.StatusChange{
			Id:          sc.ID.String(),
			StatusType:  pb.StatusType(sc.StatusType),
			Description: sc.Description,
			StartTime:   sc.StartTime.String(),
			EndTime:     sc.EndTime.String(),
		})
	}
	return &pb.AgentStatus{
		Id:            s.ID.String(),
		AgentId:       s.AgentID.String(),
		StatusChanges: statusChanges,
	}
}
