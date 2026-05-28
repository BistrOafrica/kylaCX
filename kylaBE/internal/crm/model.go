package crm

import "time"

// Pipeline is a CRM deal pipeline definition.
type Pipeline struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID       string    `gorm:"type:uuid;not null;index"                       json:"org_id"`
	WorkspaceID string    `gorm:"type:uuid;index"                                json:"workspace_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	Description string    `json:"description,omitempty"`
	Type        string    `gorm:"not null;default:'sales'"                       json:"type"`
	Color       string    `json:"color,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName returns the custom table name.
func (Pipeline) TableName() string { return "crm_pipelines" }

// PipelineStage is a single stage within a CRM pipeline.
type PipelineStage struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PipelineID  string    `gorm:"type:uuid;not null;index"                       json:"pipeline_id"`
	OrgID       string    `gorm:"type:uuid;not null;index"                       json:"org_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	Color       string    `json:"color,omitempty"`
	Index       int       `gorm:"not null;default:0"                             json:"index"`
	Probability int       `gorm:"default:0"                                      json:"probability"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName returns the custom table name.
func (PipelineStage) TableName() string { return "crm_pipeline_stages" }
