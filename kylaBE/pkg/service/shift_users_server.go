package service

import (
	"context"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

type UserAssignmentServer struct {
	pb.UnimplementedShiftServiceServer
	ShiftStore *ShiftStore
	AuthStore  *AuthStore
}

func NewUserAssignmentServer(shiftStore *ShiftStore) *UserAssignmentServer {
	return &UserAssignmentServer{
		ShiftStore: shiftStore,
	}
}

// AssignUsersToShift assigns users to a shift
func (s *UserAssignmentServer) AssignUsersToShift(ctx context.Context, req *pb.AssignUsersToShiftRequest) (*pb.AssignUsersToShiftResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to assign users to shift")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to assign users to shift")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user assignment %v", err)
	}

	shiftID, err := uuid.Parse(req.GetShiftId())
	if err != nil {
		return nil, status.Errorf(400, "invalid shift id: %v", err)
	}

	userIDs := make([]uuid.UUID, len(req.GetUserIds()))
	for i, userIDStr := range req.GetUserIds() {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, status.Errorf(400, "invalid user id: %v", err)
		}
		userIDs[i] = userID
	}

	err = s.ShiftStore.AssignUsersToShift(shiftID, userIDs)
	if err != nil {
		return nil, status.Errorf(500, "error while assigning users to shift %v", err)
	}

	return &pb.AssignUsersToShiftResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Users assigned to shift successfully",
		},
	}, nil
}

// RemoveUsersFromShift removes users from a shift
func (s *UserAssignmentServer) RemoveUsersFromShift(ctx context.Context, req *pb.RemoveUsersFromShiftRequest) (*pb.RemoveUsersFromShiftResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to remove users from shift")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to remove users from shift")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user assignment %v", err)
	}

	shiftID, err := uuid.Parse(req.GetShiftId())
	if err != nil {
		return nil, status.Errorf(400, "invalid shift id: %v", err)
	}

	userIDs := make([]uuid.UUID, len(req.GetUserIds()))
	for i, userIDStr := range req.GetUserIds() {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, status.Errorf(400, "invalid user id: %v", err)
		}
		userIDs[i] = userID
	}

	err = s.ShiftStore.RemoveUsersFromShift(shiftID, userIDs)
	if err != nil {
		return nil, status.Errorf(500, "error while removing users from shift %v", err)
	}

	return &pb.RemoveUsersFromShiftResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Users removed from shift successfully",
		},
	}, nil
}

// ListShiftUsers retrieves all users assigned to a shift
func (s *UserAssignmentServer) ListShiftUsers(ctx context.Context, req *pb.ListShiftUsersRequest) (*pb.ListShiftUsersResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to list shift users")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to list shift users")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching shift users %v", err)
	}

	shiftID, err := uuid.Parse(req.GetShiftId())
	if err != nil {
		return nil, status.Errorf(400, "invalid shift id: %v", err)
	}

	users, err := s.ShiftStore.ListShiftUsers(shiftID)
	if err != nil {
		return nil, status.Errorf(500, "error while fetching shift users %v", err)
	}

	pbUsers := make([]*pb.User, len(users))
	for i, user := range users {
		pbUsers[i] = UserToPbUser(&user)
	}

	return &pb.ListShiftUsersResponse{
		Users: pbUsers,
		Status: &pb.Status{
			Code:    200,
			Message: "Shift users fetched successfully",
		},
	}, nil
}
