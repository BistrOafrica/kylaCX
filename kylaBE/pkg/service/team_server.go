package service

import (
	"context"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

type TeamServer struct {
	pb.UnimplementedTeamServiceServer
	TeamStore *TeamStore
	AuthStore *AuthStore
}

func NewTeamServer(teamStore *TeamStore, AuthStore *AuthStore) *TeamServer {
	return &TeamServer{
		TeamStore: teamStore,
		AuthStore: AuthStore,
	}
}

func (t *TeamServer) CreateTeam(ctx context.Context, request *pb.CreateTeamRequest) (*pb.CreateTeamResponse, error) {
	log.Println("CreateTeam")

	contextData, authErr := t.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(403, "Forbidden, You do not have permission to perform this action")
	}

	team := PbTeamToTeam(request.Team)
	team.ID = uuid.New()
	team.SerialNumber = utils.CreateSerialNumber(k.SERIAL_NUMBER_ABBR()["teams"], team.ID.String())
	team.CreatedBy = contextData.UserID.String()
	team.UpdatedBy = contextData.UserID.String()

	team.CREATE_TEAM_ROLES(t.AuthStore.RbacStore)
	team.ADD_CREATOR_TO_TEAM(t.AuthStore.UserStore)

	newTeam, teamErr := t.TeamStore.CreateTeam(team)
	if teamErr != nil {
		return nil, status.Error(500, "Internal Server Error, Failed to create team")
	}

	return &pb.CreateTeamResponse{
		Team: TeamToPbTeam(newTeam),
	}, nil
}

func (t *TeamServer) ReadTeam(ctx context.Context, request *pb.ReadTeamRequest) (*pb.ReadTeamResponse, error) {
	log.Println("ReadTeam")

	contextData, authErr := t.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(403, "Forbidden, You do not have permission to perform this action")
	}
	id := uuid.MustParse(request.GetId())

	team, teamErr := t.TeamStore.ReadTeam(&id)
	if teamErr != nil {
		return nil, status.Error(404, "Not Found, Team not found")
	}

	return &pb.ReadTeamResponse{
		Team: TeamToPbTeam(team),
	}, nil
}

func (t *TeamServer) ReadTeamsByUserID(ctx context.Context, request *pb.ReadTeamListRequest) (*pb.ReadTeamListResponse, error) {
	log.Println("ReadTeamList")

	contextData, authErr := t.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(403, "Forbidden, You do not have permission to perform this action")
	}

	teams, err := t.TeamStore.ReadTeamsByUserID(contextData.UserID.String())
	if err != nil {
		return nil, status.Error(404, "Not Found, Team not found")
	}

	return &pb.ReadTeamListResponse{
		Teams: TeamsToPbTeams(teams),
		Total: int32(len(teams)),
	}, nil
}

func (t *TeamServer) UpdateTeam(ctx context.Context, request *pb.UpdateTeamRequest) (*pb.UpdateTeamResponse, error) {
	log.Println("UpdateTeam")

	contextData, authErr := t.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(403, "Forbidden, You do not have permission to perform this action")
	}

	team := PbTeamToTeam(request.Team)
	team.UpdatedBy = contextData.UserID.String()

	updatedTeam, teamErr := t.TeamStore.UpdateTeam(team)
	if teamErr != nil {
		return nil, status.Error(500, "Internal Server Error, Failed to update team")
	}

	return &pb.UpdateTeamResponse{
		Team: TeamToPbTeam(updatedTeam),
	}, nil
}

func (t *TeamServer) DeleteTeam(ctx context.Context, request *pb.DeleteTeamRequest) (*pb.DeleteTeamResponse, error) {
	log.Println("DeleteTeam")
	scope := PbScopeToOpScope(request.GetScope())
	auth, _, err := t.AuthStore.ScopeCheck(ctx, scope.ID)
	if err != nil || auth == k.NewConsts().FALSE_BOOL {
		return nil, err
	}

	if err = t.TeamStore.DeleteTeam(uuid.MustParse(request.GetId())); err != nil {
		return nil, status.Error(500, "Internal Server Error, Failed to delete team")
	}

	return &pb.DeleteTeamResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Team deleted successfully",
		},
	}, nil
}

// AddUserToTeam
func (t *TeamServer) AddUserToTeam(ctx context.Context, request *pb.AddUserToTeamRequest) (*pb.AddUserToTeamResponse, error) {
	log.Println("AddUserToTeam")

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(request.GetScope())

	go t.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
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
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching break %v", err)
	}

	userID, parseErr := uuid.Parse(request.GetUserId())
	if parseErr != nil {
		return nil, status.Error(400, "Bad Request, Invalid user ID")
	}

	teamID, parseErr := uuid.Parse(request.GetTeamId())
	if parseErr != nil {
		return nil, status.Error(400, "Bad Request, Invalid team ID")
	}

	if err := t.TeamStore.AddUserToTeam(teamID, userID); err != nil {
		return nil, status.Error(500, "Internal Server Error, Failed to add user to team")
	}

	return &pb.AddUserToTeamResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "User added to team successfully",
		},
	}, nil

}

// RemoveUserFromTeam
func (t *TeamServer) RemoveUserFromTeam(ctx context.Context, request *pb.RemoveUserFromTeamRequest) (*pb.RemoveUserFromTeamResponse, error) {
	log.Println("RemoveUserFromTeam")

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(request.GetScope())

	go t.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
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
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching break %v", err)
	}

	userID, parseErr := uuid.Parse(request.GetUserId())
	if parseErr != nil {
		return nil, status.Error(400, "Bad Request, Invalid user ID")
	}

	teamID, parseErr := uuid.Parse(request.GetTeamId())
	if parseErr != nil {
		return nil, status.Error(400, "Bad Request, Invalid team ID")
	}

	if err := t.TeamStore.RemoveUserFromTeam(teamID, userID); err != nil {
		return nil, status.Error(500, "Internal Server Error, Failed to remove user from team")
	}
	return &pb.RemoveUserFromTeamResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "User removed from team successfully",
		},
	}, nil
}

// ReadTeamUsers

func (t *TeamServer) ReadTeamUsers(ctx context.Context, request *pb.ReadTeamUsersRequest) (*pb.ReadTeamUsersResponse, error) {
	log.Println("ReadTeamUsers")

	contextData, authErr := t.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(403, "Forbidden, You do not have permission to perform this action")
	}
	teamID, err := uuid.Parse(request.GetTeamId())
	if err != nil {
		return nil, status.Error(400, "Bad Request, Invalid team ID")
	}

	roleId := uuid.Nil
	if request.GetRoleId() != "" {
		roleId, err = uuid.Parse(request.GetRoleId())
		if err != nil {
			roleId = uuid.Nil
		}
	}

	users, err := t.TeamStore.ReadTeamUsers(teamID, roleId)
	if err != nil {
		return nil, status.Error(404, "Not Found, Users not found")
	}

	usersList := make([]*pb.User, 0)
	for _, user := range users {
		usersList = append(usersList, UserToPbUser(&user))
	}

	return &pb.ReadTeamUsersResponse{
		Users: usersList,
		Total: int32(len(users)),
	}, nil
}

func (t *TeamServer) ReadTeamList(ctx context.Context, request *pb.ReadTeamListRequest) (*pb.ReadTeamListResponse, error) {
	log.Println("ReadTeamList")
	scope := PbScopeToOpScope(request.GetScope())
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go t.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to read teams resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	teams, err := t.TeamStore.ReadTeams(scope)
	if err != nil {
		return nil, status.Error(404, "Not Found, Teams not found")
	}

	return &pb.ReadTeamListResponse{
		Teams: TeamsToPbTeams(teams),
	}, nil
}

func (t *TeamServer) ReadTeamsByorganisationId(ctx context.Context, request *pb.ReadTeamListRequest) (*pb.ReadTeamListResponse, error) {
	log.Println("ReadTeamsByorganisationId")
	scope := PbScopeToOpScope(request.GetScope())

	contextData, authErr := t.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(403, "Forbidden, You do not have permission to perform this action")
	}

	teams, err := t.TeamStore.ReadTeamsByOrganisationID(scope.ID)
	if err != nil {
		return nil, status.Error(404, "Not Found, Teams not found")
	}

	return &pb.ReadTeamListResponse{
		Teams: TeamsToPbTeams(teams),
		Total: int32(len(teams)),
	}, nil
}

func (t *TeamServer) ReadTeamsByBranchID(ctx context.Context, request *pb.ReadTeamListRequest) (*pb.ReadTeamListResponse, error) {
	log.Println("ReadTeamsByBranchID")
	scope := PbScopeToOpScope(request.GetScope())

	contextData, authErr := t.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(403, "Forbidden, You do not have permission to perform this action")
	}

	teams, err := t.TeamStore.ReadTeamsByBranchID(scope.ID)
	if err != nil {
		return nil, status.Error(404, "Not Found, Teams not found")
	}

	return &pb.ReadTeamListResponse{
		Teams: TeamsToPbTeams(teams),
		Total: int32(len(teams)),
	}, nil
}

func (t *TeamServer) ReadTeamsByDepartmentID(ctx context.Context, request *pb.ReadTeamListRequest) (*pb.ReadTeamListResponse, error) {
	log.Println("ReadTeamsByDepartmentID")
	scope := PbScopeToOpScope(request.GetScope())

	contextData, authErr := t.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Error(403, "Forbidden, You do not have permission to perform this action")
	}

	teams, err := t.TeamStore.ReadTeamsByDepartmentID(scope.ID)
	if err != nil {
		return nil, status.Error(404, "Not Found, Teams not found")
	}

	return &pb.ReadTeamListResponse{
		Teams: TeamsToPbTeams(teams),
		Total: int32(len(teams)),
	}, nil
}
