package crm

import (
	"context"
	"encoding/json"
	"log"

	"kyla-be/internal/authctx"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/shared/events"

	"google.golang.org/grpc/status"
)

// AuthGateway is the subset of the auth stack CRMServer needs.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
	ScopeCheck(ctx context.Context, scopeID string) (bool, *authctx.RequestMetadata, error)
}

// CRMServer implements pb.CRMServiceServer.
type CRMServer struct {
	store    *CRMStore
	auth     AuthGateway
	eventBus events.Publisher
	pb.UnimplementedCRMServiceServer
}

// NewCRMServer constructs a CRMServer.
func NewCRMServer(store *CRMStore, auth AuthGateway, eventBus events.Publisher) *CRMServer {
	return &CRMServer{store: store, auth: auth, eventBus: eventBus}
}

// ── Pipeline RPCs ─────────────────────────────────────────────────────────────

func (s *CRMServer) CreatePipeline(ctx context.Context, req *pb.CreatePipelineRequest) (*pb.Pipeline, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId(), req.GetWorkspaceId()); err != nil {
		return nil, err
	}
	if req.GetOrgId() == "" || req.GetName() == "" {
		return nil, status.Error(400, "org_id and name are required")
	}

	p := &Pipeline{
		OrgID:       req.GetOrgId(),
		WorkspaceID: req.GetWorkspaceId(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Type:        typeFromPb(req.GetType()),
		Color:       req.GetColor(),
	}
	created, err := s.store.CreatePipeline(p)
	if err != nil {
		return nil, status.Error(500, "failed to create pipeline")
	}
	s.publishEvent(created.OrgID, created.WorkspaceID, "crm.pipeline.created", created.ID)
	return PipelineToPb(created, 0), nil
}

func (s *CRMServer) GetPipeline(ctx context.Context, req *pb.GetPipelineRequest) (*pb.Pipeline, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	p, err := s.store.FindPipelineByID(req.GetId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(404, "pipeline not found")
	}
	count, _ := s.store.PipelineStageCount(p.ID, p.OrgID)
	return PipelineToPb(p, count), nil
}

func (s *CRMServer) ListPipelines(ctx context.Context, req *pb.ListPipelinesRequest) (*pb.ListPipelinesResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId(), req.GetWorkspaceId()); err != nil {
		return nil, err
	}
	pipelines, err := s.store.ListPipelines(req.GetOrgId(), req.GetWorkspaceId())
	if err != nil {
		return nil, status.Error(500, "failed to list pipelines")
	}
	out := make([]*pb.Pipeline, len(pipelines))
	for i, p := range pipelines {
		count, _ := s.store.PipelineStageCount(p.ID, p.OrgID)
		out[i] = PipelineToPb(p, count)
	}
	return &pb.ListPipelinesResponse{Pipelines: out}, nil
}

func (s *CRMServer) UpdatePipeline(ctx context.Context, req *pb.UpdatePipelineRequest) (*pb.Pipeline, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if req.GetName() != "" {
		updates["name"] = req.GetName()
	}
	if req.GetDescription() != "" {
		updates["description"] = req.GetDescription()
	}
	if req.GetColor() != "" {
		updates["color"] = req.GetColor()
	}
	if len(updates) == 0 {
		return nil, status.Error(400, "no fields to update")
	}
	updated, err := s.store.UpdatePipeline(req.GetId(), req.GetOrgId(), updates)
	if err != nil {
		return nil, status.Error(500, "failed to update pipeline")
	}
	s.publishEvent(updated.OrgID, updated.WorkspaceID, "crm.pipeline.updated", updated.ID)
	count, _ := s.store.PipelineStageCount(updated.ID, updated.OrgID)
	return PipelineToPb(updated, count), nil
}

func (s *CRMServer) DeletePipeline(ctx context.Context, req *pb.DeletePipelineRequest) (*pb.DeletePipelineResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	if err := s.store.DeletePipeline(req.GetId(), req.GetOrgId()); err != nil {
		return nil, status.Error(500, "failed to delete pipeline")
	}
	s.publishEvent(req.GetOrgId(), "", "crm.pipeline.deleted", req.GetId())
	return &pb.DeletePipelineResponse{Success: true}, nil
}

// ── Stage RPCs ────────────────────────────────────────────────────────────────

func (s *CRMServer) CreateStage(ctx context.Context, req *pb.CreateStageRequest) (*pb.PipelineStage, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	st := &PipelineStage{
		PipelineID:  req.GetPipelineId(),
		OrgID:       req.GetOrgId(),
		Name:        req.GetName(),
		Color:       req.GetColor(),
		Probability: int(req.GetProbability()),
	}
	created, err := s.store.CreateStage(st)
	if err != nil {
		return nil, status.Error(500, "failed to create stage")
	}
	s.publishEvent(created.OrgID, "", "crm.stage.created", created.ID)
	return StageToPb(created), nil
}

func (s *CRMServer) ListStages(ctx context.Context, req *pb.ListStagesRequest) (*pb.ListStagesResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	stages, err := s.store.ListStages(req.GetPipelineId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(500, "failed to list stages")
	}
	return &pb.ListStagesResponse{Stages: StagesToPb(stages)}, nil
}

func (s *CRMServer) UpdateStage(ctx context.Context, req *pb.UpdateStageRequest) (*pb.PipelineStage, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if req.GetName() != "" {
		updates["name"] = req.GetName()
	}
	if req.GetColor() != "" {
		updates["color"] = req.GetColor()
	}
	updates["probability"] = int(req.GetProbability())
	updated, err := s.store.UpdateStage(req.GetId(), req.GetOrgId(), updates)
	if err != nil {
		return nil, status.Error(500, "failed to update stage")
	}
	s.publishEvent(updated.OrgID, "", "crm.stage.updated", updated.ID)
	return StageToPb(updated), nil
}

func (s *CRMServer) DeleteStage(ctx context.Context, req *pb.DeleteStageRequest) (*pb.DeleteStageResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	if err := s.store.DeleteStage(req.GetId(), req.GetOrgId()); err != nil {
		return nil, status.Error(500, "failed to delete stage")
	}
	s.publishEvent(req.GetOrgId(), "", "crm.stage.deleted", req.GetId())
	return &pb.DeleteStageResponse{Success: true}, nil
}

func (s *CRMServer) ReorderStages(ctx context.Context, req *pb.ReorderStagesRequest) (*pb.ListStagesResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	if err := s.store.ReorderStageIndices(req.GetPipelineId(), req.GetOrgId(), req.GetStageIds()); err != nil {
		return nil, status.Error(500, "failed to reorder stages")
	}
	s.publishEvent(req.GetOrgId(), "", "crm.stage.reordered", req.GetPipelineId())
	stages, _ := s.store.ListStages(req.GetPipelineId(), req.GetOrgId())
	return &pb.ListStagesResponse{Stages: StagesToPb(stages)}, nil
}

// ── Deal operations RPCs ──────────────────────────────────────────────────────

func (s *CRMServer) MoveDeal(ctx context.Context, req *pb.MoveDealRequest) (*pb.MoveDealResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	if err := s.store.PatchDealStage(req.GetDealObjectId(), req.GetOrgId(), req.GetStageId()); err != nil {
		return nil, status.Error(500, "failed to move deal")
	}
	payload, _ := json.Marshal(map[string]string{
		"deal_id":  req.GetDealObjectId(),
		"stage_id": req.GetStageId(),
	})
	ev, _ := events.NewEvent(req.GetOrgId(), "", "deal", "stage_changed",
		req.GetDealObjectId(), reqAuth.UserID.String(), payload)
	if ev != nil {
		ev.Subject = "kyla." + req.GetOrgId() + ".deal.stage_changed"
		_ = s.eventBus.Publish(ev)
	}
	return &pb.MoveDealResponse{
		Success:  true,
		ObjectId: req.GetDealObjectId(),
		StageId:  req.GetStageId(),
	}, nil
}

func (s *CRMServer) GetPipelineBoard(ctx context.Context, req *pb.GetPipelineBoardRequest) (*pb.GetPipelineBoardResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId(), req.GetWorkspaceId()); err != nil {
		return nil, err
	}

	pipeline, err := s.store.FindPipelineByID(req.GetPipelineId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(404, "pipeline not found")
	}
	stages, err := s.store.ListStages(req.GetPipelineId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(500, "failed to list stages")
	}

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 20
	}

	columns := make([]*pb.PipelineBoardColumn, len(stages))
	for i, st := range stages {
		deals, err := s.store.ListDealsByStage(req.GetOrgId(), req.GetWorkspaceId(), st.ID, pageSize)
		if err != nil {
			log.Printf("[crm] board load stage %s: %v", st.ID, err)
		}
		cards := make([]*pb.DealCard, len(deals))
		for j, d := range deals {
			cards[j] = DealCardFromRow(d)
		}
		columns[i] = &pb.PipelineBoardColumn{
			Stage: StageToPb(st),
			Total: int32(len(deals)),
			Deals: cards,
		}
	}

	count, _ := s.store.PipelineStageCount(pipeline.ID, pipeline.OrgID)
	return &pb.GetPipelineBoardResponse{
		Pipeline: PipelineToPb(pipeline, count),
		Columns:  columns,
	}, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (s *CRMServer) publishEvent(orgID, workspaceID, subject, entityID string) {
	ev, err := events.NewEvent(orgID, workspaceID, "crm", subject, entityID, "", nil)
	if err != nil {
		return
	}
	ev.Subject = "kyla." + orgID + "." + subject
	if err := s.eventBus.Publish(ev); err != nil {
		log.Printf("[crm] event publish error (%s): %v", subject, err)
	}
}

func (s *CRMServer) authorizeScope(ctx context.Context, scopeIDs ...string) error {
	for _, scopeID := range scopeIDs {
		if scopeID == "" {
			continue
		}
		ok, _, err := s.auth.ScopeCheck(ctx, scopeID)
		if err != nil || !ok {
			return status.Error(403, "Forbidden")
		}
	}
	return nil
}
