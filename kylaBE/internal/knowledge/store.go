package knowledge

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// KnowledgeStore is the database layer for KB categories and articles.
type KnowledgeStore struct {
	db *gorm.DB
}

// NewKnowledgeStore constructs a KnowledgeStore.
func NewKnowledgeStore(db *gorm.DB) *KnowledgeStore {
	return &KnowledgeStore{db: db}
}

// ── Categories ────────────────────────────────────────────────────────────────

func (s *KnowledgeStore) CreateCategory(c *KBCategory) (*KBCategory, error) {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	if c.Slug == "" {
		c.Slug = slugify(c.Name)
	}
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
	if err := s.db.Create(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

func (s *KnowledgeStore) ListCategories(orgID, workspaceID string) ([]*KBCategory, error) {
	q := s.db.Where("org_id = ?", orgID)
	if workspaceID != "" {
		q = q.Where("workspace_id = ?", workspaceID)
	}
	var cats []*KBCategory
	if err := q.Order("position ASC, name ASC").Find(&cats).Error; err != nil {
		return nil, err
	}
	// Populate article counts
	for _, c := range cats {
		var count int64
		s.db.Model(&KBArticle{}).Where("category_id = ? AND org_id = ?", c.ID, orgID).Count(&count)
		c.ArticleCount = int(count)
	}
	return cats, nil
}

func (s *KnowledgeStore) FindCategoryByID(id, orgID string) (*KBCategory, error) {
	var c KBCategory
	if err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *KnowledgeStore) UpdateCategory(id, orgID string, updates map[string]interface{}) (*KBCategory, error) {
	updates["updated_at"] = time.Now()
	if err := s.db.Model(&KBCategory{}).Where("id = ? AND org_id = ?", id, orgID).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.FindCategoryByID(id, orgID)
}

func (s *KnowledgeStore) DeleteCategory(id, orgID string) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&KBCategory{}).Error
}

// ── Articles ──────────────────────────────────────────────────────────────────

func (s *KnowledgeStore) CreateArticle(a *KBArticle) (*KBArticle, error) {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.Slug == "" {
		a.Slug = slugify(a.Title)
	}
	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now
	if err := s.db.Create(a).Error; err != nil {
		return nil, err
	}
	return a, nil
}

func (s *KnowledgeStore) FindArticleByID(id, orgID string) (*KBArticle, error) {
	var a KBArticle
	if err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// ListArticlesParams bundles options for ListArticles.
type ListArticlesParams struct {
	OrgID       string
	WorkspaceID string
	CategoryID  string
	Status      string
	PageSize    int
	PageToken   string
}

func (s *KnowledgeStore) ListArticles(p ListArticlesParams) ([]*KBArticle, string, int64, error) {
	if p.PageSize <= 0 || p.PageSize > 200 {
		p.PageSize = 50
	}
	q := s.db.Model(&KBArticle{}).Where("org_id = ?", p.OrgID)
	if p.WorkspaceID != "" {
		q = q.Where("workspace_id = ?", p.WorkspaceID)
	}
	if p.CategoryID != "" {
		q = q.Where("category_id = ?", p.CategoryID)
	}
	if p.Status != "" {
		q = q.Where("status = ?", p.Status)
	}

	var total int64
	q.Count(&total)

	if p.PageToken != "" {
		if ts, id, ok := decodeCursor(p.PageToken); ok {
			q = q.Where("(created_at < ?) OR (created_at = ? AND id < ?)", ts, ts, id)
		} else {
			var pivot KBArticle
			if err := s.db.Select("id", "created_at").Where("id = ? AND org_id = ?", p.PageToken, p.OrgID).First(&pivot).Error; err == nil {
				q = q.Where("(created_at < ?) OR (created_at = ? AND id < ?)", pivot.CreatedAt, pivot.CreatedAt, pivot.ID)
			}
		}
	}
	q = q.Order("created_at DESC, id DESC").Limit(p.PageSize + 1)

	var articles []*KBArticle
	if err := q.Find(&articles).Error; err != nil {
		return nil, "", total, err
	}
	nextToken := ""
	if len(articles) > p.PageSize {
		nextToken = encodeCursor(articles[p.PageSize-1].CreatedAt, articles[p.PageSize-1].ID)
		articles = articles[:p.PageSize]
	}
	return articles, nextToken, total, nil
}

func (s *KnowledgeStore) UpdateArticle(id, orgID string, updates map[string]interface{}) (*KBArticle, error) {
	updates["updated_at"] = time.Now()
	if err := s.db.Model(&KBArticle{}).Where("id = ? AND org_id = ?", id, orgID).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.FindArticleByID(id, orgID)
}

func (s *KnowledgeStore) PublishArticle(id, orgID string) (*KBArticle, error) {
	now := time.Now()
	if err := s.db.Model(&KBArticle{}).Where("id = ? AND org_id = ?", id, orgID).
		Updates(map[string]interface{}{
			"status":       "published",
			"published_at": &now,
			"updated_at":   now,
		}).Error; err != nil {
		return nil, err
	}
	return s.FindArticleByID(id, orgID)
}

func (s *KnowledgeStore) DeleteArticle(id, orgID string) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&KBArticle{}).Error
}

func (s *KnowledgeStore) SearchArticles(orgID, workspaceID, query string, limit int) ([]*KBArticle, error) {
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	q := s.db.Where("org_id = ? AND status = 'published'", orgID)
	if workspaceID != "" {
		q = q.Where("workspace_id = ?", workspaceID)
	}
	q = q.Where("to_tsvector('english', title || ' ' || content) @@ websearch_to_tsquery('english', ?)", query).
		Limit(limit)
	var articles []*KBArticle
	if err := q.Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, s)
	return fmt.Sprintf("%s-%s", s, uuid.New().String()[:8])
}

func encodeCursor(ts time.Time, id string) string {
	return ts.UTC().Format(time.RFC3339Nano) + "|" + id
}

func decodeCursor(token string) (time.Time, string, bool) {
	parts := strings.SplitN(token, "|", 2)
	if len(parts) != 2 {
		return time.Time{}, "", false
	}
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, "", false
	}
	if parts[1] == "" {
		return time.Time{}, "", false
	}
	return ts, parts[1], true
}
