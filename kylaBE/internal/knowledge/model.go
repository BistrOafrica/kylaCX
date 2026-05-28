package knowledge

import "time"

// KBCategory is a workspace-scoped knowledge base category.
type KBCategory struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID        string    `gorm:"type:uuid;not null;index"                       json:"org_id"`
	WorkspaceID  string    `gorm:"type:uuid;index"                                json:"workspace_id"`
	Name         string    `gorm:"not null"                                       json:"name"`
	Slug         string    `gorm:"not null"                                       json:"slug"`
	Icon         string    `json:"icon,omitempty"`
	ParentID     *string   `gorm:"type:uuid"                                      json:"parent_id,omitempty"`
	Position     int       `gorm:"default:0"                                      json:"position"`
	ArticleCount int       `gorm:"-"                                              json:"article_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (KBCategory) TableName() string { return "kb_categories" }

// KBArticle is a workspace-scoped knowledge base article.
type KBArticle struct {
	ID          string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID       string     `gorm:"type:uuid;not null;index"                       json:"org_id"`
	WorkspaceID string     `gorm:"type:uuid;index"                                json:"workspace_id"`
	CategoryID  string     `gorm:"type:uuid;index"                                json:"category_id"`
	Title       string     `gorm:"not null;index"                                 json:"title"`
	Slug        string     `gorm:"not null"                                       json:"slug"`
	Content     string     `gorm:"type:text"                                      json:"content"`
	Excerpt     string     `json:"excerpt,omitempty"`
	Status      string     `gorm:"not null;default:'draft'"                       json:"status"` // draft|published|archived
	AuthorID    string     `gorm:"type:uuid"                                      json:"author_id"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	ViewCount   int        `gorm:"default:0"                                      json:"view_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (KBArticle) TableName() string { return "kb_articles" }
