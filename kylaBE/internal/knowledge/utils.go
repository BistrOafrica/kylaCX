package knowledge

import (
	"kyla-be/pkg/pb"
)

// CategoryToPb converts a KBCategory model to its proto representation.
func CategoryToPb(c *KBCategory) *pb.KBCategory {
	cat := &pb.KBCategory{
		Id:           c.ID,
		OrgId:        c.OrgID,
		WorkspaceId:  c.WorkspaceID,
		Name:         c.Name,
		Slug:         c.Slug,
		Icon:         c.Icon,
		Position:     int32(c.Position),
		ArticleCount: int32(c.ArticleCount),
		CreatedAt:    c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if c.ParentID != nil {
		cat.ParentId = c.ParentID
	}
	return cat
}

// ArticleToPb converts a KBArticle model to its proto representation.
func ArticleToPb(a *KBArticle) *pb.KBArticle {
	art := &pb.KBArticle{
		Id:          a.ID,
		OrgId:       a.OrgID,
		WorkspaceId: a.WorkspaceID,
		CategoryId:  a.CategoryID,
		Title:       a.Title,
		Slug:        a.Slug,
		Content:     a.Content,
		Excerpt:     a.Excerpt,
		Status:      articleStatusStringToPb(a.Status),
		AuthorId:    a.AuthorID,
		ViewCount:   int32(a.ViewCount),
		CreatedAt:   a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   a.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if a.PublishedAt != nil {
		ts := a.PublishedAt.Format("2006-01-02T15:04:05Z07:00")
		art.PublishedAt = &ts
	}
	return art
}

func articleStatusStringToPb(s string) pb.ArticleStatus {
	switch s {
	case "published":
		return pb.ArticleStatus_ARTICLE_STATUS_PUBLISHED
	case "archived":
		return pb.ArticleStatus_ARTICLE_STATUS_ARCHIVED
	default:
		return pb.ArticleStatus_ARTICLE_STATUS_DRAFT
	}
}

func articleStatusFromPb(s pb.ArticleStatus) string {
	switch s {
	case pb.ArticleStatus_ARTICLE_STATUS_PUBLISHED:
		return "published"
	case pb.ArticleStatus_ARTICLE_STATUS_ARCHIVED:
		return "archived"
	default:
		return "draft"
	}
}
