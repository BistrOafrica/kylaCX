package branch

import (
	"context"
	"kyla-be/internal/authctx"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/service"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

// AuthGateway defines the auth methods BranchServer needs.
type AuthGateway interface {
	ScopeCheck(ctx context.Context, scopeID string) (bool, *authctx.RequestMetadata, error)
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
}

// RoleSaver can persist a Role; satisfied by *service.RbacStore.
type RoleSaver interface {
	SaveRole(role *service.Role) (*service.Role, error)
}

// BranchServer is the gRPC implementation of BranchServiceServer.
type BranchServer struct {
	branchStore *BranchStore
	AuthStore   AuthGateway
	RbacStore   RoleSaver
	pb.UnimplementedBranchServiceServer
}

// NewBranchServer creates a new BranchServer.
func NewBranchServer(branchStore *BranchStore, authStore AuthGateway, rbacStore RoleSaver) *BranchServer {
	return &BranchServer{
		branchStore: branchStore,
		AuthStore:   authStore,
		RbacStore:   rbacStore,
	}
}

// CreateBranch creates a new branch.
func (b *BranchServer) CreateBranch(ctx context.Context, request *pb.CreateBranchRequest) (*pb.CreateBranchResponse, error) {
	scope := authctx.PbScopeToOpScope(request.GetScope())
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

	if b.RbacStore != nil {
		if err := createBranchRoles(branch, b.RbacStore); err != nil {
			return nil, status.Error(500, "Internal Server Error: Failed to create branch")
		}
	}

	saveErr := b.branchStore.Save(branch)
	if saveErr != nil {
		return nil, status.Error(500, "Internal Server Error: Failed to create branch")
	}
	return &pb.CreateBranchResponse{Branch: BranchToPbBranch(branch)}, nil
}

// ReadBranch reads a single branch by ID.
func (b *BranchServer) ReadBranch(ctx context.Context, request *pb.ReadBranchRequest) (*pb.ReadBranchResponse, error) {
	scope := authctx.PbScopeToOpScope(request.GetScope())
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

// ReadBranches returns all branches for the requested scope.
func (b *BranchServer) ReadBranches(ctx context.Context, request *pb.ReadBranchesRequest) (*pb.ReadBranchesResponse, error) {
	scope := authctx.PbScopeToOpScope(request.GetScope())
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

// UpdateBranch updates an existing branch.
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

// DeleteBranch deletes a branch by ID.
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
