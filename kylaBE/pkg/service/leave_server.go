package service

import (
	"context"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LeaveServer struct {
	pb.UnimplementedLeaveServiceServer
	LeaveStore *LeaveStoreDB
	AuthStore  *AuthStore
}

func NewLeaveServer(leaveStore *LeaveStoreDB, authStore *AuthStore) *LeaveServer {
	return &LeaveServer{
		LeaveStore: leaveStore,
		AuthStore:  authStore,
	}
}

func (s *LeaveServer) CreateLeaveType(ctx context.Context, req *pb.CreateLeaveTypeRequest) (*pb.CreateLeaveTypeResponse, error) {
	err := ValidateLeaveTypeFields(req.Data)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(&pb.Scope{
		OwnerType: req.GetData().GetOwnerType(),
		OwnerId:   req.GetData().GetOwnerId(),
	})
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		if !CheckIfIDInScope(contextChanData.Scopes, scope.ID) {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	leaveType := PbLeaveTypeToLeaveType(req.GetData())
	leaveType.OwnerType = scope.Owner
	leaveType.OwnerID = uuid.MustParse(scope.ID)
	leaveType.CreatedBy = contextData.UserID.String()

	newLeaveType, leaveTypeErr := s.LeaveStore.CreateLeaveType(leaveType)
	if leaveTypeErr != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Unknown, "error while saving leave type %v", leaveTypeErr)
	}

	return &pb.CreateLeaveTypeResponse{
		Data: LeaveTypeToPbLeaveType(newLeaveType),
	}, nil
}

func (s *LeaveServer) UpdateLeaveType(ctx context.Context, request *pb.UpdateLeaveTypeRequest) (*pb.UpdateLeaveTypeResponse, error) {
	err := ValidateLeaveTypeFields(request.Data)
	if err != nil {
		return nil, err
	}

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	var contextData *RequestMetadata
	select {
	case contextData = <-contextDataChan:
		if contextData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	if contextData == nil || contextData.Scopes == nil {
		return nil, status.Error(codes.Unauthenticated, "authentication context missing")
	}

	incomingUpdatedLeaveType := PbLeaveTypeToLeaveType(request.GetData())

	prevLeaveType, err := s.LeaveStore.ReadLeaveTypeById(contextData, incomingUpdatedLeaveType.ID.String())
	if err != nil {
		return nil, status.Error(codes.NotFound, "leave type not found")
	}

	if !CheckIfIDInScope(contextData.Scopes, prevLeaveType.OwnerID.String()) {
		return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
	}

	if incomingUpdatedLeaveType.EarnedType == UPFRONT {
		err = s.LeaveStore.DeleteEarnedLeaveConditions(incomingUpdatedLeaveType.ID.String())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "error while deleting earned leave conditions: %v", err)
		}
	}

	if incomingUpdatedLeaveType.EarnedType == EARNED {
		incomingUpdatedLeaveType.TimeLimit = 0
		incomingUpdatedLeaveType.TimeLimitCycle = ""
	}

	updatedLeaveType, err := s.LeaveStore.UpdateLeaveType(incomingUpdatedLeaveType)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "error while updating leave type: %v", err)
	}

	return &pb.UpdateLeaveTypeResponse{Data: LeaveTypeToPbLeaveType(updatedLeaveType)}, nil
}

func (s *LeaveServer) ReadLeaveType(ctx context.Context, request *pb.ReadLeaveTypeRequest) (*pb.ReadLeaveTypeResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	var contextChanData *RequestMetadata

	select {
	case contextChanData = <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	leaveType, err := s.LeaveStore.ReadLeaveTypeById(contextChanData, request.GetId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "leave type not found")
	}

	return &pb.ReadLeaveTypeResponse{
		Data: LeaveTypeToPbLeaveType(leaveType),
	}, nil
}

func (s *LeaveServer) ReadLeaveTypes(ctx context.Context, req *pb.ReadLeaveTypesRequest) (*pb.ReadLeaveTypesResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to access this resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)

	}
	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "Forbidden: You don't have permission to read leave types")
	}

	idsAllowingAccess := GetScopeIDs(contextData.Scopes)

	page := req.GetPage()
	if page == 0 {
		page = 1
	}

	perPage := req.GetPerPage()
	if perPage == 0 {
		perPage = 10
	}

	leaveTypes, leaveTypesCount, leaveTypeErr := s.LeaveStore.ReadLeaveTypes(idsAllowingAccess, page, perPage)
	if leaveTypeErr != nil {
		return nil, status.Errorf(codes.Unknown, "error while fetching leave types %v", leaveTypeErr)
	}

	var pbLeaveTypes []*pb.LeaveType
	for _, leaveType := range leaveTypes {
		pbLeaveTypes = append(pbLeaveTypes, LeaveTypeToPbLeaveType(leaveType))
	}
	return &pb.ReadLeaveTypesResponse{
		TotalCount: leaveTypesCount,
		Data:       pbLeaveTypes,
	}, nil
}

func (s *LeaveServer) DeleteLeaveType(ctx context.Context, request *pb.DeleteLeaveTypeRequest) (*pb.DeleteLeaveTypeResponse, error) {
	if request.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "leave type ID is required")
	}

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	var contextData *RequestMetadata

	select {
	case contextData = <-contextDataChan:
		if contextData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	if contextData == nil || contextData.Scopes == nil {
		return nil, status.Error(codes.Unauthenticated, "authentication context missing")
	}

	prevLeaveType, err := s.LeaveStore.ReadLeaveTypeById(contextData, request.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "leave type not found")
	}

	if !CheckIfIDInScope(contextData.Scopes, prevLeaveType.OwnerID.String()) {
		return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
	}

	if err := s.LeaveStore.DeleteLeaveType(request.Id); err != nil {
		return nil, status.Error(codes.NotFound, "leave type not found")
	}

	return &pb.DeleteLeaveTypeResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Leave type deleted successfully",
		},
	}, nil
}

func (s *LeaveServer) CreateLeaveRequest(ctx context.Context, req *pb.CreateLeaveRequestRequest) (*pb.CreateLeaveRequestResponse, error) {
	err := ValidateLeaveRequestFields(req.Data)
	if err != nil {
		return nil, err
	}

	duration, err := CalculateDuration(req.Data.GetStartDate(), req.Data.GetEndDate(), req.Data.GetStartTime(), req.Data.GetEndTime(), req.Data.GetDurationUnit())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	if req.Data.UserId != contextData.UserID.String() {
		return nil, status.Error(codes.PermissionDenied, "Forbidden, You can't create a leave request for another user")
	}

	leaveReq := PbLeaveRequesttoLeaveRequest(req.GetData())
	leaveReq.Status = LEAVE_PENDING
	leaveReq.Duration = duration
	leaveReq.CreatedBy = contextData.UserID.String()

	newLeaveRequest, leaveRequestErr := s.LeaveStore.CreateLeaveRequest(leaveReq)
	if leaveRequestErr != nil {
		return nil, status.Errorf(codes.Unknown, "error while saving leave request %v", leaveRequestErr)
	}

	return &pb.CreateLeaveRequestResponse{
		Data: LeaveRequestToPbLeaveRequest(newLeaveRequest),
	}, nil
}

func (s *LeaveServer) ReadLeaveRequest(ctx context.Context, request *pb.ReadLeaveRequestRequest) (*pb.ReadLeaveRequestResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	var contextChanData *RequestMetadata

	select {
	case contextChanData = <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	leaveRequest, err := s.LeaveStore.ReadLeaveRequestById(contextChanData, request.GetId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "leave request not found")
	}

	return &pb.ReadLeaveRequestResponse{
		Data: LeaveRequestToPbLeaveRequest(leaveRequest),
	}, nil
}

func (s *LeaveServer) ReadLeaveRequests(ctx context.Context, req *pb.ReadLeaveRequestsRequest) (*pb.ReadLeaveRequestsResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to access this resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)

	}
	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "Forbidden: You don't have permission to read leave requests")
	}

	contextDataForLeaveTypes, err := s.AuthStore.AuthInternalRequests(contextData.Authorization, "READ_LEAVE_TYPES")
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}

	idsAllowingAccess := GetScopeIDs(contextDataForLeaveTypes.Scopes)

	page := req.GetPage()
	if page == 0 {
		page = 1
	}

	perPage := req.GetPerPage()
	if perPage == 0 {
		perPage = 10
	}

	var leaveTypeIDs []uuid.UUID
	leaveTypes, leaveTypesCount, leaveTypeErr := s.LeaveStore.ReadLeaveTypes(idsAllowingAccess, 1, 100)
	if leaveTypeErr != nil {
		return nil, status.Errorf(codes.Unknown, "error while fetching leave types %v", leaveTypeErr)
	}

	for _, lt := range leaveTypes {
		leaveTypeIDs = append(leaveTypeIDs, lt.ID)
	}

	if leaveTypesCount > 100 {
		noOfFetches := (leaveTypesCount + 99) / 100 // Ensure rounding up

		for i := int32(1); i < noOfFetches; i++ {
			moreLeaveTypes, _, leaveTypeErr := s.LeaveStore.ReadLeaveTypes(idsAllowingAccess, i*100+1, 100)
			if leaveTypeErr != nil {
				return nil, status.Errorf(codes.Unknown, "error while fetching leave types %v", leaveTypeErr)
			}
			for _, lt := range moreLeaveTypes {
				leaveTypeIDs = append(leaveTypeIDs, lt.ID)
			}
		}
	}

	leaveRequests, leaveRequestsCount, leaveRequestsErr := s.LeaveStore.ReadLeaveRequests(leaveTypeIDs, page, perPage)
	if leaveRequestsErr != nil {
		return nil, status.Errorf(codes.Unknown, "error while fetching leave requests %v", leaveRequestsErr)
	}

	var pbLeaveRequests []*pb.LeaveRequest
	for _, leaveReq := range leaveRequests {
		pbLeaveRequests = append(pbLeaveRequests, LeaveRequestToPbLeaveRequest(leaveReq))
	}
	return &pb.ReadLeaveRequestsResponse{
		TotalCount: leaveRequestsCount,
		Data:       pbLeaveRequests,
	}, nil
}

func (s *LeaveServer) ReadMyLeaveRequests(ctx context.Context, req *pb.ReadMyLeaveRequestsRequest) (*pb.ReadMyLeaveRequestsResponse, error) {
	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "Forbidden: You don't have permission to read leave requests")
	}

	page := req.GetPage()
	if page == 0 {
		page = 1
	}

	perPage := req.GetPerPage()
	if perPage == 0 {
		perPage = 10
	}

	leaveRequests, leaveRequestsCount, leaveRequestsErr := s.LeaveStore.ReadMyLeaveRequests(contextData.UserID, page, perPage)
	if leaveRequestsErr != nil {
		return nil, status.Errorf(codes.Unknown, "error while fetching leave requests %v", leaveRequestsErr)
	}

	var pbLeaveRequests []*pb.LeaveRequest
	for _, leaveReq := range leaveRequests {
		pbLeaveRequests = append(pbLeaveRequests, LeaveRequestToPbLeaveRequest(leaveReq))
	}
	return &pb.ReadMyLeaveRequestsResponse{
		TotalCount: leaveRequestsCount,
		Data:       pbLeaveRequests,
	}, nil
}

func (s *LeaveServer) ApproveLeaveRequest(ctx context.Context, req *pb.ApproveLeaveRequestRequest) (*pb.ApproveLeaveRequestResponse, error) {
	if req.LeaveRequestId == "" {
		return nil, status.Error(codes.InvalidArgument, "Leave request ID is required")
	}

	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "Forbidden: You don't have permission to read leave requests")
	}

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	var contextChanData *RequestMetadata

	select {
	case contextChanData = <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	leaveRequest, err := s.LeaveStore.ReadLeaveRequestById(contextChanData, req.LeaveRequestId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "leave request not found")
	}

	opScope := &OpScope{
		Owner: OwnerType(leaveRequest.LeaveType.OwnerType),
		ID:    leaveRequest.LeaveType.OwnerID.String(),
	}

	if !CheckOpScope(contextChanData, opScope) {
		return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to access this resource")
	}

	if leaveRequest.Status != LEAVE_PENDING && leaveRequest.Status != LEAVE_APPEALED {
		return nil, status.Error(codes.InvalidArgument, "Leave request already processed")
	}

	leaveRequest.Status = LEAVE_APPROVED
	leaveRequest.ApproverComment = req.GetApproverComment()
	leaveRequest.ApproverID = contextData.UserID.String()

	updatedLeaveRequest, err := s.LeaveStore.UpdateLeaveRequest(leaveRequest)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "error while updating leave request %v", err)
	}

	leaveRequestEvent := LeaveRequestEvent{
		ID:             uuid.New(),
		LeaveRequestID: leaveRequest.ID,
		Action:         APPROVED,
		UserID:         contextData.UserID,
	}

	_, err = s.LeaveStore.CreateLeaveRequestEvent(&leaveRequestEvent)
	if err != nil {
		updatedLeaveRequest.LeaveRequestEvents = append(updatedLeaveRequest.LeaveRequestEvents, leaveRequestEvent)
	}

	err = s.LeaveStore.UseLeaveBalanceUnits(leaveRequest.UserID, leaveRequest.LeaveTypeID, leaveRequest.Duration)
	if err != nil {
		log.Println(status.Errorf(codes.Unknown, "error while using leave balance units %v", err))
	}

	return &pb.ApproveLeaveRequestResponse{
		Data: LeaveRequestToPbLeaveRequest(updatedLeaveRequest),
	}, nil
}

func (s *LeaveServer) RejectLeaveRequest(ctx context.Context, req *pb.RejectLeaveRequestRequest) (*pb.RejectLeaveRequestResponse, error) {
	if req.LeaveRequestId == "" {
		return nil, status.Error(codes.InvalidArgument, "Leave request ID is required")
	}

	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "Forbidden: You don't have permission to read leave requests")
	}

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	var contextChanData *RequestMetadata

	select {
	case contextChanData = <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	leaveRequest, err := s.LeaveStore.ReadLeaveRequestById(contextChanData, req.LeaveRequestId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "leave request not found")
	}

	opScope := &OpScope{
		Owner: OwnerType(leaveRequest.LeaveType.OwnerType),
		ID:    leaveRequest.LeaveType.OwnerID.String(),
	}

	if !CheckOpScope(contextChanData, opScope) {
		return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to access this resource")
	}

	if leaveRequest.Status != LEAVE_PENDING && leaveRequest.Status != LEAVE_APPEALED {
		return nil, status.Error(codes.InvalidArgument, "Leave request already processed")
	}

	leaveRequest.Status = LEAVE_REJECTED
	leaveRequest.ApproverComment = req.GetApproverComment()
	leaveRequest.ApproverID = contextData.UserID.String()

	updatedLeaveRequest, err := s.LeaveStore.UpdateLeaveRequest(leaveRequest)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "error while updating leave request %v", err)
	}

	leaveRequestEvent := LeaveRequestEvent{
		ID:             uuid.New(),
		LeaveRequestID: leaveRequest.ID,
		Action:         REJECTED,
		UserID:         contextData.UserID,
	}

	_, err = s.LeaveStore.CreateLeaveRequestEvent(&leaveRequestEvent)
	if err != nil {
		updatedLeaveRequest.LeaveRequestEvents = append(updatedLeaveRequest.LeaveRequestEvents, leaveRequestEvent)
	}

	return &pb.RejectLeaveRequestResponse{
		Data: LeaveRequestToPbLeaveRequest(updatedLeaveRequest),
	}, nil
}

func (s *LeaveServer) AppealLeaveRequest(ctx context.Context, req *pb.AppealLeaveRequestRequest) (*pb.AppealLeaveRequestResponse, error) {
	if req.LeaveRequestId == "" {
		return nil, status.Error(codes.InvalidArgument, "Leave request ID is required")
	}

	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "Forbidden: You don't have permission to read leave requests")
	}

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	var contextChanData *RequestMetadata

	select {
	case contextChanData = <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	leaveRequest, err := s.LeaveStore.ReadLeaveRequestById(contextChanData, req.LeaveRequestId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "leave request not found")
	}

	opScope := &OpScope{
		Owner: OwnerType(leaveRequest.LeaveType.OwnerType),
		ID:    leaveRequest.LeaveType.OwnerID.String(),
	}

	if !CheckOpScope(contextChanData, opScope) {
		return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to access this resource")
	}

	if leaveRequest.UserID.String() != contextData.UserID.String() {
		return nil, status.Error(codes.PermissionDenied, "Forbidden, You can't appeal a leave request for another user")
	}

	if leaveRequest.Status != LEAVE_REJECTED {
		return nil, status.Error(codes.InvalidArgument, "You can't appeal against this leave request")
	}

	leaveRequest.Status = LEAVE_APPEALED
	leaveRequest.AppealReason = req.GetAppealComment()

	updatedLeaveRequest, err := s.LeaveStore.UpdateLeaveRequest(leaveRequest)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "error while updating leave request %v", err)
	}

	leaveRequestEvent := LeaveRequestEvent{
		ID:             uuid.New(),
		LeaveRequestID: leaveRequest.ID,
		Action:         APPEALED,
		UserID:         contextData.UserID,
	}

	_, err = s.LeaveStore.CreateLeaveRequestEvent(&leaveRequestEvent)
	if err != nil {
		updatedLeaveRequest.LeaveRequestEvents = append(updatedLeaveRequest.LeaveRequestEvents, leaveRequestEvent)
	}

	return &pb.AppealLeaveRequestResponse{
		Data: LeaveRequestToPbLeaveRequest(updatedLeaveRequest),
	}, nil
}

func (s *LeaveServer) CancelLeaveRequest(ctx context.Context, req *pb.CancelLeaveRequestRequest) (*pb.CancelLeaveRequestResponse, error) {
	if req.LeaveRequestId == "" {
		return nil, status.Error(codes.InvalidArgument, "Leave request ID is required")
	}

	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "Forbidden: You don't have permission to read leave requests")
	}

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	var contextChanData *RequestMetadata

	select {
	case contextChanData = <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	leaveRequest, err := s.LeaveStore.ReadLeaveRequestById(contextChanData, req.LeaveRequestId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "leave request not found")
	}

	opScope := &OpScope{
		Owner: OwnerType(leaveRequest.LeaveType.OwnerType),
		ID:    leaveRequest.LeaveType.OwnerID.String(),
	}

	if !CheckOpScope(contextChanData, opScope) {
		return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to access this resource")
	}

	if leaveRequest.UserID.String() == contextData.UserID.String() {
		return nil, status.Error(codes.PermissionDenied, "Forbidden, You can't cancel your own leave request")
	}

	if leaveRequest.EndTime != "" && req.EndTime == "" {
		return nil, status.Error(codes.InvalidArgument, "End time cannot be empty when cancelling a leave request that had an end time")
	}

	if req.EndDate == "" {
		return nil, status.Error(codes.InvalidArgument, "End date cannot be empty when cancelling a leave request")
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid end date format")
	}

	prevEndDate, err := time.Parse(time.RFC3339, leaveRequest.EndDate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid previous end date format")
	}

	if prevEndDate.Before(endDate) {
		return nil, status.Error(codes.InvalidArgument, "New end date cannot be after previous end date when cancelling")
	}

	if leaveRequest.DurationUnit == HOUR && prevEndDate != endDate {
		return nil, status.Error(codes.InvalidArgument, "Cannot cancel a leave request with a duration unit of HOUR when the end date is different")
	}

	if leaveRequest.Status != LEAVE_ONGOING && leaveRequest.Status != LEAVE_APPROVED {
		return nil, status.Error(codes.InvalidArgument, "You cannot cancel this leave request")
	}

	duration, err := CalculateDuration(req.EndDate, leaveRequest.EndDate, req.EndTime, leaveRequest.EndTime, pb.TimeUnit(pb.TimeUnit_value[string(leaveRequest.DurationUnit)]))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	leaveRequest.Status = LEAVE_CANCELLED
	leaveRequest.ApproverComment = req.GetApproverComment()
	leaveRequest.EndDate = req.EndDate
	leaveRequest.EndTime = req.EndTime

	updatedLeaveRequest, err := s.LeaveStore.UpdateLeaveRequest(leaveRequest)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "error while updating leave request %v", err)
	}

	leaveRequestEvent := LeaveRequestEvent{
		ID:             uuid.New(),
		LeaveRequestID: leaveRequest.ID,
		Action:         CANCELLED,
		UserID:         contextData.UserID,
	}

	_, err = s.LeaveStore.CreateLeaveRequestEvent(&leaveRequestEvent)
	if err != nil {
		updatedLeaveRequest.LeaveRequestEvents = append(updatedLeaveRequest.LeaveRequestEvents, leaveRequestEvent)
	}

	err = s.LeaveStore.RevertLeaveBalanceUnits(leaveRequest.UserID, leaveRequest.LeaveTypeID, duration)
	if err != nil {
		log.Println(status.Errorf(codes.Unknown, "error while reverting leave balance units %v", err))
	}

	return &pb.CancelLeaveRequestResponse{
		Data: LeaveRequestToPbLeaveRequest(updatedLeaveRequest),
	}, nil
}

func (s *LeaveServer) EndLeaveRequest(ctx context.Context, req *pb.EndLeaveRequestRequest) (*pb.EndLeaveRequestResponse, error) {
	if req.LeaveRequestId == "" {
		return nil, status.Error(codes.InvalidArgument, "Leave request ID is required")
	}

	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "Forbidden: You don't have permission to read leave requests")
	}

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	var contextChanData *RequestMetadata

	select {
	case contextChanData = <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	leaveRequest, err := s.LeaveStore.ReadLeaveRequestById(contextChanData, req.LeaveRequestId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "leave request not found")
	}

	if leaveRequest.UserID.String() != contextData.UserID.String() {
		return nil, status.Error(codes.PermissionDenied, "Forbidden, You can't end a leave request that is not yours")
	}

	if leaveRequest.EndTime != "" && req.EndTime == "" {
		return nil, status.Error(codes.InvalidArgument, "End time cannot be empty when ending a leave request that had an end time")
	}

	if req.EndDate == "" {
		return nil, status.Error(codes.InvalidArgument, "End date cannot be empty when ending a leave request")
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid end date format")
	}

	prevEndDate, err := time.Parse(time.RFC3339, leaveRequest.EndDate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid previous end date format")
	}

	if prevEndDate.Before(endDate) {
		return nil, status.Error(codes.InvalidArgument, "New end date cannot be after previous end date when cancelling")
	}

	if leaveRequest.DurationUnit == HOUR && prevEndDate != endDate {
		return nil, status.Error(codes.InvalidArgument, "Cannot end a leave request with a duration unit of HOUR when the end date is different")
	}

	if leaveRequest.Status != LEAVE_ONGOING && leaveRequest.Status != LEAVE_APPROVED {
		return nil, status.Error(codes.InvalidArgument, "You cannot end this leave request")
	}

	duration, err := CalculateDuration(req.EndDate, leaveRequest.EndDate, req.EndTime, leaveRequest.EndTime, pb.TimeUnit(pb.TimeUnit_value[string(leaveRequest.DurationUnit)]))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	leaveRequest.Status = LEAVE_ENDED
	leaveRequest.EndDate = req.EndDate
	leaveRequest.EndTime = req.EndTime

	updatedLeaveRequest, err := s.LeaveStore.UpdateLeaveRequest(leaveRequest)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "error while updating leave request %v", err)
	}

	leaveRequestEvent := LeaveRequestEvent{
		ID:             uuid.New(),
		LeaveRequestID: leaveRequest.ID,
		Action:         ENDED,
		UserID:         contextData.UserID,
	}

	_, err = s.LeaveStore.CreateLeaveRequestEvent(&leaveRequestEvent)
	if err != nil {
		updatedLeaveRequest.LeaveRequestEvents = append(updatedLeaveRequest.LeaveRequestEvents, leaveRequestEvent)
	}

	err = s.LeaveStore.RevertLeaveBalanceUnits(leaveRequest.UserID, leaveRequest.LeaveTypeID, duration)
	if err != nil {
		log.Println(status.Errorf(codes.Unknown, "error while reverting leave balance units %v", err))
	}

	return &pb.EndLeaveRequestResponse{
		Data: LeaveRequestToPbLeaveRequest(updatedLeaveRequest),
	}, nil
}

func (s *LeaveServer) ReadLeaveBalance(ctx context.Context, req *pb.ReadLeaveBalanceRequest) (*pb.ReadLeaveBalanceResponse, error) {
	if req.UserId == "" || req.LeaveTypeId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID and leave type ID are required")
	}

	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "Forbidden: You don't have permission to read leave balance")
	}

	leaveBalance, err := s.LeaveStore.ReadLeaveBalance(uuid.MustParse(req.UserId), uuid.MustParse(req.LeaveTypeId))
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "error while fetching leave balance %v", err)
	}

	return &pb.ReadLeaveBalanceResponse{
		Data: LeaveBalanceToPbLeaveBalance(leaveBalance),
	}, nil
}

func (s *LeaveServer) ReadLeaveRequestsMetrics(ctx context.Context, req *pb.ReadLeaveRequestsMetricsRequest) (*pb.ReadLeaveRequestsMetricsResponse, error) {
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to access this resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)

	}
	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "Forbidden: You don't have permission to read leave requests metrics")
	}

	contextDataForLeaveTypes, err := s.AuthStore.AuthInternalRequests(contextData.Authorization, "READ_LEAVE_TYPES")
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}

	idsAllowingAccess := GetScopeIDs(contextDataForLeaveTypes.Scopes)

	var leaveTypeIDs []uuid.UUID
	leaveTypes, leaveTypesCount, leaveTypeErr := s.LeaveStore.ReadLeaveTypes(idsAllowingAccess, 1, 100)
	if leaveTypeErr != nil {
		return nil, status.Errorf(codes.Unknown, "error while fetching leave types %v", leaveTypeErr)
	}

	for _, lt := range leaveTypes {
		leaveTypeIDs = append(leaveTypeIDs, lt.ID)
	}

	if leaveTypesCount > 100 {
		noOfFetches := (leaveTypesCount + 99) / 100 // Ensure rounding up

		for i := int32(1); i < noOfFetches; i++ {
			moreLeaveTypes, _, leaveTypeErr := s.LeaveStore.ReadLeaveTypes(idsAllowingAccess, i*100+1, 100)
			if leaveTypeErr != nil {
				return nil, status.Errorf(codes.Unknown, "error while fetching leave types %v", leaveTypeErr)
			}
			for _, lt := range moreLeaveTypes {
				leaveTypeIDs = append(leaveTypeIDs, lt.ID)
			}
		}
	}

	leaveRequestsMetrics, err := s.LeaveStore.ReadLeaveRequestsMetrics(leaveTypeIDs)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "error while fetching leave requests metrics %v", err)
	}

	var pbLeaveRequestsMetrics []*pb.LeaveRequestsMetrics
	for _, leaveRequestsMetric := range leaveRequestsMetrics {
		pbLeaveRequestsMetrics = append(pbLeaveRequestsMetrics, &pb.LeaveRequestsMetrics{
			TotalCount: leaveRequestsMetric.TotalCount,
			Status:     pb.LeaveStatus(pb.LeaveStatus_value[string(leaveRequestsMetric.Status)]),
		})
	}

	return &pb.ReadLeaveRequestsMetricsResponse{
		Data: pbLeaveRequestsMetrics,
	}, nil
}
