package service

import (
	"context"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

type ShiftServer struct {
	pb.UnimplementedShiftServiceServer
	ShiftStore *ShiftStore
	AuthStore  *AuthStore
}

func NewShiftServer(shiftStore *ShiftStore, authStore *AuthStore) *ShiftServer {
	return &ShiftServer{
		ShiftStore: shiftStore,
		AuthStore:  authStore,
	}
}

// CreateShift creates a new shift
func (s *ShiftServer) CreateShift(ctx context.Context, req *pb.CreateShiftRequest) (*pb.CreateShiftResponse, error) {
	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	if req.Scope != nil {
		return nil, status.Error(403, "Forbidden, You do not have access to create shift ")
	}
	scope := PbScopeToOpScope(req.GetScope())

	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			log.Printf("error in contextChanData.RequestAuth: %v", contextChanData)
			return nil, status.Error(403, "Forbidden, You do not have access to create shift ")
		}
		if scope.ID == "" {
			scope.ID = contextChanData.OrganisationID.String()
			scope.Owner = OwnerType(ORGANISATIONS)
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			log.Printf("error in CheckIfIDInScope: %v", contextChanData.Scopes)
			return nil, status.Error(403, "Forbidden, You do not have access to create shift ")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching shift  %v", err)
	}

	shift := PbShiftToShift(req.GetShift())
	shift.OwnerType = scope.Owner
	shift.OwnerID = uuid.MustParse(scope.ID)
	shift.CreatedBy = contextData.UserID.String()
	shift.UpdatedBy = contextData.UserID.String()
	shift.SerialNumber = k.SERIAL_NUMBER_ABBR()["shifts"] + "-" + shift.ID.String()
	shift.CreatedBy = contextData.UserID.String()
	shift.UpdatedBy = contextData.UserID.String()
	shift.ShiftStatus = k.GENERAL_STATUSES()["ACTIVE"]

	newShift, shiftErr := s.ShiftStore.SaveShift(shift)
	if shiftErr != nil {
		return nil, status.Errorf(500, "error while saving shift %v", shiftErr)
	}

	return &pb.CreateShiftResponse{
		Shift: ShiftToPbShift(newShift),
		Status: &pb.Status{
			Code:    200,
			Message: "Shift created successfully",
		},
	}, nil
}

func (s *ShiftServer) ReadShift(ctx context.Context, req *pb.ReadShiftRequest) (*pb.ReadShiftResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get shift  resource")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to get shift  resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching shift  %v", err)
	}

	shiftID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(400, "invalid shift id: %v", err)
	}

	shift, shiftErr := s.ShiftStore.ReadShift(shiftID, scope)
	if shiftErr != nil {
		return nil, status.Errorf(500, "error while fetching shift %v", shiftErr)
	}

	return &pb.ReadShiftResponse{
		Shift: ShiftToPbShift(shift),
		Status: &pb.Status{
			Code:    200,
			Message: "Shift fetched successfully",
		},
	}, nil
}

func (s *ShiftServer) UpdateShift(ctx context.Context, req *pb.UpdateShiftRequest) (*pb.UpdateShiftResponse, error) {
	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get shift  resource")
		}
		if scope.ID == "" {
			scope.ID = contextChanData.UserID.String()
			scope.Owner = OwnerType(USERS)
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to get shift  resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching shift  %v", err)
	}

	shift := PbShiftToShift(req.GetShift())
	shift.OwnerType = scope.Owner
	shift.OwnerID = uuid.MustParse(scope.ID)
	shift.UpdatedBy = contextData.UserID.String()

	updatedShift, shiftErr := s.ShiftStore.UpdateShift(shift)
	if shiftErr != nil {
		return nil, status.Errorf(500, "error while updating shift %v", shiftErr)
	}

	return &pb.UpdateShiftResponse{
		Shift: ShiftToPbShift(updatedShift),
		Status: &pb.Status{
			Code:    200,
			Message: "Shift updated successfully",
		},
	}, nil
}

func (s *ShiftServer) DeleteShift(ctx context.Context, req *pb.DeleteShiftRequest) (*pb.DeleteShiftResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get shift  resource")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to get shift  resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching shift  %v", err)
	}

	shiftID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(400, "invalid shift id: %v", err)
	}

	shiftErr := s.ShiftStore.DeleteShift(shiftID)
	if shiftErr != nil {
		return nil, status.Errorf(500, "error while deleting shift %v", shiftErr)
	}

	return &pb.DeleteShiftResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Shift deleted successfully",
		},
	}, nil
}

// List Shifts

func (s *ShiftServer) ListShifts(ctx context.Context, req *pb.ListShiftsRequest) (*pb.ListShiftsResponse, error) {
	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to list shifts")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user metadata: %v", err)
	}

	scope := &OpScope{}
	if req.GetScope() != nil {
		scope = PbScopeToOpScope(req.GetScope())
	}
	if scope.ID == "" {
		scope.ID = contextData.UserID.String()
		scope.Owner = OwnerType(USERS)
	}
	scopeIds := GetScopeIDs(contextData.Scopes)

	shifts, err := s.ShiftStore.ReadShifts(scopeIds)
	if err != nil {
		return nil, status.Errorf(500, "error while listing shifts: %v", err)
	}

	var pbShifts []*pb.Shift
	for _, shift := range shifts {
		pbShifts = append(pbShifts, ShiftToPbShift(shift))
	}

	return &pb.ListShiftsResponse{
		Shifts: pbShifts,
		Status: &pb.Status{
			Code:    200,
			Message: "Shifts fetched successfully",
		},
	}, nil
}

func (s *ShiftServer) GetUserShifts(ctx context.Context, req *pb.GetUserShiftsRequest) (*pb.GetUserShiftsResponse, error) {
	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to list shifts")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user metadata: %v", err)
	}

	scopeIds := GetScopeIDs(contextData.Scopes)

	userShifts, err := s.ShiftStore.ReadShifts(scopeIds)
	if err != nil {
		return nil, status.Errorf(500, "error while fetching user shifts: %v", err)
	}

	var pbUserShifts []*pb.ShiftMD
	for _, userShift := range userShifts {
		pbUserShifts = append(pbUserShifts, &pb.ShiftMD{
			Id:        userShift.ID.String(),
			Name:      userShift.Name,
			Color:     userShift.Color,
			StartTime: userShift.StartTime,
			EndTime:   userShift.EndTime,
		})
	}

	return &pb.GetUserShiftsResponse{
		Shifts: pbUserShifts,
		Status: &pb.Status{
			Code:    200,
			Message: "User shifts fetched successfully",
		},
	}, nil
}
