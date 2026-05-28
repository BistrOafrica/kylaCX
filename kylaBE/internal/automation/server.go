package automation

import (
	"context"
	"log"

	"kyla-be/internal/authctx"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/shared/events"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthGateway is the subset of the auth stack the WorkflowServer needs.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
	ScopeCheck(ctx context.Context, scopeID string) (bool, *authctx.RequestMetadata, error)
}

// Server implements pb.WorkflowServiceServer.
// Workflow execution is delegated to Temporal via Executor; this server
// manages definitions and exposes run-history projections.
type Server struct {
	store    *Store
	auth     AuthGateway
	executor *Executor // optional — nil when Temporal is unavailable
	pb.UnimplementedWorkflowServiceServer
}

// NewServer constructs a WorkflowService server.
// executor may be nil; in that case TestRunWorkflow returns FailedPrecondition.
func NewServer(store *Store, auth AuthGateway, executor *Executor) *Server {
	return &Server{store: store, auth: auth, executor: executor}
}

// ── Workflow CRUD ────────────────────────────────────────────────────────────

func (s *Server) CreateWorkflow(ctx context.Context, req *pb.CreateWorkflowRequest) (*pb.Workflow, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	w := &Workflow{
		OrgID:       md.OrganisationID.String(),
		WorkspaceID: req.GetWorkspaceId(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Status:      normaliseStatus(req.GetStatus()),
		Trigger:     triggerFromStruct(req.GetTrigger()),
		Conditions:  conditionsFromStructs(req.GetConditions()),
		Actions:     actionsFromStructs(req.GetActions()),
		CreatedBy:   md.UserID.String(),
	}
	created, err := s.store.CreateWorkflow(w)
	if err != nil {
		log.Printf("[automation] create workflow: %v", err)
		return nil, status.Error(codes.Internal, "failed to create workflow")
	}
	return WorkflowToPb(created), nil
}

func (s *Server) UpdateWorkflow(ctx context.Context, req *pb.UpdateWorkflowRequest) (*pb.Workflow, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	existing, err := s.store.GetWorkflow(req.GetId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "workflow not found")
	}
	if existing.OrgID != md.OrganisationID.String() {
		return nil, status.Error(codes.PermissionDenied, "workflow does not belong to caller's org")
	}

	updates := map[string]interface{}{}
	if req.GetName() != "" {
		updates["name"] = req.GetName()
	}
	if req.GetDescription() != "" {
		updates["description"] = req.GetDescription()
	}
	if req.GetStatus() != "" {
		updates["status"] = string(normaliseStatus(req.GetStatus()))
	}
	if req.GetTrigger() != nil {
		updates["trigger"] = triggerFromStruct(req.GetTrigger())
	}
	if req.GetConditions() != nil {
		updates["conditions"] = conditionsFromStructs(req.GetConditions())
	}
	if req.GetActions() != nil {
		updates["actions"] = actionsFromStructs(req.GetActions())
	}
	if len(updates) == 0 {
		return WorkflowToPb(existing), nil
	}
	updated, err := s.store.UpdateWorkflow(req.GetId(), updates)
	if err != nil {
		log.Printf("[automation] update workflow %s: %v", req.GetId(), err)
		return nil, status.Error(codes.Internal, "failed to update workflow")
	}
	return WorkflowToPb(updated), nil
}

func (s *Server) GetWorkflow(ctx context.Context, req *pb.GetWorkflowRequest) (*pb.Workflow, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	w, err := s.store.GetWorkflow(req.GetId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "workflow not found")
	}
	if w.OrgID != md.OrganisationID.String() {
		return nil, status.Error(codes.PermissionDenied, "workflow does not belong to caller's org")
	}
	return WorkflowToPb(w), nil
}

func (s *Server) ListWorkflows(ctx context.Context, req *pb.ListWorkflowsRequest) (*pb.ListWorkflowsResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	items, total, err := s.store.ListWorkflows(
		md.OrganisationID.String(),
		req.GetWorkspaceId(),
		int(req.GetPageNumber()),
		int(req.GetPageSize()),
	)
	if err != nil {
		log.Printf("[automation] list workflows: %v", err)
		return nil, status.Error(codes.Internal, "failed to list workflows")
	}
	out := make([]*pb.Workflow, len(items))
	for i, w := range items {
		out[i] = WorkflowToPb(w)
	}
	return &pb.ListWorkflowsResponse{Items: out, Total: total}, nil
}

func (s *Server) DeleteWorkflow(ctx context.Context, req *pb.DeleteWorkflowRequest) (*pb.DeleteWorkflowResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	existing, err := s.store.GetWorkflow(req.GetId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "workflow not found")
	}
	if existing.OrgID != md.OrganisationID.String() {
		return nil, status.Error(codes.PermissionDenied, "workflow does not belong to caller's org")
	}
	if err := s.store.DeleteWorkflow(req.GetId()); err != nil {
		log.Printf("[automation] delete workflow %s: %v", req.GetId(), err)
		return nil, status.Error(codes.Internal, "failed to delete workflow")
	}
	return &pb.DeleteWorkflowResponse{Success: true}, nil
}

// ── Run history ──────────────────────────────────────────────────────────────

func (s *Server) GetRunHistory(ctx context.Context, req *pb.GetRunHistoryRequest) (*pb.GetRunHistoryResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	wf, err := s.store.GetWorkflow(req.GetWorkflowId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "workflow not found")
	}
	if wf.OrgID != md.OrganisationID.String() {
		return nil, status.Error(codes.PermissionDenied, "workflow does not belong to caller's org")
	}
	runs, total, err := s.store.ListRuns(
		req.GetWorkflowId(),
		int(req.GetPageNumber()),
		int(req.GetPageSize()),
	)
	if err != nil {
		log.Printf("[automation] list runs: %v", err)
		return nil, status.Error(codes.Internal, "failed to list runs")
	}
	out := make([]*pb.WorkflowRun, len(runs))
	for i, r := range runs {
		out[i] = WorkflowRunToPb(r)
	}
	return &pb.GetRunHistoryResponse{Runs: out, Total: total}, nil
}

// TestRunWorkflow kicks off a single execution of the named workflow with a
// synthetic event payload — used by the visual builder's "Test run" button.
func (s *Server) TestRunWorkflow(ctx context.Context, req *pb.TestRunWorkflowRequest) (*pb.TestRunWorkflowResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if s.executor == nil || !s.executor.Enabled() {
		return nil, status.Error(codes.FailedPrecondition, "automation executor not configured (Temporal unavailable)")
	}
	wf, err := s.store.GetWorkflow(req.GetWorkflowId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "workflow not found")
	}
	if wf.OrgID != md.OrganisationID.String() {
		return nil, status.Error(codes.PermissionDenied, "workflow does not belong to caller's org")
	}

	event, err := events.NewEvent(
		wf.OrgID, wf.WorkspaceID,
		"test", "run",
		uuid.NewString(),
		md.UserID.String(),
		structToMap(req.GetSampleEvent()),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "build synthetic event: "+err.Error())
	}
	event.Subject = "kyla." + wf.OrgID + ".test.run"

	runID, err := s.executor.StartWorkflow(ctx, wf, event)
	if err != nil {
		log.Printf("[automation] test run %s: %v", wf.ID, err)
		return nil, status.Error(codes.Internal, "failed to start test run")
	}
	return &pb.TestRunWorkflowResponse{TemporalRunId: runID}, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func (s *Server) requireAuth(ctx context.Context) (*authctx.RequestMetadata, error) {
	md, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || md.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}
	return md, nil
}

func normaliseStatus(s string) WorkflowStatus {
	switch WorkflowStatus(s) {
	case WorkflowStatusActive, WorkflowStatusInactive, WorkflowStatusDraft:
		return WorkflowStatus(s)
	default:
		return WorkflowStatusDraft
	}
}
