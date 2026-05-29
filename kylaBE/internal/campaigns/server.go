package campaigns

import (
	"context"
	"log"

	"kyla-be/internal/authctx"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthGateway is the subset of the auth stack the CampaignServer needs.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
}

// Launcher is the subset of the campaign Executor the gRPC server depends on.
// Defined as an interface so the server file doesn't pull in Temporal types.
type Launcher interface {
	Enabled() bool
	Launch(ctx context.Context, c *Campaign) (workflowID, scheduleID string, err error)
	Pause(ctx context.Context, c *Campaign) error
	Cancel(ctx context.Context, c *Campaign) error
}

// Server implements pb.CampaignServiceServer.
type Server struct {
	store    *Store
	auth     AuthGateway
	launcher Launcher
	pb.UnimplementedCampaignServiceServer
}

// NewServer wires the campaign service. launcher may be nil when Temporal is
// unavailable — lifecycle RPCs (Launch / Pause / Cancel) then return
// FailedPrecondition, but CRUD continues to work.
func NewServer(store *Store, auth AuthGateway, launcher Launcher) *Server {
	return &Server{store: store, auth: auth, launcher: launcher}
}

// ── Campaign CRUD ────────────────────────────────────────────────────────────

func (s *Server) CreateCampaign(ctx context.Context, req *pb.CreateCampaignRequest) (*pb.Campaign, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	in := req.GetCampaign()
	if in == nil || in.GetName() == "" || in.GetChannel() == "" {
		return nil, status.Error(codes.InvalidArgument, "name and channel are required")
	}
	c := CampaignFromProto(in)
	c.OrgID = md.OrganisationID.String()
	c.CreatedBy = md.UserID.String()
	// Force initial status — clients cannot pre-launch via Create.
	c.Status = string(StatusDraft)

	created, err := s.store.CreateCampaign(c)
	if err != nil {
		log.Printf("[campaigns] create: %v", err)
		return nil, status.Error(codes.Internal, "failed to create campaign")
	}
	return CampaignToProto(created), nil
}

func (s *Server) GetCampaign(ctx context.Context, req *pb.GetCampaignRequest) (*pb.Campaign, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	c, err := s.store.GetCampaign(req.GetId(), md.OrganisationID.String())
	if err != nil {
		return nil, status.Error(codes.NotFound, "campaign not found")
	}
	return CampaignToProto(c), nil
}

func (s *Server) ListCampaigns(ctx context.Context, req *pb.ListCampaignsRequest) (*pb.ListCampaignsResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	rows, total, err := s.store.ListCampaigns(ListCampaignsParams{
		OrgID:       md.OrganisationID.String(),
		WorkspaceID: req.GetWorkspaceId(),
		Status:      req.GetStatus(),
		Channel:     req.GetChannel(),
		PageSize:    int(req.GetPageSize()),
		PageToken:   req.GetPageToken(),
	})
	if err != nil {
		log.Printf("[campaigns] list: %v", err)
		return nil, status.Error(codes.Internal, "failed to list campaigns")
	}
	out := make([]*pb.Campaign, 0, len(rows))
	for _, r := range rows {
		out = append(out, CampaignToProto(r))
	}
	return &pb.ListCampaignsResponse{Campaigns: out, Total: total}, nil
}

func (s *Server) UpdateCampaign(ctx context.Context, req *pb.UpdateCampaignRequest) (*pb.Campaign, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	in := req.GetCampaign()
	if in == nil || in.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "campaign.id is required")
	}
	existing, err := s.store.GetCampaign(in.GetId(), md.OrganisationID.String())
	if err != nil {
		return nil, status.Error(codes.NotFound, "campaign not found")
	}
	// Disallow editing once launched — protects the integrity of an in-flight
	// audience snapshot. Pause + clone is the supported pattern for changes.
	if CampaignStatus(existing.Status) == StatusRunning {
		return nil, status.Error(codes.FailedPrecondition, "cannot edit a running campaign; pause or cancel first")
	}
	c := CampaignFromProto(in)
	c.OrgID = md.OrganisationID.String()
	updated, err := s.store.UpdateCampaign(c)
	if err != nil {
		log.Printf("[campaigns] update: %v", err)
		return nil, status.Error(codes.Internal, "failed to update campaign")
	}
	return CampaignToProto(updated), nil
}

func (s *Server) DeleteCampaign(ctx context.Context, req *pb.DeleteCampaignRequest) (*pb.DeleteCampaignResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	existing, err := s.store.GetCampaign(req.GetId(), md.OrganisationID.String())
	if err != nil {
		return nil, status.Error(codes.NotFound, "campaign not found")
	}
	if CampaignStatus(existing.Status) == StatusRunning {
		return nil, status.Error(codes.FailedPrecondition, "cannot delete a running campaign; cancel first")
	}
	if err := s.store.DeleteCampaign(req.GetId(), md.OrganisationID.String()); err != nil {
		log.Printf("[campaigns] delete: %v", err)
		return nil, status.Error(codes.Internal, "failed to delete campaign")
	}
	return &pb.DeleteCampaignResponse{Ok: true}, nil
}

// ── Lifecycle ────────────────────────────────────────────────────────────────

func (s *Server) LaunchCampaign(ctx context.Context, req *pb.LaunchCampaignRequest) (*pb.LaunchCampaignResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if s.launcher == nil || !s.launcher.Enabled() {
		return nil, status.Error(codes.FailedPrecondition, "campaign executor not configured (Temporal unavailable)")
	}
	c, err := s.store.GetCampaign(req.GetId(), md.OrganisationID.String())
	if err != nil {
		return nil, status.Error(codes.NotFound, "campaign not found")
	}
	switch CampaignStatus(c.Status) {
	case StatusRunning, StatusScheduled:
		return nil, status.Error(codes.FailedPrecondition, "campaign already launched")
	case StatusCompleted, StatusCancelled:
		return nil, status.Error(codes.FailedPrecondition, "campaign already finalised")
	}
	wfID, schedID, err := s.launcher.Launch(ctx, c)
	if err != nil {
		log.Printf("[campaigns] launch %s: %v", c.ID, err)
		return nil, status.Error(codes.Internal, "failed to launch campaign")
	}
	if err := s.store.SetTemporalIDs(c.ID, c.OrgID, wfID, schedID); err != nil {
		log.Printf("[campaigns] persist temporal ids %s: %v", c.ID, err)
	}
	nextStatus := StatusRunning
	sched, _ := DecodeSchedule(c.Schedule)
	if sched.Mode == ScheduleOnce || sched.Mode == ScheduleRecurring {
		nextStatus = StatusScheduled
	}
	if _, err := s.store.SetStatus(c.ID, c.OrgID, nextStatus); err != nil {
		log.Printf("[campaigns] set status %s: %v", c.ID, err)
	}
	return &pb.LaunchCampaignResponse{
		Id:                 c.ID,
		Status:             string(nextStatus),
		TemporalWorkflowId: wfID,
	}, nil
}

func (s *Server) PauseCampaign(ctx context.Context, req *pb.PauseCampaignRequest) (*pb.PauseCampaignResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if s.launcher == nil || !s.launcher.Enabled() {
		return nil, status.Error(codes.FailedPrecondition, "campaign executor not configured")
	}
	c, err := s.store.GetCampaign(req.GetId(), md.OrganisationID.String())
	if err != nil {
		return nil, status.Error(codes.NotFound, "campaign not found")
	}
	if err := s.launcher.Pause(ctx, c); err != nil {
		log.Printf("[campaigns] pause %s: %v", c.ID, err)
		return nil, status.Error(codes.Internal, "failed to pause campaign")
	}
	if _, err := s.store.SetStatus(c.ID, c.OrgID, StatusPaused); err != nil {
		log.Printf("[campaigns] set status paused %s: %v", c.ID, err)
	}
	return &pb.PauseCampaignResponse{Id: c.ID, Status: string(StatusPaused)}, nil
}

func (s *Server) CancelCampaign(ctx context.Context, req *pb.CancelCampaignRequest) (*pb.CancelCampaignResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if s.launcher == nil || !s.launcher.Enabled() {
		return nil, status.Error(codes.FailedPrecondition, "campaign executor not configured")
	}
	c, err := s.store.GetCampaign(req.GetId(), md.OrganisationID.String())
	if err != nil {
		return nil, status.Error(codes.NotFound, "campaign not found")
	}
	if err := s.launcher.Cancel(ctx, c); err != nil {
		log.Printf("[campaigns] cancel %s: %v", c.ID, err)
		return nil, status.Error(codes.Internal, "failed to cancel campaign")
	}
	if _, err := s.store.SetStatus(c.ID, c.OrgID, StatusCancelled); err != nil {
		log.Printf("[campaigns] set status cancelled %s: %v", c.ID, err)
	}
	return &pb.CancelCampaignResponse{Id: c.ID, Status: string(StatusCancelled)}, nil
}

// ── Recipients ───────────────────────────────────────────────────────────────

func (s *Server) ListRecipients(ctx context.Context, req *pb.ListRecipientsRequest) (*pb.ListRecipientsResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetCampaignId() == "" {
		return nil, status.Error(codes.InvalidArgument, "campaign_id is required")
	}
	// Org scoping via campaign ownership — cheaper than join-then-filter.
	if _, err := s.store.GetCampaign(req.GetCampaignId(), md.OrganisationID.String()); err != nil {
		return nil, status.Error(codes.NotFound, "campaign not found")
	}
	rows, total, err := s.store.ListRecipients(ListRecipientsParams{
		CampaignID: req.GetCampaignId(),
		Status:     req.GetStatus(),
		PageSize:   int(req.GetPageSize()),
		PageToken:  req.GetPageToken(),
	})
	if err != nil {
		log.Printf("[campaigns] list recipients: %v", err)
		return nil, status.Error(codes.Internal, "failed to list recipients")
	}
	out := make([]*pb.CampaignRecipient, 0, len(rows))
	for _, r := range rows {
		out = append(out, RecipientToProto(r))
	}
	return &pb.ListRecipientsResponse{Recipients: out, Total: total}, nil
}

// ── WhatsApp templates ──────────────────────────────────────────────────────

func (s *Server) UpsertWhatsAppTemplate(ctx context.Context, req *pb.UpsertWhatsAppTemplateRequest) (*pb.WhatsAppTemplate, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	in := req.GetTemplate()
	if in == nil || in.GetName() == "" || in.GetLanguage() == "" {
		return nil, status.Error(codes.InvalidArgument, "name and language are required")
	}
	t := TemplateFromProto(in)
	t.OrgID = md.OrganisationID.String()
	saved, err := s.store.UpsertTemplate(t)
	if err != nil {
		log.Printf("[campaigns] upsert template: %v", err)
		return nil, status.Error(codes.Internal, "failed to upsert template")
	}
	return TemplateToProto(saved), nil
}

func (s *Server) GetWhatsAppTemplate(ctx context.Context, req *pb.GetWhatsAppTemplateRequest) (*pb.WhatsAppTemplate, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	t, err := s.store.GetTemplate(req.GetId(), md.OrganisationID.String())
	if err != nil {
		return nil, status.Error(codes.NotFound, "template not found")
	}
	return TemplateToProto(t), nil
}

func (s *Server) ListWhatsAppTemplates(ctx context.Context, req *pb.ListWhatsAppTemplatesRequest) (*pb.ListWhatsAppTemplatesResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.store.ListTemplates(md.OrganisationID.String(), req.GetStatus(), req.GetPhoneNumberId())
	if err != nil {
		log.Printf("[campaigns] list templates: %v", err)
		return nil, status.Error(codes.Internal, "failed to list templates")
	}
	out := make([]*pb.WhatsAppTemplate, 0, len(rows))
	for _, r := range rows {
		out = append(out, TemplateToProto(r))
	}
	return &pb.ListWhatsAppTemplatesResponse{Templates: out}, nil
}

func (s *Server) DeleteWhatsAppTemplate(ctx context.Context, req *pb.DeleteWhatsAppTemplateRequest) (*pb.DeleteWhatsAppTemplateResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.store.DeleteTemplate(req.GetId(), md.OrganisationID.String()); err != nil {
		log.Printf("[campaigns] delete template: %v", err)
		return nil, status.Error(codes.Internal, "failed to delete template")
	}
	return &pb.DeleteWhatsAppTemplateResponse{Ok: true}, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func (s *Server) requireAuth(ctx context.Context) (*authctx.RequestMetadata, error) {
	md, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || md.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}
	return md, nil
}
