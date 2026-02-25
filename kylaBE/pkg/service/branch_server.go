package service

import (
	"context"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

// BranchServer is the implementation of the BranchServiceServer
type BranchServer struct {
	branchStore *BranchStore
	AuthStore   *AuthStore
	pb.UnimplementedBranchServiceServer
}

// NewBranchServer creates a new instance of BranchServer
func NewBranchServer(branchStore *BranchStore, authStore *AuthStore) *BranchServer {
	return &BranchServer{branchStore: branchStore}
}

// CreateBranch creates a new branch
func (b *BranchServer) CreateBranch(ctx context.Context, request *pb.CreateBranchRequest) (*pb.CreateBranchResponse, error) {
	scope := PbScopeToOpScope(request.GetScope())
	auth, contextData, err := b.AuthStore.ScopeCheck(ctx, scope.ID)
	if err != nil {
		return nil, err
	}
	if auth == k.NewConsts().FALSE_BOOL {
		return nil, status.Error(403, "Forbidden, You do not have access to create app")
	}

	branch := PbBranchToBranch(request.Branch)
	branch.ID = uuid.New()
	branch.Status = k.GENERAL_STATUSES()["ACTIVE"]
	branch.OwnerType = scope.Owner
	branch.OwnerID = uuid.MustParse(scope.ID)
	branch.CreatedBy = contextData.UserID.String()

	if err := branch.CREATE_BRANCH_ROLES(b.AuthStore.RbacStore); err != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to create branch")
	}
	saveErr := b.branchStore.Save(branch)

	if saveErr != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to create branch")
	}
	return &pb.CreateBranchResponse{Branch: BranchToPbBranch(branch)}, nil
}

func (b *BranchServer) ReadBranch(ctx context.Context, request *pb.ReadBranchRequest) (*pb.ReadBranchResponse, error) {
	scope := PbScopeToOpScope(request.GetScope())
	auth, _, err := b.AuthStore.ScopeCheck(ctx, scope.ID)
	if err != nil {
		return nil, err
	}
	if auth == k.NewConsts().FALSE_BOOL {
		return nil, status.Error(403, "Forbidden, You do not have access to create app")
	}

	branch, readErr := b.branchStore.FindByID(request.GetId())
	if readErr != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to read branch")
	}
	return &pb.ReadBranchResponse{Branch: BranchToPbBranch(branch)}, nil
}

func (b *BranchServer) ReadBranches(ctx context.Context, request *pb.ReadBranchesRequest) (*pb.ReadBranchesResponse, error) {
	scope := PbScopeToOpScope(request.GetScope())
	auth, _, err := b.AuthStore.ScopeCheck(ctx, scope.ID)
	if err != nil {
		return nil, err
	}
	if auth == k.NewConsts().FALSE_BOOL {
		return nil, status.Error(403, "Forbidden, You do not have access to read resource")
	}

	branches, readErr := b.branchStore.FindByOwner(scope)
	if readErr != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to read all branches")
	}
	return &pb.ReadBranchesResponse{Branches: BranchesToPbBranches(branches)}, nil
}

func (b *BranchServer) UpdateBranch(ctx context.Context, request *pb.UpdateBranchRequest) (*pb.UpdateBranchResponse, error) {
	reqAuth, authErr := b.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil {
		return nil, status.Error(403, "Forbidden: You don't have permission to update a branch")
	}
	if reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to update a branch")
	}

	branch := PbBranchToBranch(request.Branch)
	saveErr := b.branchStore.Update(branch)
	if saveErr != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to update branch")
	}
	return &pb.UpdateBranchResponse{Branch: BranchToPbBranch(branch)}, nil
}

func (b *BranchServer) DeleteBranch(ctx context.Context, request *pb.DeleteBranchRequest) (*pb.DeleteBranchResponse, error) {
	reqAuth, authErr := b.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to delete a branch")
	}

	if deleteErr := b.branchStore.Delete(request.GetId()); deleteErr != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to delete branch")
	}
	return &pb.DeleteBranchResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Branch deleted successfully",
		},
	}, nil
}
