package team

import (
	"kyla-be/internal/authctx"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
)

func TeamToPbTeam(team *Team) *pb.Team {
	return &pb.Team{
		Id:           team.ID.String(),
		SerialNumber: team.SerialNumber,
		Name:         team.Name,
		Description:  team.Description,
		OwnerId:      team.OwnerID.String(),
		OwnerType:    pb.OwnerType(pb.OwnerType_value[string(team.OwnerType)]),
		CreatedBy:    team.CreatedBy,
		UpdatedBy:    team.UpdatedBy,
		CreatedAt:    team.CreatedAt.String(),
		UpdatedAt:    team.UpdatedAt.String(),
	}
}

func PbTeamToTeam(team *pb.Team) *Team {
	id, err := uuid.Parse(team.Id)
	if err != nil {
		id = uuid.New()
	}
	return &Team{
		ID:           id,
		SerialNumber: utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["teams"], id.String()),
		Name:         team.Name,
		Description:  team.Description,
		OwnerID:      uuid.MustParse(team.OwnerId),
		OwnerType:    authctx.OwnerType(pb.OwnerType_name[int32(team.OwnerType)]),
		CreatedBy:    team.CreatedBy,
		UpdatedBy:    team.UpdatedBy,
	}
}

func PbTeamsToTeams(teams []*pb.Team) []*Team {
	var result []*Team
	for _, team := range teams {
		result = append(result, PbTeamToTeam(team))
	}
	return result
}

func TeamsToPbTeams(teams []*Team) []*pb.Team {
	var result []*pb.Team
	for _, team := range teams {
		result = append(result, TeamToPbTeam(team))
	}
	return result
}
