package branch

import (
	"fmt"
	"kyla-be/internal/authctx"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/service"
	"kyla-be/pkg/utils"
	"time"

	"github.com/google/uuid"
)

func BranchToPbBranch(branch *Branch) *pb.Branch {
	return &pb.Branch{
		Id:           branch.ID.String(),
		Name:         branch.Name,
		SerialNumber: branch.SerialNumber,
		IsDefault:    branch.IsDefault,
		Description:  branch.Description,
		Status:       branch.Status,
		ParentId:     branch.ParentID.String(),
		Location:     branch.Location,
		Address:      branch.Address,
		OwnerType:    pb.OwnerType(pb.OwnerType_value[string(branch.OwnerType)]),
		OwnerId:      branch.OwnerID.String(),
		CreatedBy:    branch.CreatedBy,
		UpdatedBy:    branch.UpdatedBy,
		CreatedAt:    branch.CreatedAt.String(),
		UpdatedAt:    branch.UpdatedAt.String(),
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
		Address:      branch.Address,
		Location:     branch.Location,
		OwnerType:    authctx.OwnerType(branch.OwnerType),
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

// createBranchRoles creates the default admin/supervisor/agent roles for a branch.
func createBranchRoles(b *Branch, roleStore RoleSaver) error {
	adminId := uuid.New()
	supervisorId := uuid.New()
	agentId := uuid.New()

	roles := []service.Role{
		{
			ID:                  adminId,
			Name:                fmt.Sprintf("%s Admin", b.Name),
			Description:         "Branch Admin Role",
			PermissionCodeNames: k.ADMIN_PERMISSIONS(),
			CreatedBy:           "USERS",
			UpdatedAt:           time.Now(),
			CreatedAt:           time.Now(),
			IsDefault:           false,
			OwnerType:           service.OwnerType(authctx.BRANCHES),
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
			OwnerType:           service.OwnerType(authctx.BRANCHES),
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
			OwnerType:           service.OwnerType(authctx.BRANCHES),
			OwnerID:             b.ID,
		},
	}

	for i := range roles {
		roles[i].SerialNumber = utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["roles"], roles[i].ID.String())
		if _, err := roleStore.SaveRole(&roles[i]); err != nil {
			return err
		}
	}
	return nil
}
