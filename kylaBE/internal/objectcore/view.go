package objectcore

import (
	"encoding/json"
	"time"
)

// ViewFilter is a single filter condition stored inside SavedView.Filters.
type ViewFilter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // "eq", "neq", "contains", "gt", "lt", "in", "is_null"
	Value    interface{} `json:"value,omitempty"`
}

// ViewSort defines the sort order for a view.
type ViewSort struct {
	Field string `json:"field"`
	Desc  bool   `json:"desc"`
}

// ViewColumn defines a visible column in a view.
type ViewColumn struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Width int    `json:"width,omitempty"` // px, 0 = auto
}

// SavedView is a named, persisted filter+sort+column configuration for an
// object type.  Views belong to a workspace and can be shared with members.
type SavedView struct {
	ID          string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID string          `gorm:"type:uuid;not null;index"                       json:"workspace_id"`
	OrgID       string          `gorm:"type:uuid;not null;index"                       json:"org_id"`
	Name        string          `gorm:"not null"                                       json:"name"`
	TypeSlug    string          `gorm:"not null;index"                                 json:"type_slug"`
	Filters     json.RawMessage `gorm:"type:jsonb;not null;default:'[]'"               json:"filters"`
	Sort        json.RawMessage `gorm:"type:jsonb;not null;default:'{}'"               json:"sort"`
	Columns     json.RawMessage `gorm:"type:jsonb;not null;default:'[]'"               json:"columns"`
	IsShared    bool            `gorm:"default:false"                                  json:"is_shared"`
	CreatedBy   string          `gorm:"type:uuid"                                      json:"created_by,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
