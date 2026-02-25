package service

import (
	"kyla-be/pkg/pb"
	"log"

	"github.com/google/uuid"
)

// Shift conversions
func ShiftToPbShift(shift *Shift) *pb.Shift {
	users := []*User{}
	for _, user := range shift.ShiftUsers {
		users = append(users, &user)
	}

	return &pb.Shift{
		Id:                       shift.ID.String(),
		SerialNumber:             shift.SerialNumber,
		Name:                     shift.Name,
		Description:              shift.Description,
		StartTime:                shift.StartTime,
		EndTime:                  shift.EndTime,
		ShiftType:                shift.ShiftType,
		ShiftPeriod:              pb.ShiftPeriod(shift.ShiftPeriod),
		ShiftStatus:              shift.ShiftStatus,
		MinUsersPerShift:         int32(shift.MinUsersPerShift),
		MaxUsersPerShift:         int32(shift.MaxUsersPerShift),
		OvertimeAllowed:          shift.OverTimeAllowed,
		OvertimeLimit:            int32(shift.OverTimeLimit),
		OvertimeLimitUnit:        shift.OverTimeLimitUnit,
		SchedulingType:           pb.SchedulingType(shift.SchedulingType),
		HourlyRate:               shift.HourlyRate,
		HourlyRateCurrency:       shift.HourlyRateCurrency,
		HourlyRateCurrencySymbol: pb.CurrencySymbol(shift.HourlyRateCurrencySymbol),
		Users:                    UsersToPbUsers(users),
		Color:                    shift.Color,
		Banner:                   shift.Banner,
		OwnerType:                pb.OwnerType(pb.OwnerType_value[string(shift.OwnerType)]),
		OwnerId:                  shift.OwnerID.String(),
	}
}

func PbShiftToShift(pbShift *pb.Shift) *Shift {
	id, err := uuid.Parse(pbShift.Id)
	if err != nil {
		id = uuid.New()
	}
	users := []*User{}
	for _, user := range pbShift.Users {
		users = append(users, PbUserToUser(user))
	}

	return &Shift{
		ID:                       id,
		SerialNumber:             pbShift.SerialNumber,
		Name:                     pbShift.Name,
		Description:              pbShift.Description,
		StartTime:                pbShift.StartTime,
		EndTime:                  pbShift.EndTime,
		ShiftType:                pbShift.ShiftType,
		ShiftPeriod:              ShiftPeriod(pbShift.ShiftPeriod),
		ShiftStatus:              pbShift.ShiftStatus,
		MinUsersPerShift:         int(pbShift.MinUsersPerShift),
		MaxUsersPerShift:         int(pbShift.MaxUsersPerShift),
		OverTimeAllowed:          pbShift.OvertimeAllowed,
		OverTimeLimit:            int(pbShift.OvertimeLimit),
		OverTimeLimitUnit:        pbShift.OvertimeLimitUnit,
		SchedulingType:           SchedulingType(pbShift.SchedulingType),
		HourlyRate:               pbShift.HourlyRate,
		HourlyRateCurrency:       pbShift.HourlyRateCurrency,
		HourlyRateCurrencySymbol: CurrencySymbol(pbShift.HourlyRateCurrencySymbol),
		ShiftUsers:               convertToUserSlice(users),
		Color:                    pbShift.Color,
		Banner:                   pbShift.Banner,
		OwnerType:                OwnerType(pbShift.OwnerType),
		OwnerID:                  uuid.MustParse(pbShift.OwnerId),
		ShiftSchedules:           PbShiftSchedulesToShiftSchedules(pbShift.Schedules),
	}
}

// Schedule conversions
func ShiftScheduleToPbShiftSchedule(schedule *ShiftSchedule) *pb.ShiftSchedule {
	breaks := []*pb.ScheduleBreak{}
	for _, b := range schedule.ScheduleBreaks {
		breaks = append(breaks, ScheduleBreakToPbScheduleBreak(&b))
	}

	return &pb.ShiftSchedule{
		Id:           schedule.ID.String(),
		SerialNumber: schedule.SerialNumber,
		ShiftId:      schedule.ShiftID.String(),
		StartDate:    schedule.StartDate,
		EndDate:      schedule.EndDate,
		StartTime:    schedule.StartTime,
		EndTime:      schedule.EndTime,
		MinUsers:     int32(schedule.MinUsers),
		MaxUsers:     int32(schedule.MaxUsers),
		Breaks:       breaks,
		ShiftStatus:  schedule.ShiftStatus,
		OwnerType:    pb.OwnerType(pb.OwnerType_value[string(schedule.OwnerType)]),
		OwnerId:      schedule.OwnerID.String(),
	}
}

func PbShiftScheduleToShiftSchedule(pbSchedule *pb.ShiftSchedule) *ShiftSchedule {
	id, err := uuid.Parse(pbSchedule.Id)
	if err != nil {
		log.Println("Error parsing schedule id")
		id = uuid.New()
	}
	shiftID, err := uuid.Parse(pbSchedule.ShiftId)
	if err != nil {
		log.Println("Error parsing shift id")
		shiftID = uuid.New()
	}

	breaks := []ScheduleBreak{}
	for _, b := range pbSchedule.Breaks {
		breaks = append(breaks, *PbScheduleBreakToScheduleBreak(b))
	}

	return &ShiftSchedule{
		ID:                id,
		SerialNumber:      pbSchedule.SerialNumber,
		ShiftID:           shiftID,
		StartDate:         pbSchedule.StartDate,
		EndDate:           pbSchedule.EndDate,
		StartTime:         pbSchedule.StartTime,
		EndTime:           pbSchedule.EndTime,
		MinUsers:          int(pbSchedule.MinUsers),
		MaxUsers:          int(pbSchedule.MaxUsers),
		ScheduleBreaks:    breaks,
		ShiftStatus:       pbSchedule.ShiftStatus,
		OwnerType:         OwnerType(pbSchedule.OwnerType),
		OwnerID:           uuid.MustParse(pbSchedule.OwnerId),
		ShiftScheduleType: ShiftScheduleType(pbSchedule.ScheduleType),
	}
}

func PbShiftSchedulesToShiftSchedules(schedules []*pb.ShiftSchedule) []ShiftSchedule {
	items := []ShiftSchedule{}
	for _, schedule := range schedules {
		items = append(items, *PbShiftScheduleToShiftSchedule(schedule))
	}
	return items
}

// Break conversions
func ScheduleBreakToPbScheduleBreak(breakItem *ScheduleBreak) *pb.ScheduleBreak {
	return &pb.ScheduleBreak{
		Id:                breakItem.ID.String(),
		SerialNumber:      breakItem.SerialNumber,
		ScheduleId:        breakItem.ShiftScheduleID.String(),
		StartTime:         breakItem.StartTime,
		EndTime:           breakItem.EndTime,
		BreakPeriod:       breakItem.BreakPeriod,
		BreakPeriodUnit:   pb.TimeUnits(breakItem.BreakPeriodUnit),
		MinUsersRequired:  int32(breakItem.MinUsersRequired),
		AgentStatusId:     breakItem.AgentStatusID.String(),
		AgentStatusChange: StatusChangeToPbStatusChange(breakItem.AgentStatusChange),
		OwnerType:         pb.OwnerType(pb.OwnerType_value[string(breakItem.OwnerType)]),
		OwnerId:           breakItem.OwnerID.String(),
		Partial:           breakItem.Partial,
		Mandatory:         breakItem.Mandatory,
	}
}

func PbScheduleBreakToScheduleBreak(pbBreak *pb.ScheduleBreak) *ScheduleBreak {
	id, err := uuid.Parse(pbBreak.Id)
	if err != nil {
		log.Println("Error parsing break id")
		id = uuid.New()
	}
	agentStatusId, err := uuid.Parse(pbBreak.AgentStatusId)
	if err != nil {
		log.Println("Error parsing agent status id")
	}
	scheduleID, err := uuid.Parse(pbBreak.ScheduleId)
	if err != nil {
		log.Println("Error parsing schedule id")
		scheduleID = uuid.New()
	}

	agentStatusChange := StatusChange{}
	if pbBreak.AgentStatusChange != nil {
		agentStatusChange = *PbStatusChangeToStatusChange(pbBreak.AgentStatusChange)
	}

	return &ScheduleBreak{
		ID:                id,
		SerialNumber:      pbBreak.SerialNumber,
		ShiftScheduleID:   scheduleID,
		StartTime:         pbBreak.StartTime,
		EndTime:           pbBreak.EndTime,
		BreakPeriod:       pbBreak.BreakPeriod,
		BreakPeriodUnit:   TimeUnits(pbBreak.BreakPeriodUnit),
		ShiftScheduleType: ShiftScheduleType(pbBreak.ShiftScheduleType),
		MinUsersRequired:  int(pbBreak.MinUsersRequired),
		AgentStatusID:     agentStatusId,
		AgentStatusChange: agentStatusChange,
		OwnerType:         OwnerType(pbBreak.OwnerType),
		OwnerID:           uuid.MustParse(pbBreak.OwnerId),
		Partial:           pbBreak.Partial,
		Mandatory:         pbBreak.Mandatory,
	}
}

// Helper functions
func convertToUserSlice(users []*User) []User {
	result := make([]User, len(users))
	for i, user := range users {
		result[i] = *user
	}
	return result
}

// UserShiftRecord
func PbUserShiftRecordToUserShiftRecord(pbRecord *pb.UserShiftRecord) *UserShiftRecord {
	id, err := uuid.Parse(pbRecord.Id)
	if err != nil {
		log.Println("Error parsing user shift record id")
		id = uuid.New()
	}
	shiftID, err := uuid.Parse(pbRecord.ShiftId)
	if err != nil {
		log.Println("Error parsing shift id")
		shiftID = uuid.New()
	}

	return &UserShiftRecord{
		ID:        id,
		UserID:    uuid.MustParse(pbRecord.UserId),
		ShiftID:   shiftID,
		StartTime: pbRecord.StartTime,
		EndTime:   pbRecord.EndTime,
		CreatedBy: pbRecord.CreatedBy,
		UpdatedBy: pbRecord.UpdatedBy,
	}
}

func UserShiftRecordToPbUserShiftRecord(record *UserShiftRecord) *pb.UserShiftRecord {
	return &pb.UserShiftRecord{
		Id:        record.ID.String(),
		UserId:    record.UserID.String(),
		ShiftId:   record.ShiftID.String(),
		StartTime: record.StartTime,
		EndTime:   record.EndTime,
		CreatedBy: record.CreatedBy,
		UpdatedBy: record.UpdatedBy,
		CreatedAt: record.CreatedAt.String(),
		UpdatedAt: record.UpdatedAt.String(),
	}
}

func PbBreakRecordToBreakRecord(pbBreak *pb.BreakRecord) *BreakRecord {
	id, err := uuid.Parse(pbBreak.Id)
	if err != nil {
		log.Println("Error parsing break record id")
		id = uuid.New()
	}
	shiftID, err := uuid.Parse(pbBreak.UserShiftRecordId)
	if err != nil {
		log.Println("Error parsing shift id")
		shiftID = uuid.New()
	}
	breakID, err := uuid.Parse(pbBreak.BreakId)
	if err != nil {
		log.Println("Error parsing break id")
		breakID = uuid.New()
	}

	return &BreakRecord{
		ID:                id,
		BreakID:           breakID,
		UserShiftRecordID: shiftID,
		StartTime:         pbBreak.StartTime,
		EndTime:           pbBreak.EndTime,
		CreatedBy:         pbBreak.CreatedBy,
		UpdatedBy:         pbBreak.UpdatedBy,
	}
}
