package projects

import "time"

// Project is the backend persistence model for project.proto service.
type Project struct {
	ID          string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID       string     `gorm:"type:uuid;not null;index" json:"org_id"`
	Title       string     `gorm:"not null" json:"title"`
	Status      string     `gorm:"not null;default:'active'" json:"status"`
	Description string     `gorm:"type:text" json:"description"`
	Visibility  string     `gorm:"not null;default:'private'" json:"visibility"`
	ArchivedAt  *time.Time `json:"archived_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Project) TableName() string { return "projects" }
