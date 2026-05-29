package ivr

import (
	"context"
	"log"

	"kyla-be/internal/authctx"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthGateway is the subset of the auth stack this server needs. Local
// interface preserves the boundary-interface pattern used elsewhere.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
}

// Server implements pb.IVRServiceServer.
type Server struct {
	store *Store
	auth  AuthGateway
	pb.UnimplementedIVRServiceServer
}

func NewServer(store *Store, auth AuthGateway) *Server {
	return &Server{store: store, auth: auth}
}

// ── Flow CRUD ────────────────────────────────────────────────────────────────

func (s *Server) CreateIVRFlow(ctx context.Context, req *pb.CreateIVRFlowRequest) (*pb.IVRFlow, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	in := FlowFromPb(req.GetFlow())
	if in == nil || in.Name == "" || in.WorkspaceID == "" {
		return nil, status.Error(codes.InvalidArgument, "name and workspace_id are required")
	}
	in.OrgID = md.OrganisationID.String()
	in.CreatedBy = md.UserID.String()
	created, err := s.store.CreateFlow(in)
	if err != nil {
		log.Printf("[ivr] create flow: %v", err)
		return nil, status.Error(codes.Internal, "failed to create flow")
	}
	return FlowToPb(created), nil
}

func (s *Server) GetIVRFlow(ctx context.Context, req *pb.GetIVRFlowRequest) (*pb.IVRFlow, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	f, err := s.store.GetFlow(req.GetId(), md.OrganisationID.String())
	if err != nil {
		return nil, status.Error(codes.NotFound, "flow not found")
	}
	return FlowToPb(f), nil
}

func (s *Server) ListIVRFlows(ctx context.Context, req *pb.ListIVRFlowsRequest) (*pb.ListIVRFlowsResponse, error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	rows, err := s.store.ListFlows(req.GetWorkspaceId(), req.GetActiveOnly())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list flows")
	}
	out := make([]*pb.IVRFlow, 0, len(rows))
	for _, r := range rows {
		out = append(out, FlowToPb(r))
	}
	return &pb.ListIVRFlowsResponse{Flows: out}, nil
}

func (s *Server) UpdateIVRFlow(ctx context.Context, req *pb.UpdateIVRFlowRequest) (*pb.IVRFlow, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	in := FlowFromPb(req.GetFlow())
	if in == nil || in.ID == "" {
		return nil, status.Error(codes.InvalidArgument, "flow.id is required")
	}
	in.OrgID = md.OrganisationID.String()
	updated, err := s.store.UpdateFlow(in)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update flow")
	}
	return FlowToPb(updated), nil
}

func (s *Server) DeleteIVRFlow(ctx context.Context, req *pb.DeleteIVRFlowRequest) (*pb.DeleteIVRFlowResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.store.DeleteFlow(req.GetId(), md.OrganisationID.String()); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete flow")
	}
	return &pb.DeleteIVRFlowResponse{Ok: true}, nil
}

// ── DID mappings ────────────────────────────────────────────────────────────

func (s *Server) CreateIVRDIDMapping(ctx context.Context, req *pb.CreateIVRDIDMappingRequest) (*pb.IVRDIDMapping, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	in := DIDMappingFromPb(req.GetMapping())
	if in == nil || in.DID == "" || in.FlowID == "" {
		return nil, status.Error(codes.InvalidArgument, "did and flow_id are required")
	}
	in.OrgID = md.OrganisationID.String()
	created, err := s.store.CreateDIDMapping(in)
	if err != nil {
		log.Printf("[ivr] create did mapping: %v", err)
		return nil, status.Error(codes.Internal, "failed to create mapping")
	}
	return DIDMappingToPb(created), nil
}

func (s *Server) ListIVRDIDMappings(ctx context.Context, req *pb.ListIVRDIDMappingsRequest) (*pb.ListIVRDIDMappingsResponse, error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	rows, err := s.store.ListDIDMappings(req.GetWorkspaceId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list mappings")
	}
	out := make([]*pb.IVRDIDMapping, 0, len(rows))
	for _, r := range rows {
		out = append(out, DIDMappingToPb(r))
	}
	return &pb.ListIVRDIDMappingsResponse{Mappings: out}, nil
}

func (s *Server) DeleteIVRDIDMapping(ctx context.Context, req *pb.DeleteIVRDIDMappingRequest) (*pb.DeleteIVRDIDMappingResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.store.DeleteDIDMapping(req.GetId(), md.OrganisationID.String()); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete mapping")
	}
	return &pb.DeleteIVRDIDMappingResponse{Ok: true}, nil
}

// ── Runs ────────────────────────────────────────────────────────────────────

func (s *Server) GetIVRRun(ctx context.Context, req *pb.GetIVRRunRequest) (*pb.IVRRun, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	r, err := s.store.GetRun(req.GetId(), md.OrganisationID.String())
	if err != nil {
		return nil, status.Error(codes.NotFound, "run not found")
	}
	return RunToPb(r), nil
}

func (s *Server) ListIVRRuns(ctx context.Context, req *pb.ListIVRRunsRequest) (*pb.ListIVRRunsResponse, error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	rows, total, err := s.store.ListRuns(ListRunsParams{
		FlowID:    req.GetFlowId(),
		CallID:    req.GetCallId(),
		PageSize:  int(req.GetPageSize()),
		PageToken: req.GetPageToken(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list runs")
	}
	out := make([]*pb.IVRRun, 0, len(rows))
	for _, r := range rows {
		out = append(out, RunToPb(r))
	}
	return &pb.ListIVRRunsResponse{Runs: out, Total: total}, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func (s *Server) requireAuth(ctx context.Context) (*authctx.RequestMetadata, error) {
	md, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || md.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}
	return md, nil
}
