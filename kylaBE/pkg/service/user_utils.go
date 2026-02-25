package service

import (
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
)

func PbUserToUser(user *pb.User) *User {
	id, err := uuid.Parse(user.Id)
	if err != nil {
		id = uuid.New()
	}
	roles := make([]Role, 0)
	for _, role := range PbRolesToRoles(user.Roles) {
		roles = append(roles, *role)
	}
	return &User{
		ID:             id,
		SerialNumber:   utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["users"], id.String()),
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Username:       user.Username,
		Email:          user.Email,
		Phone:          user.Phone,
		Status:         user.Status,
		IsDefault:      false,
		CreatedBy:      user.CreatedBy,
		Roles:          roles,
		EmailSignature: user.EmailSignature,
		OwnerType:      OwnerType(user.OwnerType),
		OwnerID:        uuid.MustParse(user.OwnerId),
	}
}

func UserToPbUser(user *User) *pb.User {
	roles := make([]*pb.Role, 0)
	for _, role := range user.Roles {
		roles = append(roles, RoleToPbRole(&role))
	}
	branches := make([]*pb.Branch, 0)
	for _, branch := range user.Branches {
		branches = append(branches, BranchToPbBranch(&branch))
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
		AgentStatus:           StatusToPbStatus(&user.AgentStatus),
		EmailSignature:        user.EmailSignature,
		CurrentBranchId:       user.CurrentBranchID.String(),
		CurrentOrganisationId: user.CurrentOrganisationID.String(),
		Branches:              branches,
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

func PbUsersToUsers(users []*pb.User) []User {
	var us []User
	for _, user := range users {
		us = append(us, *PbUserToUser(user))
	}
	return us
}

func UsersToPbUsers(users []*User) []*pb.User {
	var us []*pb.User
	for _, user := range users {
		us = append(us, UserToPbUser(user))
	}
	return us
}
