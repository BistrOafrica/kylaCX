package forms

import "time"

// Form is a data-collection form definition.
type Form struct {
	ID              string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID           string    `gorm:"type:uuid;not null;index"                       json:"org_id"`
	WorkspaceID     string    `gorm:"type:uuid;index"                                json:"workspace_id"`
	Name            string    `gorm:"not null"                                       json:"name"`
	Description     string    `json:"description,omitempty"`
	Fields          string    `gorm:"type:jsonb;not null;default:'[]'"               json:"fields"` // JSON array of field definitions
	Status          string    `gorm:"not null;default:'draft'"                       json:"status"` // draft|active|closed
	SubmitRedirect  string    `json:"submit_redirect,omitempty"`
	SubmissionCount int       `gorm:"default:0"                                      json:"submission_count"`
	CreatedBy       string    `gorm:"type:uuid"                                      json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (Form) TableName() string { return "forms" }

// FormSubmission is a single response to a Form.
type FormSubmission struct {
	ID        string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FormID    string     `gorm:"type:uuid;not null;index"                       json:"form_id"`
	OrgID     string     `gorm:"type:uuid;not null;index"                       json:"org_id"`
	Data      string     `gorm:"type:jsonb;not null;default:'{}'"               json:"data"`  // JSON {field_key: value}
	ObjectID  *string    `gorm:"type:uuid"                                      json:"object_id,omitempty"` // Object Core record
	CreatedAt time.Time  `json:"created_at"`
}

func (FormSubmission) TableName() string { return "form_submissions" }
