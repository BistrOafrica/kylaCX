package objectcore

import (
	"context"

	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"

	"google.golang.org/grpc/status"
)

// ViewServer implements pb.ViewServiceServer for managing SavedViews.
type ViewServer struct {
	store *ViewStore
	auth  AuthGateway
	pb.UnimplementedViewServiceServer
}

// NewViewServer constructs a ViewServer.
func NewViewServer(store *ViewStore, auth AuthGateway) *ViewServer {
	return &ViewServer{store: store, auth: auth}
}

func (s *ViewServer) CreateSavedView(ctx context.Context, req *pb.CreateSavedViewRequest) (*pb.SavedView, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	v := req.GetView()
	if v == nil {
		return nil, status.Error(400, "view is required")
	}

	model := PbToView(v)
	if model.CreatedBy == "" {
		model.CreatedBy = reqAuth.UserID.String()
	}

	created, err := s.store.Create(model)
	if err != nil {
		return nil, status.Error(500, "failed to create saved view")
	}
	return ViewToPb(created), nil
}

func (s *ViewServer) GetSavedView(ctx context.Context, req *pb.GetSavedViewRequest) (*pb.SavedView, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	v, err := s.store.FindByID(req.GetId(), req.GetWorkspaceId())
	if err != nil {
		return nil, status.Error(404, "saved view not found")
	}
	return ViewToPb(v), nil
}

func (s *ViewServer) ListSavedViews(ctx context.Context, req *pb.ListSavedViewsRequest) (*pb.ListSavedViewsResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	views, err := s.store.ListByWorkspace(req.GetWorkspaceId(), req.GetTypeSlug())
	if err != nil {
		return nil, status.Error(500, "failed to list saved views")
	}
	return &pb.ListSavedViewsResponse{Views: ViewsToPb(views)}, nil
}

func (s *ViewServer) UpdateSavedView(ctx context.Context, req *pb.UpdateSavedViewRequest) (*pb.SavedView, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	v := req.GetView()
	if v == nil {
		return nil, status.Error(400, "view is required")
	}

	model := PbToView(v)
	updated, err := s.store.Update(model)
	if err != nil {
		return nil, status.Error(500, "failed to update saved view")
	}
	return ViewToPb(updated), nil
}

func (s *ViewServer) DeleteSavedView(ctx context.Context, req *pb.DeleteSavedViewRequest) (*pb.DeleteSavedViewResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	if err := s.store.Delete(req.GetId(), req.GetWorkspaceId()); err != nil {
		return nil, status.Error(500, "failed to delete saved view")
	}
	return &pb.DeleteSavedViewResponse{Success: true}, nil
}
