package service

import (
	"context"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

type ShiftScheduleServer struct {
	pb.UnimplementedShiftScheduleServiceServer
	ShiftScheduleStore *ShiftScheduleStore
	AuthStore          *AuthStore
}

func NewScheduleServer(scheduleStore *ShiftScheduleStore, authStore *AuthStore) *ShiftScheduleServer {
	return &ShiftScheduleServer{
		ShiftScheduleStore: scheduleStore,
		AuthStore:          authStore,
	}
}

// CreateSchedule creates a new schedule
func (s *ShiftScheduleServer) CreateSchedule(ctx context.Context, req *pb.CreateScheduleRequest) (*pb.CreateScheduleResponse, error) {
	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())

	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			log.Printf("error in contextChanData.RequestAuth: %v", contextChanData)
			return nil, status.Error(403, "Forbidden, You do not have access to create schedule")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			log.Printf("error in CheckIfIDInScope: %v", contextChanData.Scopes)
			return nil, status.Error(403, "Forbidden, You do not have access to create schedule")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching schedule %v", err)
	}

	schedule := PbShiftScheduleToShiftSchedule(req.GetSchedule())
	schedule.OwnerType = scope.Owner
	schedule.OwnerID = uuid.MustParse(scope.ID)
	schedule.CreatedBy = contextData.UserID.String()
	schedule.UpdatedBy = contextData.UserID.String()
	schedule.ID = uuid.New()
	schedule.SerialNumber = k.SERIAL_NUMBER_ABBR()["schedules"] + "-" + schedule.ID.String()
	schedule.ShiftStatus = k.GENERAL_STATUSES()["ACTIVE"]

	newSchedule, scheduleErr := s.ShiftScheduleStore.SaveSchedule(schedule)
	if scheduleErr != nil {
		return nil, status.Errorf(500, "error while saving schedule %v", scheduleErr)
	}

	return &pb.CreateScheduleResponse{
		Schedule: ShiftScheduleToPbShiftSchedule(newSchedule),
		Status: &pb.Status{
			Code:    200,
			Message: "Schedule created successfully",
		},
	}, nil
}

// ReadSchedule retrieves a schedule by ID
func (s *ShiftScheduleServer) ReadSchedule(ctx context.Context, req *pb.ReadScheduleRequest) (*pb.ReadScheduleResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get schedule resource")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to get schedule resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching schedule %v", err)
	}

	scheduleID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(400, "invalid schedule id: %v", err)
	}

	schedule, scheduleErr := s.ShiftScheduleStore.ReadSchedule(scheduleID, scope)
	if scheduleErr != nil {
		return nil, status.Errorf(500, "error while fetching schedule %v", scheduleErr)
	}

	return &pb.ReadScheduleResponse{
		Schedule: ShiftScheduleToPbShiftSchedule(schedule),
		Status: &pb.Status{
			Code:    200,
			Message: "Schedule fetched successfully",
		},
	}, nil
}

// UpdateSchedule updates an existing schedule
func (s *ShiftScheduleServer) UpdateSchedule(ctx context.Context, req *pb.UpdateScheduleRequest) (*pb.UpdateScheduleResponse, error) {
	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get schedule resource")
		}
		if scope.ID == "" {
			scope.ID = contextChanData.UserID.String()
			scope.Owner = OwnerType(USERS)
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to get schedule resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching schedule %v", err)
	}

	schedule := PbShiftScheduleToShiftSchedule(req.GetSchedule())
	schedule.OwnerType = scope.Owner
	schedule.OwnerID = uuid.MustParse(scope.ID)
	schedule.UpdatedBy = contextData.UserID.String()

	updatedSchedule, scheduleErr := s.ShiftScheduleStore.UpdateSchedule(schedule)
	if scheduleErr != nil {
		return nil, status.Errorf(500, "error while updating schedule %v", scheduleErr)
	}

	return &pb.UpdateScheduleResponse{
		Schedule: ShiftScheduleToPbShiftSchedule(updatedSchedule),
		Status: &pb.Status{
			Code:    200,
			Message: "Schedule updated successfully",
		},
	}, nil
}

// DeleteSchedule deletes a schedule by ID
func (s *ShiftScheduleServer) DeleteSchedule(ctx context.Context, req *pb.DeleteScheduleRequest) (*pb.DeleteScheduleResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get schedule resource")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to get schedule resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching schedule %v", err)
	}

	scheduleID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(400, "invalid schedule id: %v", err)
	}

	scheduleErr := s.ShiftScheduleStore.DeleteSchedule(scheduleID)
	if scheduleErr != nil {
		return nil, status.Errorf(500, "error while deleting schedule %v", scheduleErr)
	}

	return &pb.DeleteScheduleResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Schedule deleted successfully",
		},
	}, nil
}

// ListSchedules retrieves all schedules for a shift
func (s *ShiftScheduleServer) ListSchedules(ctx context.Context, req *pb.ListSchedulesRequest) (*pb.ListSchedulesResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get schedules")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to get schedules")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching schedules %v", err)
	}

	shiftID, err := uuid.Parse(req.GetShiftId())
	if err != nil {
		return nil, status.Errorf(400, "invalid shift id: %v", err)
	}

	schedules, scheduleErr := s.ShiftScheduleStore.ListSchedules(shiftID, scope)
	if scheduleErr != nil {
		return nil, status.Errorf(500, "error while fetching schedules %v", scheduleErr)
	}

	pbSchedules := make([]*pb.ShiftSchedule, len(schedules))
	for i, schedule := range schedules {
		pbSchedules[i] = ShiftScheduleToPbShiftSchedule(schedule)
	}

	return &pb.ListSchedulesResponse{
		Schedules: pbSchedules,
		Status: &pb.Status{
			Code:    200,
			Message: "Schedules fetched successfully",
		},
	}, nil
}

// Clock in and out

func (s *ShiftScheduleServer) ClockIn(ctx context.Context, req *pb.ClockInRequest) (*pb.ClockInResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	contextData := &RequestMetadata{}
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get schedules")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to get schedules")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching schedules %v", err)
	}

	// Clock in the user
	clockInRecord := &UserShiftRecord{
		ID:        uuid.New(),
		UserID:    uuid.MustParse(contextData.UserID.String()),
		ShiftID:   uuid.MustParse(req.GetShiftId()),
		StartTime: time.Now().String(),
		CreatedBy: contextData.UserID.String(),
		UpdatedBy: contextData.UserID.String(),
	}
	if err := s.ShiftScheduleStore.ClockOut(clockInRecord); err != nil {
		return nil, status.Errorf(500, "error while clocking in: %v", err)
	}

	return &pb.ClockInResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Clocked in successfully",
		},
	}, nil
}

func (s *ShiftScheduleServer) ClockOut(ctx context.Context, req *pb.ClockOutRequest) (*pb.ClockOutResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	contextData := &RequestMetadata{}
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get schedules")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to get schedules")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching schedules %v", err)
	}

	// Clock out the user
	clockOutRecord := &UserShiftRecord{
		ID:        uuid.New(),
		UserID:    uuid.MustParse(contextData.UserID.String()),
		ShiftID:   uuid.MustParse(req.GetShiftId()),
		EndTime:   time.Now().String(),
		CreatedBy: contextData.UserID.String(),
		UpdatedBy: contextData.UserID.String(),
	}
	if err := s.ShiftScheduleStore.ClockOut(clockOutRecord); err != nil {
		return nil, status.Errorf(500, "error while clocking out: %v", err)
	}

	return &pb.ClockOutResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Clocked out successfully",
		},
	}, nil
}

// take break.

func (s *ShiftScheduleServer) TakeBreak(ctx context.Context, req *pb.TakeBreakRequest) (*pb.TakeBreakResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	contextData := &RequestMetadata{}
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get schedules")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to get schedules")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching schedules %v", err)
	}

	pbBR := &pb.BreakRecord{
		Id:                uuid.New().String(),
		UserId:            req.UserId,
		BreakId:           req.BreakId,
		UserShiftRecordId: req.UserShiftRecordId,
		StartTime:         time.Now().String(),
		EndTime:           time.Now().String(),
		CreatedBy:         contextData.UserID.String(),
		UpdatedBy:         contextData.UserID.String(),
	}
	// Take a break
	takeBreakRecord := PbBreakRecordToBreakRecord(pbBR)
	if err := s.ShiftScheduleStore.TakeBreak(takeBreakRecord); err != nil {
		return nil, status.Errorf(500, "error while taking break: %v", err)
	}

	return &pb.TakeBreakResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Break taken successfully",
		},
	}, nil
}

func (s *ShiftScheduleServer) ResumeBreak(ctx context.Context, req *pb.ResumeBreakRequest) (*pb.ResumeBreakResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(req.GetScope())
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get schedules")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(403, "Forbidden, You do not have access to get schedules")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching schedules %v", err)
	}
	// Take a break
	takeBreakRecord := PbBreakRecordToBreakRecord(req.BreakRecord)
	takeBreakRecord.EndTime = time.Now().String()
	if err := s.ShiftScheduleStore.ResumeBreak(takeBreakRecord); err != nil {
		return nil, status.Errorf(500, "error while taking break: %v", err)
	}

	return &pb.ResumeBreakResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Break resumed successfully",
		},
	}, nil
}
