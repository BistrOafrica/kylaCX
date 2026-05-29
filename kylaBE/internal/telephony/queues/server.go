package queues

import (
	"context"
	"log"

	"kyla-be/internal/authctx"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthGateway is the subset of the auth stack the server depends on.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
}

// Server implements pb.QueueServiceServer.
type Server struct {
	store *Store
	auth  AuthGateway
	pb.UnimplementedQueueServiceServer
}

func NewServer(store *Store, auth AuthGateway) *Server {
	return &Server{store: store, auth: auth}
}

// ── Queue CRUD ───────────────────────────────────────────────────────────────

func (s *Server) CreateQueue(ctx context.Context, req *pb.QueueCreateRequest) (*pb.Queue, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	in := QueueFromPb(req.GetQueue())
	if in == nil || in.Name == "" || in.WorkspaceID == "" {
		return nil, status.Error(codes.InvalidArgument, "name and workspace_id are required")
	}
	in.OrgID = md.OrganisationID.String()
	created, err := s.store.CreateQueue(in)
	if err != nil {
		log.Printf("[queues] create: %v", err)
		return nil, status.Error(codes.Internal, "failed to create queue")
	}
	return QueueToPb(created), nil
}

func (s *Server) GetQueue(ctx context.Context, req *pb.QueueGetRequest) (*pb.Queue, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	q, err := s.store.GetQueue(req.GetId(), md.OrganisationID.String())
	if err != nil {
		return nil, status.Error(codes.NotFound, "queue not found")
	}
	return QueueToPb(q), nil
}

func (s *Server) ListQueues(ctx context.Context, req *pb.QueueListRequest) (*pb.QueueListResponse, error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	rows, err := s.store.ListQueues(req.GetWorkspaceId(), req.GetActiveOnly())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list queues")
	}
	out := make([]*pb.Queue, 0, len(rows))
	for _, r := range rows {
		out = append(out, QueueToPb(r))
	}
	return &pb.QueueListResponse{Queues: out}, nil
}

func (s *Server) UpdateQueue(ctx context.Context, req *pb.QueueUpdateRequest) (*pb.Queue, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	in := QueueFromPb(req.GetQueue())
	if in == nil || in.ID == "" {
		return nil, status.Error(codes.InvalidArgument, "queue.id is required")
	}
	in.OrgID = md.OrganisationID.String()
	updated, err := s.store.UpdateQueue(in)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update queue")
	}
	return QueueToPb(updated), nil
}

func (s *Server) DeleteQueue(ctx context.Context, req *pb.QueueDeleteRequest) (*pb.QueueDeleteResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.store.DeleteQueue(req.GetId(), md.OrganisationID.String()); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete queue")
	}
	return &pb.QueueDeleteResponse{Ok: true}, nil
}

// ── Members ─────────────────────────────────────────────────────────────────

func (s *Server) AddQueueMember(ctx context.Context, req *pb.AddQueueMemberRequest) (*pb.QueueMembership, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	in := MembershipFromPb(req.GetMember())
	if in == nil || in.QueueID == "" || in.UserID == "" {
		return nil, status.Error(codes.InvalidArgument, "queue_id and user_id are required")
	}
	in.OrgID = md.OrganisationID.String()
	created, err := s.store.AddMember(in)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to add member")
	}
	return MembershipToPb(created), nil
}

func (s *Server) RemoveQueueMember(ctx context.Context, req *pb.RemoveQueueMemberRequest) (*pb.RemoveQueueMemberResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.store.RemoveMember(req.GetId(), md.OrganisationID.String()); err != nil {
		return nil, status.Error(codes.Internal, "failed to remove member")
	}
	return &pb.RemoveQueueMemberResponse{Ok: true}, nil
}

func (s *Server) ListQueueMembers(ctx context.Context, req *pb.ListQueueMembersRequest) (*pb.ListQueueMembersResponse, error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	rows, err := s.store.ListMembers(req.GetQueueId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list members")
	}
	out := make([]*pb.QueueMembership, 0, len(rows))
	for _, r := range rows {
		out = append(out, MembershipToPb(r))
	}
	return &pb.ListQueueMembersResponse{Members: out}, nil
}

func (s *Server) SetQueueMemberActive(ctx context.Context, req *pb.SetQueueMemberActiveRequest) (*pb.SetQueueMemberActiveResponse, error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	updated, err := s.store.SetMemberActive(req.GetQueueId(), req.GetUserId(), req.GetIsActive())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to set active state")
	}
	return &pb.SetQueueMemberActiveResponse{Member: MembershipToPb(updated)}, nil
}

// ── Wallboard ────────────────────────────────────────────────────────────────

func (s *Server) ListQueueEntries(ctx context.Context, req *pb.ListQueueEntriesRequest) (*pb.ListQueueEntriesResponse, error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	rows, err := s.store.ListLiveEntries(req.GetQueueId(), req.GetStatus())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list entries")
	}
	out := make([]*pb.QueueEntry, 0, len(rows))
	for _, r := range rows {
		out = append(out, EntryToPb(r))
	}
	return &pb.ListQueueEntriesResponse{Entries: out}, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func (s *Server) requireAuth(ctx context.Context) (*authctx.RequestMetadata, error) {
	md, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || md.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}
	return md, nil
}
