package forms

import (
	"context"
	"fmt"
	"log"

	"kyla-be/internal/authctx"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/shared/events"

	"google.golang.org/grpc/status"
)

// AuthGateway is the subset of the auth stack FormsServer needs.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
	ScopeCheck(ctx context.Context, scopeID string) (bool, *authctx.RequestMetadata, error)
}

// FormsServer implements pb.FormsServiceServer.
type FormsServer struct {
	store    *FormsStore
	auth     AuthGateway
	eventBus events.Publisher
	pb.UnimplementedFormsServiceServer
}

// NewFormsServer constructs a FormsServer.
func NewFormsServer(store *FormsStore, auth AuthGateway, eventBus events.Publisher) *FormsServer {
	return &FormsServer{store: store, auth: auth, eventBus: eventBus}
}

// ── Form RPCs ─────────────────────────────────────────────────────────────────

func (s *FormsServer) CreateForm(ctx context.Context, req *pb.CreateFormRequest) (*pb.FormDefinition, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId(), req.GetWorkspaceId()); err != nil {
		return nil, err
	}
	f := &Form{
		OrgID:          req.GetOrgId(),
		WorkspaceID:    req.GetWorkspaceId(),
		Name:           req.GetName(),
		Description:    req.GetDescription(),
		Fields:         req.GetFields(),
		Status:         "draft",
		SubmitRedirect: req.GetSubmitRedirect(),
		CreatedBy:      reqAuth.UserID.String(),
	}
	if f.Fields == "" {
		f.Fields = "[]"
	}
	created, err := s.store.CreateForm(f)
	if err != nil {
		return nil, status.Error(500, "failed to create form")
	}
	s.publishEvent(req.GetOrgId(), req.GetWorkspaceId(), "created", created.ID, reqAuth.UserID.String(), map[string]string{"name": created.Name})
	return FormToPb(created), nil
}

func (s *FormsServer) GetForm(ctx context.Context, req *pb.GetFormRequest) (*pb.FormDefinition, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	f, err := s.store.FindFormByID(req.GetId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(404, "form not found")
	}
	return FormToPb(f), nil
}

func (s *FormsServer) ListForms(ctx context.Context, req *pb.ListFormsRequest) (*pb.ListFormsResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId(), req.GetWorkspaceId()); err != nil {
		return nil, err
	}
	filterStatus := formStatusFromPb(req.GetStatus())
	if req.GetStatus() == pb.FormStatus_FORM_STATUS_UNSPECIFIED {
		filterStatus = ""
	}
	forms, err := s.store.ListForms(req.GetOrgId(), req.GetWorkspaceId(), filterStatus)
	if err != nil {
		return nil, status.Error(500, "failed to list forms")
	}
	out := make([]*pb.FormDefinition, len(forms))
	for i, f := range forms {
		out[i] = FormToPb(f)
	}
	return &pb.ListFormsResponse{Forms: out}, nil
}

func (s *FormsServer) UpdateForm(ctx context.Context, req *pb.UpdateFormRequest) (*pb.FormDefinition, error) {
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
	if req.GetFields() != "" {
		updates["fields"] = req.GetFields()
	}
	if req.GetSubmitRedirect() != "" {
		updates["submit_redirect"] = req.GetSubmitRedirect()
	}
	if req.GetStatus() != pb.FormStatus_FORM_STATUS_UNSPECIFIED {
		updates["status"] = formStatusFromPb(req.GetStatus())
	}
	updated, err := s.store.UpdateForm(req.GetId(), req.GetOrgId(), updates)
	if err != nil {
		return nil, status.Error(500, "failed to update form")
	}
	s.publishEvent(req.GetOrgId(), updated.WorkspaceID, "updated", updated.ID, reqAuth.UserID.String(), nil)
	return FormToPb(updated), nil
}

func (s *FormsServer) DeleteForm(ctx context.Context, req *pb.DeleteFormRequest) (*pb.DeleteFormResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	if err := s.store.DeleteForm(req.GetId(), req.GetOrgId()); err != nil {
		return nil, status.Error(500, "failed to delete form")
	}
	s.publishEvent(req.GetOrgId(), "", "deleted", req.GetId(), reqAuth.UserID.String(), nil)
	return &pb.DeleteFormResponse{Success: true}, nil
}

// ── Submission RPCs ───────────────────────────────────────────────────────────

func (s *FormsServer) SubmitForm(ctx context.Context, req *pb.SubmitFormRequest) (*pb.FormSubmission, error) {
	// Submissions are public — no auth check (form submitters may not be authenticated agents)
	if req.GetFormId() == "" || req.GetOrgId() == "" {
		return nil, status.Error(400, "form_id and org_id are required")
	}

	// Fetch the form to check it's active
	form, err := s.store.FindFormByID(req.GetFormId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(404, "form not found")
	}
	if form.Status != "active" {
		return nil, status.Error(400, "form is not accepting submissions")
	}

	sub := &FormSubmission{
		FormID: req.GetFormId(),
		OrgID:  req.GetOrgId(),
		Data:   req.GetData(),
	}
	if sub.Data == "" {
		sub.Data = "{}"
	}

	// Create an Object Core record for the submission (best-effort)
	objID := s.store.CreateObjectRecord(form.OrgID, form.WorkspaceID, sub.Data, "system")
	if objID != "" {
		sub.ObjectID = &objID
	}

	created, err := s.store.CreateSubmission(sub)
	if err != nil {
		return nil, status.Error(500, "failed to record submission")
	}
	s.publishEvent(req.GetOrgId(), form.WorkspaceID, "submitted", created.ID, "system", map[string]string{"form_id": req.GetFormId()})
	return SubmissionToPb(created), nil
}

func (s *FormsServer) ListSubmissions(ctx context.Context, req *pb.ListSubmissionsRequest) (*pb.ListSubmissionsResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	subs, nextToken, total, err := s.store.ListSubmissions(ListSubmissionsParams{
		OrgID:     req.GetOrgId(),
		FormID:    req.GetFormId(),
		PageSize:  int(req.GetPageSize()),
		PageToken: req.GetPageToken(),
	})
	if err != nil {
		return nil, status.Error(500, "failed to list submissions")
	}
	out := make([]*pb.FormSubmission, len(subs))
	for i, s := range subs {
		out[i] = SubmissionToPb(s)
	}
	return &pb.ListSubmissionsResponse{
		Submissions:   out,
		NextPageToken: nextToken,
		Total:         int32(total),
	}, nil
}

func (s *FormsServer) authorizeScope(ctx context.Context, scopeIDs ...string) error {
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

func (s *FormsServer) publishEvent(orgID, workspaceID, action, entityID, actorID string, payload interface{}) {
	ev, err := events.NewEvent(orgID, workspaceID, "form", action, entityID, actorID, payload)
	if err != nil {
		return
	}
	ev.Subject = fmt.Sprintf("kyla.%s.form.%s", orgID, action)
	if err := s.eventBus.Publish(ev); err != nil {
		log.Printf("[forms] event publish error (%s): %v", action, err)
	}
}
