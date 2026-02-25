package service

import (
	"context"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

type BreakServer struct {
	pb.UnimplementedBreakServiceServer
	BreakStore *BreakStoreDB
	AuthStore  *AuthStore
}

func NewBreakServer(breakStore *BreakStoreDB, authStore *AuthStore) *BreakServer {
	return &BreakServer{
		BreakStore: breakStore,
		AuthStore:  authStore,
	}
}

// CreateBreak creates a new break
func (s *BreakServer) CreateBreak(ctx context.Context, req *pb.CreateBreakRequest) (*pb.CreateBreakResponse, error) {
	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())

	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			log.Printf("error in contextChanData.RequestAuth: %v", contextChanData)
			return nil, status.Error(403, "Forbidden, You do not have access to create break")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			log.Printf("error in CheckIfIDInScope: %v", contextChanData.Scopes)
			return nil, status.Error(403, "Forbidden, You do not have access to create break")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching break %v", err)
	}

	breakItem := PbScheduleBreakToScheduleBreak(req.GetBreak())
	breakItem.OwnerType = scope.Owner
	breakItem.OwnerID = uuid.MustParse(scope.ID)
	breakItem.CreatedBy = contextData.UserID.String()
	breakItem.UpdatedBy = contextData.UserID.String()
	breakItem.ID = uuid.New()
	breakItem.SerialNumber = k.SERIAL_NUMBER_ABBR()["breaks"] + "-" + breakItem.ID.String()

	newBreak, breakErr := s.BreakStore.SaveBreak(breakItem)
	if breakErr != nil {
		return nil, status.Errorf(500, "error while saving break %v", breakErr)
	}

	return &pb.CreateBreakResponse{
		Break: ScheduleBreakToPbScheduleBreak(newBreak),
		Status: &pb.Status{
			Code:    200,
			Message: "Break created successfully",
		},
	}, nil
}

// ReadBreak retrieves a break by ID
func (s *BreakServer) ReadBreak(ctx context.Context, req *pb.ReadBreakRequest) (*pb.ReadBreakResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get break resource")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to get break resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching break %v", err)
	}

	breakID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(400, "invalid break id: %v", err)
	}

	breakItem, breakErr := s.BreakStore.ReadBreak(breakID, scope)
	if breakErr != nil {
		return nil, status.Errorf(500, "error while fetching break %v", breakErr)
	}

	return &pb.ReadBreakResponse{
		Break: ScheduleBreakToPbScheduleBreak(breakItem),
		Status: &pb.Status{
			Code:    200,
			Message: "Break fetched successfully",
		},
	}, nil
}

// UpdateBreak updates an existing break
func (s *BreakServer) UpdateBreak(ctx context.Context, req *pb.UpdateBreakRequest) (*pb.UpdateBreakResponse, error) {
	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get break resource")
		}
		if scope.ID == "" {
			scope.ID = contextChanData.UserID.String()
			scope.Owner = OwnerType(USERS)
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to get break resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching break %v", err)
	}

	breakItem := PbScheduleBreakToScheduleBreak(req.GetBreak())
	breakItem.OwnerType = scope.Owner
	breakItem.OwnerID = uuid.MustParse(scope.ID)
	breakItem.UpdatedBy = contextData.UserID.String()

	updatedBreak, breakErr := s.BreakStore.UpdateBreak(breakItem)
	if breakErr != nil {
		return nil, status.Errorf(500, "error while updating break %v", breakErr)
	}

	return &pb.UpdateBreakResponse{
		Break: ScheduleBreakToPbScheduleBreak(updatedBreak),
		Status: &pb.Status{
			Code:    200,
			Message: "Break updated successfully",
		},
	}, nil
}

// DeleteBreak deletes a break by ID
func (s *BreakServer) DeleteBreak(ctx context.Context, req *pb.DeleteBreakRequest) (*pb.DeleteBreakResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get break resource")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to get break resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching break %v", err)
	}

	breakID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(400, "invalid break id: %v", err)
	}

	breakErr := s.BreakStore.DeleteBreak(breakID)
	if breakErr != nil {
		return nil, status.Errorf(500, "error while deleting break %v", breakErr)
	}

	return &pb.DeleteBreakResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Break deleted successfully",
		},
	}, nil
}

// ListBreaks retrieves all breaks for a schedule
func (s *BreakServer) ListBreaks(ctx context.Context, req *pb.ListBreaksRequest) (*pb.ListBreaksResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get breaks")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to get breaks")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching breaks %v", err)
	}

	scheduleID, err := uuid.Parse(req.GetScheduleId())
	if err != nil {
		return nil, status.Errorf(400, "invalid schedule id: %v", err)
	}

	breaks, breakErr := s.BreakStore.ListBreaks(scheduleID, scope)
	if breakErr != nil {
		return nil, status.Errorf(500, "error while fetching breaks %v", breakErr)
	}

	pbBreaks := make([]*pb.ScheduleBreak, len(breaks))
	for i, breakItem := range breaks {
		pbBreaks[i] = ScheduleBreakToPbScheduleBreak(breakItem)
	}

	return &pb.ListBreaksResponse{
		Breaks: pbBreaks,
		Status: &pb.Status{
			Code:    200,
			Message: "Breaks fetched successfully",
		},
	}, nil
}
