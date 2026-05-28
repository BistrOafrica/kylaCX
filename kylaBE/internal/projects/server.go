package projects

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

// AuthGateway is the subset of auth stack required by ProjectServer.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
	ScopeCheck(ctx context.Context, scopeID string) (bool, *authctx.RequestMetadata, error)
}

// ProjectServer implements pb.ProjectServiceServer.
type ProjectServer struct {
	store    *Store
	auth     AuthGateway
	eventBus events.Publisher
	pb.UnimplementedProjectServiceServer
}

func NewProjectServer(store *Store, auth AuthGateway, eventBus events.Publisher) *ProjectServer {
	return &ProjectServer{store: store, auth: auth, eventBus: eventBus}
}

func (s *ProjectServer) CreateProject(ctx context.Context, req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	orgID := reqAuth.OrganisationID.String()
	if req.GetProject() == nil || req.GetProject().GetTitle() == "" {
		return nil, status.Error(400, "project.title is required")
	}
	created, err := s.store.Create(&Project{
		OrgID:       orgID,
		Title:       req.GetProject().GetTitle(),
		Status:      req.GetProject().GetStatus(),
		Description: req.GetProject().GetDescription(),
		Visibility:  req.GetProject().GetVisibility(),
	})
	if err != nil {
		return nil, status.Error(500, "failed to create project")
	}
	s.publishEvent(orgID, "created", created.ID, reqAuth.UserID.String(), map[string]string{"title": created.Title})
	return &pb.CreateProjectResponse{Project: toPB(created)}, nil
}

func (s *ProjectServer) ReadProject(ctx context.Context, req *pb.ReadProjectRequest) (*pb.ReadProjectResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	orgID := reqAuth.OrganisationID.String()
	p, err := s.store.Read(orgID, req.GetId())
	if err != nil {
		return nil, status.Error(404, "project not found")
	}
	return &pb.ReadProjectResponse{Project: toPB(p)}, nil
}

func (s *ProjectServer) UpdateProject(ctx context.Context, req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if req.GetProject() == nil || req.GetProject().GetId() == "" {
		return nil, status.Error(400, "project.id is required")
	}
	orgID := reqAuth.OrganisationID.String()
	updated, err := s.store.Update(orgID, &Project{
		ID:          req.GetProject().GetId(),
		Title:       req.GetProject().GetTitle(),
		Status:      req.GetProject().GetStatus(),
		Description: req.GetProject().GetDescription(),
		Visibility:  req.GetProject().GetVisibility(),
	})
	if err != nil {
		return nil, status.Error(500, "failed to update project")
	}
	s.publishEvent(orgID, "updated", updated.ID, reqAuth.UserID.String(), map[string]string{"status": updated.Status})
	return &pb.UpdateProjectResponse{Project: toPB(updated)}, nil
}

func (s *ProjectServer) DeleteProject(ctx context.Context, req *pb.DeleteProjectRequest) (*pb.DeleteProjectResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	orgID := reqAuth.OrganisationID.String()
	if err := s.store.Delete(orgID, req.GetId()); err != nil {
		return nil, status.Error(500, "failed to delete project")
	}
	s.publishEvent(orgID, "deleted", req.GetId(), reqAuth.UserID.String(), nil)
	return &pb.DeleteProjectResponse{Id: req.GetId(), Success: true}, nil
}

func (s *ProjectServer) ReadProjects(ctx context.Context, req *pb.ReadProjectsRequest) (*pb.ReadProjectsResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	orgID := reqAuth.OrganisationID.String()
	projects, err := s.store.List(orgID, req.GetPage(), req.GetPerPage())
	if err != nil {
		return nil, status.Error(500, "failed to read projects")
	}
	out := make([]*pb.Project, len(projects))
	for i, p := range projects {
		out[i] = toPB(p)
	}
	return &pb.ReadProjectsResponse{Projects: out}, nil
}

func (s *ProjectServer) ArchiveProject(ctx context.Context, req *pb.ArchiveProjectRequest) (*pb.ArchiveProjectResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	orgID := reqAuth.OrganisationID.String()
	if err := s.store.Archive(orgID, req.GetId()); err != nil {
		return nil, status.Error(500, "failed to archive project")
	}
	s.publishEvent(orgID, "archived", req.GetId(), reqAuth.UserID.String(), nil)
	return &pb.ArchiveProjectResponse{Archived: true}, nil
}

func toPB(p *Project) *pb.Project {
	if p == nil {
		return nil
	}
	return &pb.Project{
		Id:          p.ID,
		Title:       p.Title,
		Status:      p.Status,
		Description: p.Description,
		Visibility:  p.Visibility,
	}
}

func (s *ProjectServer) publishEvent(orgID, action, entityID, actorID string, payload interface{}) {
	ev, err := events.NewEvent(orgID, "", "project", action, entityID, actorID, payload)
	if err != nil {
		return
	}
	ev.Subject = fmt.Sprintf("kyla.%s.project.%s", orgID, action)
	if err := s.eventBus.Publish(ev); err != nil {
		log.Printf("[project] event publish error (%s): %v", action, err)
	}
}
