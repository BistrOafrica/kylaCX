package knowledge

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

// AuthGateway is the subset of the auth stack KnowledgeServer needs.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
	ScopeCheck(ctx context.Context, scopeID string) (bool, *authctx.RequestMetadata, error)
}

// KnowledgeServer implements pb.KnowledgeServiceServer.
type KnowledgeServer struct {
	store    *KnowledgeStore
	auth     AuthGateway
	eventBus events.Publisher
	pb.UnimplementedKnowledgeServiceServer
}

// NewKnowledgeServer constructs a KnowledgeServer.
func NewKnowledgeServer(store *KnowledgeStore, auth AuthGateway, eventBus events.Publisher) *KnowledgeServer {
	return &KnowledgeServer{store: store, auth: auth, eventBus: eventBus}
}

// ── Category RPCs ─────────────────────────────────────────────────────────────

func (s *KnowledgeServer) CreateKBCategory(ctx context.Context, req *pb.CreateKBCategoryRequest) (*pb.KBCategory, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId(), req.GetWorkspaceId()); err != nil {
		return nil, err
	}
	cat := &KBCategory{
		OrgID:       req.GetOrgId(),
		WorkspaceID: req.GetWorkspaceId(),
		Name:        req.GetName(),
		Icon:        req.GetIcon(),
		Position:    int(req.GetPosition()),
	}
	if pid := req.GetParentId(); pid != "" {
		cat.ParentID = &pid
	}
	created, err := s.store.CreateCategory(cat)
	if err != nil {
		return nil, status.Error(500, "failed to create category")
	}
	s.publishEvent(req.GetOrgId(), req.GetWorkspaceId(), "category.created", created.ID, reqAuth.UserID.String(), map[string]string{"name": created.Name})
	return CategoryToPb(created), nil
}

func (s *KnowledgeServer) ListKBCategories(ctx context.Context, req *pb.ListKBCategoriesRequest) (*pb.ListKBCategoriesResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId(), req.GetWorkspaceId()); err != nil {
		return nil, err
	}
	cats, err := s.store.ListCategories(req.GetOrgId(), req.GetWorkspaceId())
	if err != nil {
		return nil, status.Error(500, "failed to list categories")
	}
	out := make([]*pb.KBCategory, len(cats))
	for i, c := range cats {
		out[i] = CategoryToPb(c)
	}
	return &pb.ListKBCategoriesResponse{Categories: out}, nil
}

func (s *KnowledgeServer) UpdateKBCategory(ctx context.Context, req *pb.UpdateKBCategoryRequest) (*pb.KBCategory, error) {
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
	if req.GetIcon() != "" {
		updates["icon"] = req.GetIcon()
	}
	updates["position"] = int(req.GetPosition())
	updated, err := s.store.UpdateCategory(req.GetId(), req.GetOrgId(), updates)
	if err != nil {
		return nil, status.Error(500, "failed to update category")
	}
	s.publishEvent(req.GetOrgId(), updated.WorkspaceID, "category.updated", updated.ID, reqAuth.UserID.String(), nil)
	return CategoryToPb(updated), nil
}

func (s *KnowledgeServer) DeleteKBCategory(ctx context.Context, req *pb.DeleteKBCategoryRequest) (*pb.DeleteKBCategoryResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	if err := s.store.DeleteCategory(req.GetId(), req.GetOrgId()); err != nil {
		return nil, status.Error(500, "failed to delete category")
	}
	s.publishEvent(req.GetOrgId(), "", "category.deleted", req.GetId(), reqAuth.UserID.String(), nil)
	return &pb.DeleteKBCategoryResponse{Success: true}, nil
}

// ── Article RPCs ──────────────────────────────────────────────────────────────

func (s *KnowledgeServer) CreateKBArticle(ctx context.Context, req *pb.CreateKBArticleRequest) (*pb.KBArticle, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId(), req.GetWorkspaceId()); err != nil {
		return nil, err
	}
	article := &KBArticle{
		OrgID:       req.GetOrgId(),
		WorkspaceID: req.GetWorkspaceId(),
		CategoryID:  req.GetCategoryId(),
		Title:       req.GetTitle(),
		Content:     req.GetContent(),
		Excerpt:     req.GetExcerpt(),
		Status:      "draft",
		AuthorID:    reqAuth.UserID.String(),
	}
	created, err := s.store.CreateArticle(article)
	if err != nil {
		return nil, status.Error(500, "failed to create article")
	}
	s.publishEvent(req.GetOrgId(), req.GetWorkspaceId(), "article.created", created.ID, reqAuth.UserID.String(), map[string]string{"title": created.Title})
	return ArticleToPb(created), nil
}

func (s *KnowledgeServer) GetKBArticle(ctx context.Context, req *pb.GetKBArticleRequest) (*pb.KBArticle, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	article, err := s.store.FindArticleByID(req.GetId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(404, "article not found")
	}
	return ArticleToPb(article), nil
}

func (s *KnowledgeServer) ListKBArticles(ctx context.Context, req *pb.ListKBArticlesRequest) (*pb.ListKBArticlesResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId(), req.GetWorkspaceId()); err != nil {
		return nil, err
	}
	articles, nextToken, total, err := s.store.ListArticles(ListArticlesParams{
		OrgID:       req.GetOrgId(),
		WorkspaceID: req.GetWorkspaceId(),
		CategoryID:  req.GetCategoryId(),
		Status:      articleStatusFromPb(req.GetStatus()),
		PageSize:    int(req.GetPageSize()),
		PageToken:   req.GetPageToken(),
	})
	if err != nil {
		return nil, status.Error(500, "failed to list articles")
	}
	out := make([]*pb.KBArticle, len(articles))
	for i, a := range articles {
		out[i] = ArticleToPb(a)
	}
	return &pb.ListKBArticlesResponse{
		Articles:      out,
		NextPageToken: nextToken,
		Total:         int32(total),
	}, nil
}

func (s *KnowledgeServer) UpdateKBArticle(ctx context.Context, req *pb.UpdateKBArticleRequest) (*pb.KBArticle, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if req.GetTitle() != "" {
		updates["title"] = req.GetTitle()
	}
	if req.GetContent() != "" {
		updates["content"] = req.GetContent()
	}
	if req.GetExcerpt() != "" {
		updates["excerpt"] = req.GetExcerpt()
	}
	if req.GetCategoryId() != "" {
		updates["category_id"] = req.GetCategoryId()
	}
	updated, err := s.store.UpdateArticle(req.GetId(), req.GetOrgId(), updates)
	if err != nil {
		return nil, status.Error(500, "failed to update article")
	}
	s.publishEvent(req.GetOrgId(), updated.WorkspaceID, "article.updated", updated.ID, reqAuth.UserID.String(), nil)
	return ArticleToPb(updated), nil
}

func (s *KnowledgeServer) DeleteKBArticle(ctx context.Context, req *pb.DeleteKBArticleRequest) (*pb.DeleteKBArticleResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	if err := s.store.DeleteArticle(req.GetId(), req.GetOrgId()); err != nil {
		return nil, status.Error(500, "failed to delete article")
	}
	s.publishEvent(req.GetOrgId(), "", "article.deleted", req.GetId(), reqAuth.UserID.String(), nil)
	return &pb.DeleteKBArticleResponse{Success: true}, nil
}

func (s *KnowledgeServer) PublishKBArticle(ctx context.Context, req *pb.PublishKBArticleRequest) (*pb.KBArticle, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	published, err := s.store.PublishArticle(req.GetId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(500, "failed to publish article")
	}
	s.publishEvent(req.GetOrgId(), published.WorkspaceID, "article.published", published.ID, reqAuth.UserID.String(), nil)
	return ArticleToPb(published), nil
}

func (s *KnowledgeServer) SearchKBArticles(ctx context.Context, req *pb.SearchKBArticlesRequest) (*pb.SearchKBArticlesResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId(), req.GetWorkspaceId()); err != nil {
		return nil, err
	}
	articles, err := s.store.SearchArticles(req.GetOrgId(), req.GetWorkspaceId(), req.GetQuery(), int(req.GetLimit()))
	if err != nil {
		return nil, status.Error(500, "failed to search articles")
	}
	out := make([]*pb.KBArticle, len(articles))
	for i, a := range articles {
		out[i] = ArticleToPb(a)
	}
	return &pb.SearchKBArticlesResponse{Articles: out}, nil
}

func (s *KnowledgeServer) authorizeScope(ctx context.Context, scopeIDs ...string) error {
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

func (s *KnowledgeServer) publishEvent(orgID, workspaceID, action, entityID, actorID string, payload interface{}) {
	ev, err := events.NewEvent(orgID, workspaceID, "knowledge", action, entityID, actorID, payload)
	if err != nil {
		return
	}
	ev.Subject = fmt.Sprintf("kyla.%s.knowledge.%s", orgID, action)
	if err := s.eventBus.Publish(ev); err != nil {
		log.Printf("[knowledge] event publish error (%s): %v", action, err)
	}
}
