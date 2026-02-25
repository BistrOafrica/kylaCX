package service

import (
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
)

func BranchToPbBranch(branch *Branch) *pb.Branch {
	branchIds := make([]string, 0)
	for _, user := range branch.Users {
		branchIds = append(branchIds, user.ID.String())
	}
	roleIds := make([]string, 0)
	for _, role := range branch.Roles {
		roleIds = append(roleIds, role.ID.String())
	}
	departmentIds := make([]string, 0)
	for _, department := range branch.Departments {
		departmentIds = append(departmentIds, department.ID.String())
	}
	teamIds := make([]string, 0)
	for _, team := range branch.Teams {
		teamIds = append(teamIds, team.ID.String())
	}
	return &pb.Branch{
		Id:            branch.ID.String(),
		Name:          branch.Name,
		SerialNumber:  branch.SerialNumber,
		IsDefault:     branch.IsDefault,
		Description:   branch.Description,
		Status:        branch.Status,
		ParentId:      branch.ParentID.String(),
		UserIds:       branchIds,
		RoleIds:       roleIds,
		Location:      branch.Location,
		Address:       branch.Address,
		DepartmentIds: departmentIds,
		TeamIds:       teamIds,
		OwnerType:     pb.OwnerType(pb.OwnerType_value[string(branch.OwnerType)]),
		OwnerId:       branch.OwnerID.String(),
		CreatedBy:     branch.CreatedBy,
		UpdatedBy:     branch.UpdatedBy,
		CreatedAt:     branch.CreatedAt.String(),
		UpdatedAt:     branch.UpdatedAt.String(),
	}
}

func PbBranchToBranch(branch *pb.Branch) *Branch {
	id, err := uuid.Parse(branch.Id)
	if err != nil {
		id = uuid.New()
	}
	parentId, err := uuid.Parse(branch.ParentId)
	if err != nil {
		parentId = uuid.Nil
	}
	ownerID, err := uuid.Parse(branch.OwnerId)
	if err != nil {
		ownerID = uuid.Nil
	}

	return &Branch{
		ID:           id,
		Name:         branch.Name,
		SerialNumber: utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["branches"], id.String()),
		Description:  branch.Description,
		Status:       branch.Status,
		ParentID:     parentId,
		CreatedBy:    branch.CreatedBy,
		IsDefault:    branch.IsDefault,
		Users:        []User{},
		Roles:        []Role{},
		Departments:  []Department{},
		Teams:        []Team{},
		Address:      branch.Address,
		Location:     branch.Location,
		OwnerType:    OwnerType(branch.OwnerType),
		OwnerID:      ownerID,
	}
}

func BranchesToPbBranches(branches []*Branch) []*pb.Branch {
	var pbBranches []*pb.Branch
	for _, branch := range branches {
		pbBranches = append(pbBranches, BranchToPbBranch(branch))
	}
	return pbBranches
}

func PbBranchesToBranches(branches []*pb.Branch) []*Branch {
	var pbBranches []*Branch
	for _, branch := range branches {
		pbBranches = append(pbBranches, PbBranchToBranch(branch))
	}
	return pbBranches
}
