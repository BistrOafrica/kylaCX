package service

import (
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
)

func StatusToPbStatus(status *AgentStatus) *pb.AgentStatus {
	return &pb.AgentStatus{
		Id:            status.ID.String(),
		AgentId:       status.AgentID.String(),
		StatusChanges: StatusChangesToPbStatusChanges(status.StatusChanges),
	}
}
func PbStatusToStatus(pbStatus *pb.AgentStatus) *AgentStatus {
	id, err := uuid.Parse(pbStatus.Id)
	if err != nil {
		id = uuid.New()
	}
	agentID, err := uuid.Parse(pbStatus.AgentId)
	if err != nil {
		agentID = uuid.New()
	}
	return &AgentStatus{
		ID:            id,
		AgentID:       agentID,
		StatusChanges: PbStatusChangesToStatusChanges(pbStatus.StatusChanges),
	}
}

func StatusChangeToPbStatusChange(statusChange StatusChange) *pb.StatusChange {
	return &pb.StatusChange{
		Id:          statusChange.ID.String(),
		StatusType:  pb.StatusType(statusChange.StatusType),
		Description: statusChange.Description,
		StartTime:   statusChange.StartTime.String(),
		EndTime:     statusChange.EndTime.String(),
	}
}

func PbStatusChangeToStatusChange(pbStatusChange *pb.StatusChange) *StatusChange {
	startTime, _ := utils.ConvertStringToTime(pbStatusChange.StartTime)
	endTime, _ := utils.ConvertStringToTime(pbStatusChange.EndTime)
	id, err := uuid.Parse(pbStatusChange.Id)
	if err != nil {
		id = uuid.New()
	}
	return &StatusChange{
		ID:          id,
		StatusType:  AgentStatusChange(pbStatusChange.StatusType),
		Description: pbStatusChange.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		OwnerType:   OwnerType(pbStatusChange.OwnerType),
	}
}

func StatusChangesToPbStatusChanges(statusChanges []StatusChange) []*pb.StatusChange {
	var pbStatusChanges []*pb.StatusChange
	for _, statusChange := range statusChanges {
		pbStatusChanges = append(pbStatusChanges, StatusChangeToPbStatusChange(statusChange))
	}
	return pbStatusChanges
}

func PbStatusChangesToStatusChanges(pbStatusChanges []*pb.StatusChange) []StatusChange {
	var statusChanges []StatusChange

	for _, pbStatusChange := range pbStatusChanges {
		statusChanges = append(statusChanges, *PbStatusChangeToStatusChange(pbStatusChange))
	}
	return statusChanges
}
