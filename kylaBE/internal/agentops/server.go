package agentops

import (
	"context"
	"kyla-be/internal/authctx"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthGateway defines the auth methods AgentStatusServer needs.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
}

type AgentStatusServer struct {
	pb.UnimplementedAgentStatusServiceServer
	AuthStore        AuthGateway
	AgentStatusStore *StatusStore
}

func NewAgentStatusServer(authStore AuthGateway, agentStatusStore *StatusStore) *AgentStatusServer {
	return &AgentStatusServer{
		AuthStore:        authStore,
		AgentStatusStore: agentStatusStore,
	}
}

func (s *AgentStatusServer) CreateStatusChange(ctx context.Context, req *pb.StatusChangeRequest) (*pb.StatusChangeResponse, error) {
	log.Println("Create Status Change request received")

	contextData, err := s.AuthStore.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Errorf(codes.PermissionDenied, "unauthorised request: %v", err)
	}

	agentStatus, err := s.AgentStatusStore.ReadAgentStatus(contextData.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read agent status: %v", err)
	}

	statusChange := PbStatusChangeToStatusChange(req.GetStatusChange())
	statusChange.ID = uuid.New()
	statusChange.AgentStatusID = agentStatus.ID

	if err := s.AgentStatusStore.SaveStatusChange(agentStatus, statusChange); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save status change: %v", err)
	}

	return &pb.StatusChangeResponse{
		StatusChange: StatusChangeToPbStatusChange(*statusChange),
	}, nil
}

func (s *AgentStatusServer) ReadLatestStatusChange(ctx context.Context, req *pb.AgentStatusRequest) (*pb.StatusChangeResponse, error) {
	log.Println("Read Latest Status Change request received")

	contextData, err := s.AuthStore.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Errorf(codes.PermissionDenied, "unauthorised request: %v", err)
	}

	agentStatus, err := s.AgentStatusStore.ReadAgentStatus(contextData.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read agent status: %v", err)
	}

	statusChange, err := s.AgentStatusStore.ReadLatestStatusChange(agentStatus.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read latest status change: %v", err)
	}

	return &pb.StatusChangeResponse{
		StatusChange: StatusChangeToPbStatusChange(*statusChange),
	}, nil
}

func (s *AgentStatusServer) ReadUserStatusHistory(ctx context.Context, req *pb.AgentStatusRequest) (*pb.AgentStatusResponse, error) {
	log.Println("Read User Status History request received")

	contextData, err := s.AuthStore.GetServiceAuthMetadata(ctx)
	if err != nil || contextData.RequestAuth == k.NewConsts().FALSE {
		return nil, status.Errorf(codes.PermissionDenied, "unauthorised request: %v", err)
	}

	agentStatus, err := s.AgentStatusStore.ReadAgentStatus(contextData.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read agent status: %v", err)
	}

	return &pb.AgentStatusResponse{
		AgentStatus: StatusToPbStatus(agentStatus),
	}, nil
}

func (s *AgentStatusServer) GetAgentAvailability(ctx context.Context, req *pb.AgentAvailabilityRequest) (*pb.AgentAvailabilityResponse, error) {
	log.Println("Get Agent Availability request received")

	agentID, err := uuid.Parse(req.GetAgentId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid agent ID: %v", err)
	}

	agentStatus, err := s.AgentStatusStore.ReadAgentStatus(agentID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read agent status: %v", err)
	}

	statusChange, err := s.AgentStatusStore.ReadLatestStatusChange(agentStatus.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read latest status change: %v", err)
	}

	return &pb.AgentAvailabilityResponse{
		CurrentStatus: StatusChangeToPbStatusChange(*statusChange),
	}, nil
}
