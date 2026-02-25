package service

import (
	"fmt"
	"kyla-be/pkg/pb"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LeaveTypeToPbLeaveType(leaveType *LeaveType) *pb.LeaveType {
	var pbEarnedLeaveConditions []*pb.EarnedLeaveCondition
	for _, c := range leaveType.EarnedLeaveConditions {
		pbEarnedLeaveConditions = append(pbEarnedLeaveConditions, EarnedLeaveConditionToPbEarnedLeaveCondition(&c))
	}

	return &pb.LeaveType{
		Id:                     leaveType.ID.String(),
		Name:                   leaveType.Name,
		Description:            leaveType.Description,
		RemunerationType:       pb.RemunerationType(pb.RemunerationType_value[string(leaveType.RemunerationType)]),
		EarnedType:             pb.EarnedType(pb.EarnedType_value[string(leaveType.EarnedType)]),
		EarnedLeaveConditions:  pbEarnedLeaveConditions,
		TimeLimit:              int32(leaveType.TimeLimit),
		TimeLimitUnit:          pb.TimeUnit(pb.TimeUnit_value[string(leaveType.TimeLimitUnit)]),
		TimeLimitCycle:         pb.TimeLimitCycle(pb.TimeLimitCycle_value[string(leaveType.TimeLimitCycle)]),
		ApplyToWorkingTimeOnly: leaveType.ApplyToWorkingTimeOnly,
		AccrualPolicy:          pb.AccrualPolicy(pb.AccrualPolicy_value[string(leaveType.AccrualPolicy)]),
		CreatedBy:              leaveType.CreatedBy,
		OwnerType:              pb.OwnerType(pb.OwnerType_value[string(leaveType.OwnerType)]),
		OwnerId:                leaveType.OwnerID.String(),
		CreatedAt:              leaveType.CreatedAt.Format(time.RFC3339),
	}
}

func PbLeaveTypeToLeaveType(pbLeaveType *pb.LeaveType) *LeaveType {
	id, err := uuid.Parse(pbLeaveType.Id)
	if err != nil {
		log.Println("Error parsing leaveType id")
		id = uuid.New()
	}

	var earnedLeaveConditions []EarnedLeaveCondition
	for _, c := range pbLeaveType.EarnedLeaveConditions {
		earnedLeaveConditions = append(earnedLeaveConditions, *PbEarnedLeaveConditionToEarnedLeaveCondition(c))
	}

	return &LeaveType{
		ID:                     id,
		Name:                   pbLeaveType.Name,
		Description:            pbLeaveType.Description,
		RemunerationType:       MapRemunerationType(pbLeaveType.RemunerationType),
		EarnedType:             MapEarnedType(pbLeaveType.EarnedType),
		EarnedLeaveConditions:  earnedLeaveConditions,
		TimeLimit:              int(pbLeaveType.TimeLimit),
		TimeLimitUnit:          MapTimeUnit(pbLeaveType.TimeLimitUnit),
		TimeLimitCycle:         MapTimeLimitCycle(pbLeaveType.TimeLimitCycle),
		ApplyToWorkingTimeOnly: pbLeaveType.ApplyToWorkingTimeOnly,
		AccrualPolicy:          MapAccrualPolicy(pbLeaveType.AccrualPolicy),
	}
}

func EarnedLeaveConditionToPbEarnedLeaveCondition(earnedLeaveCondition *EarnedLeaveCondition) *pb.EarnedLeaveCondition {
	return &pb.EarnedLeaveCondition{
		Id:              earnedLeaveCondition.ID.String(),
		LeaveTypeId:     earnedLeaveCondition.LeaveTypeID.String(),
		WorkingTime:     int32(earnedLeaveCondition.WorkingTime),
		WorkingTimeUnit: pb.TimeUnit(pb.TimeUnit_value[string(earnedLeaveCondition.WorkingTimeUnit)]),
		RewardTime:      int32(earnedLeaveCondition.RewardTime),
		RewardTimeUnit:  pb.TimeUnit(pb.TimeUnit_value[string(earnedLeaveCondition.RewardTimeUnit)]),
	}
}

func PbEarnedLeaveConditionToEarnedLeaveCondition(pbEarnedLeaveCondition *pb.EarnedLeaveCondition) *EarnedLeaveCondition {
	id, err := uuid.Parse(pbEarnedLeaveCondition.Id)
	if err != nil {
		log.Println("Error parsing earnedLeaveCondition id")
		id = uuid.New()
	}
	leaveTypeID, err := uuid.Parse(pbEarnedLeaveCondition.LeaveTypeId)
	if err != nil {
		log.Println("Error parsing leaveTypeID id")
		leaveTypeID = uuid.New()
	}

	return &EarnedLeaveCondition{
		ID:              id,
		LeaveTypeID:     leaveTypeID,
		WorkingTime:     int(pbEarnedLeaveCondition.WorkingTime),
		WorkingTimeUnit: MapTimeUnit(pbEarnedLeaveCondition.WorkingTimeUnit),
		RewardTime:      int(pbEarnedLeaveCondition.RewardTime),
		RewardTimeUnit:  MapTimeUnit(pbEarnedLeaveCondition.RewardTimeUnit),
	}
}

func LeaveRequestToPbLeaveRequest(leaveRequest *LeaveRequest) *pb.LeaveRequest {
	var pbLeaveRequestAttachments []*pb.LeaveRequestAttachment
	for _, a := range leaveRequest.LeaveRequestAttachments {
		pbLeaveRequestAttachments = append(pbLeaveRequestAttachments, LeaveRequestAttachmentToPbLeaveRequestAttachment(&a))
	}

	var pbLeaveRequestEvents []*pb.LeaveRequestEvent
	for _, e := range leaveRequest.LeaveRequestEvents {
		pbLeaveRequestEvents = append(pbLeaveRequestEvents, LeaveRequestEventToPbLeaveRequestEvent(&e))
	}

	return &pb.LeaveRequest{
		Id:                      leaveRequest.ID.String(),
		UserId:                  leaveRequest.UserID.String(),
		User:                    UserToPbUser(&leaveRequest.User),
		LeaveTypeId:             leaveRequest.LeaveTypeID.String(),
		LeaveType:               LeaveTypeToPbLeaveType(&leaveRequest.LeaveType),
		StartTime:               leaveRequest.StartTime,
		EndTime:                 leaveRequest.EndTime,
		StartDate:               leaveRequest.StartDate,
		EndDate:                 leaveRequest.EndDate,
		Duration:                int32(leaveRequest.Duration),
		DurationUnit:            pb.TimeUnit(pb.TimeUnit_value[string(leaveRequest.DurationUnit)]),
		ApplierComment:          leaveRequest.ApplierComment,
		Status:                  pb.LeaveStatus(pb.LeaveStatus_value[string(leaveRequest.Status)]),
		LeaveRequestAttachments: pbLeaveRequestAttachments,
		ApproverId:              leaveRequest.ApproverID,
		ApproverComment:         leaveRequest.ApproverComment,
		CreatedAt:               leaveRequest.CreatedAt.Format(time.RFC3339),
		UpdatedAt:               leaveRequest.UpdatedAt.Format(time.RFC3339),
		CreatedBy:               leaveRequest.CreatedBy,
		IsRetrospective:         leaveRequest.IsRetrospective,
		AppealReason:            leaveRequest.AppealReason,
		LeaveRequestEvents:      pbLeaveRequestEvents,
	}
}

func PbLeaveRequesttoLeaveRequest(pbLeaveRequest *pb.LeaveRequest) *LeaveRequest {
	id, err := uuid.Parse(pbLeaveRequest.Id)
	if err != nil {
		id = uuid.New()
	}

	var leaveRequestAttachments []LeaveRequestAttachment
	for _, a := range pbLeaveRequest.LeaveRequestAttachments {
		a.LeaveRequestId = id.String()
		leaveRequestAttachments = append(leaveRequestAttachments, *PbLeaveRequestAttachmentToLeaveRequestAttachment(a))
	}

	return &LeaveRequest{
		ID:                      id,
		UserID:                  uuid.MustParse(pbLeaveRequest.UserId),
		LeaveTypeID:             uuid.MustParse(pbLeaveRequest.LeaveTypeId),
		StartTime:               pbLeaveRequest.StartTime,
		EndTime:                 pbLeaveRequest.EndTime,
		StartDate:               pbLeaveRequest.StartDate,
		EndDate:                 pbLeaveRequest.EndDate,
		Duration:                int(pbLeaveRequest.Duration),
		DurationUnit:            MapTimeUnit(pbLeaveRequest.DurationUnit),
		ApplierComment:          pbLeaveRequest.ApplierComment,
		LeaveRequestAttachments: leaveRequestAttachments,
		ApproverID:              pbLeaveRequest.ApproverId,
		ApproverComment:         pbLeaveRequest.ApproverComment,
		IsRetrospective:         pbLeaveRequest.IsRetrospective,
		AppealReason:            pbLeaveRequest.AppealReason,
	}
}

func LeaveRequestAttachmentToPbLeaveRequestAttachment(leaveRequestAttachment *LeaveRequestAttachment) *pb.LeaveRequestAttachment {
	return &pb.LeaveRequestAttachment{
		Id:             leaveRequestAttachment.ID.String(),
		LeaveRequestId: leaveRequestAttachment.LeaveRequestID.String(),
		FileName:       leaveRequestAttachment.FileName,
		FileUrl:        leaveRequestAttachment.FileUrl,
	}
}

func PbLeaveRequestAttachmentToLeaveRequestAttachment(pbLeaveRequestAttachment *pb.LeaveRequestAttachment) *LeaveRequestAttachment {
	id, err := uuid.Parse(pbLeaveRequestAttachment.Id)
	if err != nil {
		id = uuid.New()
	}

	return &LeaveRequestAttachment{
		ID:             id,
		LeaveRequestID: uuid.MustParse(pbLeaveRequestAttachment.LeaveRequestId),
		FileName:       pbLeaveRequestAttachment.FileName,
		FileUrl:        pbLeaveRequestAttachment.FileUrl,
	}
}

func LeaveRequestEventToPbLeaveRequestEvent(leaveRequestEvent *LeaveRequestEvent) *pb.LeaveRequestEvent {
	return &pb.LeaveRequestEvent{
		Id:             leaveRequestEvent.ID.String(),
		LeaveRequestId: leaveRequestEvent.LeaveRequestID.String(),
		Action:         string(leaveRequestEvent.Action),
		UserId:         leaveRequestEvent.UserID.String(),
		User:           UserToPbUser(&leaveRequestEvent.User),
		CreatedAt:      leaveRequestEvent.CreatedAt.Format(time.RFC3339),
	}
}

func LeaveBalanceToPbLeaveBalance(leaveBalance *LeaveBalance) *pb.LeaveBalance {
	return &pb.LeaveBalance{
		Id:            leaveBalance.ID.String(),
		UserId:        leaveBalance.UserID.String(),
		User:          UserToPbUser(&leaveBalance.User),
		LeaveTypeId:   leaveBalance.LeaveTypeID.String(),
		LeaveType:     LeaveTypeToPbLeaveType(&leaveBalance.LeaveType),
		TotalEligible: int32(leaveBalance.TotalEligible),
		Used:          int32(leaveBalance.Used),
		Remaining:     int32(leaveBalance.Remaining),
		BalanceUnit:   pb.TimeUnit(pb.TimeUnit_value[string(leaveBalance.BalanceUnit)]),
	}
}

func CalculateDuration(startDateStr, endDateStr, startTimeStr, endTimeStr string, unit pb.TimeUnit) (int, error) {
	startDate, err := ParseDateOnly(startDateStr)
	if err != nil {
		return 0, fmt.Errorf("invalid start date format")
	}

	endDate, err := ParseDateOnly(endDateStr)
	if err != nil {
		return 0, fmt.Errorf("invalid end date format")
	}

	// Calculate duration in days (inclusive of both start and end date)
	if unit == pb.TimeUnit_DAY {
		duration := int(endDate.Sub(startDate).Hours()/24) + 1
		if duration < 1 {
			return 0, fmt.Errorf("end date cannot be before start date")
		}
		return duration, nil
	}

	// Calculate duration in hours if the unit is hour
	if unit == pb.TimeUnit_HOUR {
		// Parse StartTime and EndTime (Expecting HH:MM format)
		startTime, err := time.Parse("15:04", startTimeStr)
		if err != nil {
			return 0, fmt.Errorf("invalid start time format")
		}

		endTime, err := time.Parse("15:04", endTimeStr)
		if err != nil {
			return 0, fmt.Errorf("invalid end time format")
		}

		// Combine Date and Time to create full DateTime objects
		startDateTime := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
		endDateTime := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)

		// Ensure EndTime is not before StartTime
		if endDateTime.Before(startDateTime) {
			return 0, fmt.Errorf("end time cannot be before start time")
		}

		// Calculate duration in hours
		duration := int(endDateTime.Sub(startDateTime).Hours())
		return duration, nil
	}

	return 0, fmt.Errorf("invalid duration unit")
}

func ValidateLeaveTypeFields(pbLeaveType *pb.LeaveType) (err error) {
	if pbLeaveType.GetName() == "" {
		return status.Error(codes.InvalidArgument, "Leave type name cannot be empty")
	}

	if pbLeaveType.GetRemunerationType() == pb.RemunerationType_REMUNERATION_TYPE_UNSPECIFIED {
		return status.Error(codes.InvalidArgument, "Remuneration type cannot be empty")
	}

	if pbLeaveType.GetEarnedType() == pb.EarnedType_EARNED_TYPE_UNSPECIFIED {
		return status.Error(codes.InvalidArgument, "Earned type cannot be empty")
	}

	if pbLeaveType.GetEarnedType() == pb.EarnedType_UPFRONT && (pbLeaveType.GetTimeLimit() == 0 || pbLeaveType.GetTimeLimitCycle() == pb.TimeLimitCycle_TIME_LIMIT_CYCLE_UNSPECIFIED) {
		return status.Error(codes.InvalidArgument, "Time limit and cycle cannot be empty for upfront leave type")
	}

	if pbLeaveType.GetEarnedType() == pb.EarnedType_EARNED && len(pbLeaveType.EarnedLeaveConditions) == 0 {
		return status.Error(codes.InvalidArgument, "Earned leave conditions cannot be empty for earned leave type")
	}

	if pbLeaveType.GetAccrualPolicy() == pb.AccrualPolicy_ACCRUAL_POLICY_UNSPECIFIED {
		return status.Error(codes.InvalidArgument, "Accrual policy cannot be empty")
	}

	if pbLeaveType.GetTimeLimitUnit() == pb.TimeUnit_TIME_UNIT_UNSPECIFIED {
		return status.Error(codes.InvalidArgument, "Time unit cannot be empty")
	}

	return nil
}

func ParseDateOnly(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr) // Ensure no extra spaces

	formats := []string{
		"2006-01-02T15:04:05.000Z", // With milliseconds
		"2006-01-02T15:04:05Z",     // Without milliseconds
		time.RFC3339,               // Standard format
	}

	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
		log.Printf("Failed to parse date %s with format %s: %v", dateStr, format, err)
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
}

func ParseTimeOnly(timeStr string) (time.Time, error) {
	// Parse as 24-hour time (HH:MM)
	return time.Parse("15:04", timeStr)
}

func CombineDateTime(date time.Time, timeStr string) (time.Time, error) {
	parsedTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time format: %w", err)
	}

	return time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		parsedTime.Hour(),
		parsedTime.Minute(),
		0, 0, // seconds, nanoseconds
		date.Location(),
	), nil
}

func ValidateLeaveRequestFields(pbLeaveRequest *pb.LeaveRequest) (err error) {
	if pbLeaveRequest.GetUserId() == "" {
		return status.Error(codes.InvalidArgument, "User ID cannot be empty")
	}

	if pbLeaveRequest.GetLeaveTypeId() == "" {
		return status.Error(codes.InvalidArgument, "Leave type ID cannot be empty")
	}

	if pbLeaveRequest.GetDurationUnit() == pb.TimeUnit(pb.TimeUnit_value[""]) {
		return status.Error(codes.InvalidArgument, "Duration unit cannot be empty")
	}

	startDate, err := ParseDateOnly(pbLeaveRequest.GetStartDate())
	if err != nil {
		return status.Error(codes.InvalidArgument, "Invalid start date format")
	}
	endDate, err := ParseDateOnly(pbLeaveRequest.GetEndDate())
	if err != nil {
		return status.Error(codes.InvalidArgument, "Invalid end date format")
	}

	if startDate.UnixNano() < time.Now().UnixNano() && !pbLeaveRequest.GetIsRetrospective() {
		return status.Error(codes.InvalidArgument, "Start date cannot be in the past")
	}

	if endDate.Before(startDate) {
		return status.Error(codes.InvalidArgument, "end date cannot be older than start date")
	}

	if pbLeaveRequest.GetDurationUnit() == pb.TimeUnit_HOUR && (pbLeaveRequest.GetStartTime() == "" || pbLeaveRequest.GetEndTime() == "") {
		return status.Error(codes.InvalidArgument, "Start time and end time are required for hour duration")
	}

	if pbLeaveRequest.GetDurationUnit() == pb.TimeUnit_HOUR && pbLeaveRequest.GetStartDate() != pbLeaveRequest.GetEndDate() {
		return status.Error(codes.InvalidArgument, "Time unit HOUR requires start and end dates to be the same")
	}

	if pbLeaveRequest.GetStartTime() != "" {
		if pbLeaveRequest.GetEndTime() == "" {
			return status.Error(codes.InvalidArgument, "End time is required if start time is provided")
		}

		startTime, err := ParseTimeOnly(pbLeaveRequest.GetStartTime())
		if err != nil {
			return status.Error(codes.InvalidArgument, "Invalid start time format")
		}

		endTime, err := ParseTimeOnly(pbLeaveRequest.GetEndTime())
		if err != nil {
			return status.Error(codes.InvalidArgument, "Invalid end time format")
		}

		if endTime.Before(startTime) {
			return status.Error(codes.InvalidArgument, "End time cannot be before start time")
		}

		if !pbLeaveRequest.GetIsRetrospective() {
			startDateTime, err := CombineDateTime(startDate, pbLeaveRequest.GetStartTime())
			if err != nil {
				return status.Error(codes.InvalidArgument, err.Error())
			}

			if startDateTime.Before(time.Now()) {
				return status.Error(
					codes.InvalidArgument,
					fmt.Sprintf("Start date/time cannot be in the past (was %s, now is %s)",
						startDateTime.Format("2006-01-02 15:04"),
						time.Now().Format("2006-01-02 15:04")),
				)
			}
		}
	}

	return nil
}

func MapRemunerationType(pbType pb.RemunerationType) RemunerationType {
	switch pbType {
	case pb.RemunerationType_PAID:
		return PAID
	case pb.RemunerationType_UNPAID:
		return UNPAID
	case pb.RemunerationType_HALF_PAID:
		return HALF_PAID
	default:
		return ""
	}
}

func MapEarnedType(pbType pb.EarnedType) EarnedType {
	switch pbType {
	case pb.EarnedType_UPFRONT:
		return UPFRONT
	case pb.EarnedType_EARNED:
		return EARNED
	default:
		return ""
	}
}

func MapTimeUnit(pbType pb.TimeUnit) TimeUnit {
	switch pbType {
	case pb.TimeUnit_DAY:
		return DAY
	case pb.TimeUnit_HOUR:
		return HOUR
	default:
		return ""
	}
}

func MapTimeLimitCycle(pbType pb.TimeLimitCycle) TimeLimitCycle {
	switch pbType {
	case pb.TimeLimitCycle_ANNUALLY:
		return ANNUALLY
	case pb.TimeLimitCycle_SEMI_ANNUALLY:
		return SEMI_ANNUALLY
	case pb.TimeLimitCycle_QUARTER_ANNUALLY:
		return QUARTER_ANNUALLY
	case pb.TimeLimitCycle_MONTHLY:
		return MONTHLY
	case pb.TimeLimitCycle_PER_WEEK:
		return PER_WEEK
	default:
		return ""
	}
}

func MapAccrualPolicy(pbType pb.AccrualPolicy) AccrualPolicy {
	switch pbType {
	case pb.AccrualPolicy_CARRY_FORWARD:
		return CARRY_FORWARD
	case pb.AccrualPolicy_EXPIRE:
		return EXPIRE
	default:
		return ""
	}
}
