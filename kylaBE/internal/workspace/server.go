package workspace

import (
	"context"
	"fmt"
	"log"
	"time"

	"kyla-be/internal/authctx"
	casbinsvc "kyla-be/internal/casbin"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/shared/events"

	"google.golang.org/grpc/status"
)

// AuthGateway defines the auth methods WorkspaceServer needs.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
	ScopeCheck(ctx context.Context, scopeID string) (bool, *authctx.RequestMetadata, error)
}

// WorkspaceServer is the gRPC implementation of WorkspaceServiceServer.
type WorkspaceServer struct {
	store    *WorkspaceStore
	auth     AuthGateway
	eventBus events.Publisher
	enforcer *casbinsvc.Enforcer
	ocSeeder SystemTypeSeedable
	pb.UnimplementedWorkspaceServiceServer
}

// NewWorkspaceServer constructs a new WorkspaceServer.
func NewWorkspaceServer(store *WorkspaceStore, auth AuthGateway, eventBus events.Publisher, enforcer *casbinsvc.Enforcer, ocSeeder SystemTypeSeedable) *WorkspaceServer {
	return &WorkspaceServer{store: store, auth: auth, eventBus: eventBus, enforcer: enforcer, ocSeeder: ocSeeder}
}

// CreateWorkspace creates a new workspace and seeds its domain defaults.
func (s *WorkspaceServer) CreateWorkspace(ctx context.Context, req *pb.CreateWorkspaceRequest) (*pb.CreateWorkspaceResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: you don't have permission to create a workspace")
	}

	orgID := req.GetOrgId()
	if orgID == "" {
		return nil, status.Error(400, "org_id is required")
	}
	if req.GetWorkspace() == nil {
		return nil, status.Error(400, "workspace is required")
	}

	w := PbToWorkspace(req.GetWorkspace())
	w.OrgID = orgID
	if w.Status == "" {
		w.Status = WorkspaceStatusActive
	}
	if w.DomainTemplate == "" {
		w.DomainTemplate = DomainTemplateCustom
	}

	created, err := s.store.Create(w)
	if err != nil {
		return nil, status.Error(500, "failed to create workspace")
	}

	// Seed domain defaults: creates system object types via the injected ocSeeder.
	if seedErr := SeedWorkspace(created, s.ocSeeder); seedErr != nil {
		log.Printf("[workspace] seed warning for %s: %v", created.ID, seedErr)
	}

	// Publish workspace.created event.
	s.publishEvent(created, "created", reqAuth.UserID.String())

	return &pb.CreateWorkspaceResponse{Workspace: WorkspaceToPb(created)}, nil
}

// GetWorkspace retrieves a workspace by ID.
func (s *WorkspaceServer) GetWorkspace(ctx context.Context, req *pb.GetWorkspaceRequest) (*pb.GetWorkspaceResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: you don't have permission to read a workspace")
	}

	w, err := s.store.FindByID(req.GetId())
	if err != nil {
		return nil, status.Error(404, "workspace not found")
	}
	return &pb.GetWorkspaceResponse{Workspace: WorkspaceToPb(w)}, nil
}

// ListWorkspaces returns all workspaces for an organisation.
func (s *WorkspaceServer) ListWorkspaces(ctx context.Context, req *pb.ListWorkspacesRequest) (*pb.ListWorkspacesResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: you don't have permission to list workspaces")
	}

	workspaces, err := s.store.FindByOrgID(req.GetOrgId())
	if err != nil {
		return nil, status.Error(500, "failed to list workspaces")
	}
	return &pb.ListWorkspacesResponse{Workspaces: WorkspacesToPb(workspaces)}, nil
}

// UpdateWorkspace updates an existing workspace.
func (s *WorkspaceServer) UpdateWorkspace(ctx context.Context, req *pb.UpdateWorkspaceRequest) (*pb.UpdateWorkspaceResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: you don't have permission to update a workspace")
	}

	w := PbToWorkspace(req.GetWorkspace())
	updated, err := s.store.Update(w)
	if err != nil {
		return nil, status.Error(500, "failed to update workspace")
	}

	s.publishEvent(updated, "updated", reqAuth.UserID.String())
	return &pb.UpdateWorkspaceResponse{Workspace: WorkspaceToPb(updated)}, nil
}

// ArchiveWorkspace sets a workspace to archived status.
func (s *WorkspaceServer) ArchiveWorkspace(ctx context.Context, req *pb.ArchiveWorkspaceRequest) (*pb.ArchiveWorkspaceResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: you don't have permission to archive a workspace")
	}

	if err := s.store.Archive(req.GetId()); err != nil {
		return nil, status.Error(500, "failed to archive workspace")
	}

	// Publish archived event with minimal payload.
	ev, evErr := events.NewEvent(
		reqAuth.OrganisationID.String(), "", "workspace", "archived", req.GetId(), reqAuth.UserID.String(), map[string]string{"id": req.GetId()},
	)
	if evErr == nil {
		_ = s.eventBus.Publish(ev)
	}

	return &pb.ArchiveWorkspaceResponse{
		Status: &pb.ArchiveWorkspaceResponse_Status{Code: 200, Message: "workspace archived"},
	}, nil
}

// AddMember adds a user as a member of a workspace.
func (s *WorkspaceServer) AddMember(ctx context.Context, req *pb.AddMemberRequest) (*pb.AddMemberResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: you don't have permission to add members")
	}

	m := &WorkspaceMember{
		WorkspaceID: req.GetWorkspaceId(),
		UserID:      req.GetUserId(),
		Role:        MemberRole(req.GetRole()),
		JoinedAt:    time.Now(),
	}
	created, err := s.store.AddMember(m)
	if err != nil {
		return nil, status.Error(500, "failed to add member")
	}

	// Seed Casbin workspace role for the new member.
	if s.enforcer != nil {
		if cErr := casbinsvc.SeedWorkspaceMember(s.enforcer, req.GetWorkspaceId(), req.GetUserId(), req.GetRole()); cErr != nil {
			log.Printf("[workspace] casbin seed member warning (ws=%s user=%s): %v", req.GetWorkspaceId(), req.GetUserId(), cErr)
		}
	}

	return &pb.AddMemberResponse{Member: MemberToPb(created)}, nil
}

// RemoveMember removes a user from a workspace.
func (s *WorkspaceServer) RemoveMember(ctx context.Context, req *pb.RemoveMemberRequest) (*pb.RemoveMemberResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: you don't have permission to remove members")
	}
	_ = reqAuth

	if err := s.store.RemoveMember(req.GetWorkspaceId(), req.GetUserId()); err != nil {
		return nil, status.Error(500, "failed to remove member")
	}

	// Revoke all Casbin workspace roles for the removed user.
	if s.enforcer != nil {
		if cErr := casbinsvc.RevokeWorkspaceMember(s.enforcer, req.GetWorkspaceId(), req.GetUserId()); cErr != nil {
			log.Printf("[workspace] casbin revoke member warning (ws=%s user=%s): %v", req.GetWorkspaceId(), req.GetUserId(), cErr)
		}
	}

	return &pb.RemoveMemberResponse{
		Status: &pb.RemoveMemberResponse_Status{Code: 200, Message: "member removed"},
	}, nil
}

// UpdateMemberRole changes the role of a workspace member.
func (s *WorkspaceServer) UpdateMemberRole(ctx context.Context, req *pb.UpdateMemberRoleRequest) (*pb.UpdateMemberRoleResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: you don't have permission to update member roles")
	}

	// Read current role before updating so we can revoke the old Casbin grant.
	_, oldRole, lookupErr := s.store.IsMember(req.GetWorkspaceId(), req.GetUserId())

	updated, err := s.store.UpdateMemberRole(req.GetWorkspaceId(), req.GetUserId(), req.GetRole())
	if err != nil {
		return nil, status.Error(500, "failed to update member role")
	}

	// Swap Casbin workspace role: revoke old, grant new.
	if s.enforcer != nil && lookupErr == nil {
		wsDomain := fmt.Sprintf("ws:%s", req.GetWorkspaceId())
		oldCasbinRole := casbinsvc.WorkspaceRoleToPolicy(string(oldRole))
		newCasbinRole := casbinsvc.WorkspaceRoleToPolicy(req.GetRole())
		if oldCasbinRole != newCasbinRole {
			if rErr := s.enforcer.RevokeRoleInDomain(req.GetUserId(), oldCasbinRole, wsDomain); rErr != nil {
				log.Printf("[workspace] casbin revoke old role warning (ws=%s user=%s role=%s): %v", req.GetWorkspaceId(), req.GetUserId(), oldCasbinRole, rErr)
			}
			if gErr := casbinsvc.SeedWorkspaceMember(s.enforcer, req.GetWorkspaceId(), req.GetUserId(), req.GetRole()); gErr != nil {
				log.Printf("[workspace] casbin grant new role warning (ws=%s user=%s role=%s): %v", req.GetWorkspaceId(), req.GetUserId(), req.GetRole(), gErr)
			}
		}
	}

	return &pb.UpdateMemberRoleResponse{Member: MemberToPb(updated)}, nil
}

// ListMembers returns all members of a workspace.
func (s *WorkspaceServer) ListMembers(ctx context.Context, req *pb.ListMembersRequest) (*pb.ListMembersResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: you don't have permission to list members")
	}
	_ = reqAuth

	members, err := s.store.ListMembers(req.GetWorkspaceId())
	if err != nil {
		return nil, status.Error(500, "failed to list members")
	}
	return &pb.ListMembersResponse{Members: MembersToPb(members)}, nil
}

// publishEvent is a helper that builds and emits a workspace domain event.
func (s *WorkspaceServer) publishEvent(w *Workspace, action, actorID string) {
	ev, err := events.NewEvent(
		w.OrgID, w.ID, "workspace", action, w.ID, actorID,
		map[string]string{"id": w.ID, "name": w.Name, "slug": w.Slug, "domain_template": fmt.Sprintf("%s", w.DomainTemplate)},
	)
	if err != nil {
		log.Printf("[workspace] event build error: %v", err)
		return
	}
	if err := s.eventBus.Publish(ev); err != nil {
		log.Printf("[workspace] event publish error: %v", err)
	}
}
