package workspace

import "time"

// DomainTemplate identifies the preconfigured template for a new workspace.
type DomainTemplate string

const (
	DomainTemplateSupport    DomainTemplate = "support"
	DomainTemplateSales      DomainTemplate = "sales"
	DomainTemplateMarketing  DomainTemplate = "marketing"
	DomainTemplateOperations DomainTemplate = "operations"
	DomainTemplateCustom     DomainTemplate = "custom"
)

// WorkspaceStatus is the lifecycle state of a workspace.
type WorkspaceStatus string

const (
	WorkspaceStatusActive    WorkspaceStatus = "active"
	WorkspaceStatusArchived  WorkspaceStatus = "archived"
	WorkspaceStatusSuspended WorkspaceStatus = "suspended"
)

// Workspace is a logical product domain within an organisation.
type Workspace struct {
	ID             string                 `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID          string                 `gorm:"type:uuid;not null;index"                       json:"org_id"`
	Name           string                 `gorm:"not null"                                       json:"name"`
	Slug           string                 `gorm:"not null"                                       json:"slug"`
	Description    string                 `json:"description,omitempty"`
	Icon           string                 `json:"icon,omitempty"`
	Color          string                 `json:"color,omitempty"`
	DomainTemplate DomainTemplate         `gorm:"not null;default:'custom'"                      json:"domain_template"`
	Status         WorkspaceStatus        `gorm:"not null;default:'active'"                      json:"status"`
	Config         map[string]interface{} `gorm:"type:jsonb;serializer:json"                     json:"config,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// MemberRole defines what a workspace member can do.
type MemberRole string

const (
	MemberRoleOwner  MemberRole = "owner"
	MemberRoleAdmin  MemberRole = "admin"
	MemberRoleMember MemberRole = "member"
	MemberRoleGuest  MemberRole = "guest"
)

// WorkspaceMember links a user to a workspace with an assigned role.
type WorkspaceMember struct {
	WorkspaceID string     `gorm:"type:uuid;primaryKey" json:"workspace_id"`
	UserID      string     `gorm:"type:uuid;primaryKey" json:"user_id"`
	Role        MemberRole `gorm:"not null;default:'member'" json:"role"`
	JoinedAt    time.Time  `json:"joined_at"`
}
