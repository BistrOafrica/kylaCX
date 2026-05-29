package queues

import (
	"context"
	"log"
	"time"

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

// WatchQueueEntries is a server-side streaming RPC that pushes a fresh
// snapshot whenever the queue's entry set changes. Uses a polling tick under
// the hood — keeps the implementation simple and provider-agnostic (no NATS
// fan-out required). Clamped to [500ms, 10s] regardless of the requested
// interval.
//
// The first message is always sent (so clients see the initial state); after
// that we send only when the snapshot changed, plus a heartbeat every 15s
// so the connection stays open through gRPC-web proxies.
func (s *Server) WatchQueueEntries(req *pb.WatchQueueEntriesRequest, stream pb.QueueService_WatchQueueEntriesServer) error {
	if _, err := s.requireAuth(stream.Context()); err != nil {
		return err
	}
	if req.GetQueueId() == "" {
		return status.Error(codes.InvalidArgument, "queue_id is required")
	}

	interval := time.Duration(req.GetIntervalMs()) * time.Millisecond
	if interval < 500*time.Millisecond {
		interval = 500 * time.Millisecond
	}
	if interval > 10*time.Second {
		interval = 10 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	// previous signature is the concatenation of (entry.id, status) — cheap
	// to compute and stable for change detection.
	var previousSig string
	first := true

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			rows, err := s.store.ListLiveEntries(req.GetQueueId(), "")
			if err != nil {
				// Don't tear the stream down on a transient DB error; log and
				// retry on the next tick.
				log.Printf("[queues] watch entries DB error: %v", err)
				continue
			}
			sig := snapshotSignature(rows)
			changed := sig != previousSig
			if !changed && !first {
				continue
			}
			previousSig = sig
			first = false
			out := make([]*pb.QueueEntry, 0, len(rows))
			for _, r := range rows {
				out = append(out, EntryToPb(r))
			}
			if err := stream.Send(&pb.WatchQueueEntriesUpdate{
				Entries: out,
				Changed: changed,
			}); err != nil {
				return err
			}
		case <-heartbeat.C:
			// Send an empty changed=false update so gRPC-web / Envoy proxies
			// don't time the connection out during quiet periods.
			if err := stream.Send(&pb.WatchQueueEntriesUpdate{Changed: false}); err != nil {
				return err
			}
		}
	}
}

// snapshotSignature builds a stable change-detection key from the entry list.
// status + id covers waiting→ringing→connected transitions and entries
// entering/leaving the live set; entered_at differences alone wouldn't be
// caught but those are not user-visible state changes.
func snapshotSignature(rows []*Entry) string {
	if len(rows) == 0 {
		return ""
	}
	const sep = "|"
	out := make([]byte, 0, len(rows)*40)
	for _, r := range rows {
		out = append(out, r.ID...)
		out = append(out, ':')
		out = append(out, r.Status...)
		out = append(out, sep...)
	}
	return string(out)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func (s *Server) requireAuth(ctx context.Context) (*authctx.RequestMetadata, error) {
	md, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || md.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}
	return md, nil
}
