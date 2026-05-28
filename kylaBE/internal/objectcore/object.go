package objectcore

import (
	"encoding/json"
	"time"
)

// FieldType enumerates supported schema field types.
type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypeNumber   FieldType = "number"
	FieldTypeDate     FieldType = "date"
	FieldTypeDateTime FieldType = "datetime"
	FieldTypeSelect   FieldType = "select"
	FieldTypeMulti    FieldType = "multi_select"
	FieldTypeBoolean  FieldType = "boolean"
	FieldTypeUser     FieldType = "user"
	FieldTypeRelation FieldType = "relation"
	FieldTypeFile     FieldType = "file"
	FieldTypeEmail    FieldType = "email"
	FieldTypePhone    FieldType = "phone"
	FieldTypeURL      FieldType = "url"
	FieldTypeCurrency FieldType = "currency"
)

// SelectOption is a named choice for select fields.
type SelectOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Color string `json:"color,omitempty"`
}

// FieldDefinition describes a single field in an object schema.
type FieldDefinition struct {
	Key        string         `json:"key"`
	Label      string         `json:"label"`
	Type       FieldType      `json:"type"`
	Required   bool           `json:"required,omitempty"`
	Unique     bool           `json:"unique,omitempty"`
	Searchable bool           `json:"searchable,omitempty"`
	Default    interface{}    `json:"default,omitempty"`
	Options    []SelectOption `json:"options,omitempty"`
	RelatesTo  string         `json:"relates_to,omitempty"`
}

// ObjectSchema holds an object type field definitions.
type ObjectSchema struct {
	Fields []FieldDefinition `json:"fields"`
}

// ObjectType defines the shape of a category of objects.
type ObjectType struct {
	ID          string       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID       string       `gorm:"type:uuid;not null;index"                       json:"org_id"`
	WorkspaceID string       `gorm:"type:uuid;index"                                json:"workspace_id,omitempty"`
	Slug        string       `gorm:"not null"                                       json:"slug"`
	Name        string       `gorm:"not null"                                       json:"name"`
	PluralName  string       `json:"plural_name,omitempty"`
	Icon        string       `json:"icon,omitempty"`
	Color       string       `json:"color,omitempty"`
	IsSystem    bool         `gorm:"default:false"                                  json:"is_system"`
	Schema      ObjectSchema `gorm:"type:jsonb;serializer:json"                     json:"schema"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Object is a single record of any object type.
type Object struct {
	ID          string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID       string          `gorm:"type:uuid;not null;index"                       json:"org_id"`
	WorkspaceID string          `gorm:"type:uuid;index"                                json:"workspace_id"`
	TypeSlug    string          `gorm:"not null;index"                                 json:"type_slug"`
	Data        json.RawMessage `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedBy   string          `gorm:"type:uuid"                                      json:"created_by,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// ObjectRelation links two objects with a named relation.
type ObjectRelation struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID     string    `gorm:"type:uuid;not null;index"                       json:"org_id"`
	FromID    string    `gorm:"type:uuid;not null;index"                       json:"from_id"`
	ToID      string    `gorm:"type:uuid;not null;index"                       json:"to_id"`
	Relation  string    `gorm:"not null"                                       json:"relation"`
	CreatedAt time.Time `json:"created_at"`
}

// ObjectEvent is an immutable timeline entry for an object.
type ObjectEvent struct {
	ID        string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID     string          `gorm:"type:uuid;not null;index"                       json:"org_id"`
	ObjectID  string          `gorm:"type:uuid;not null;index"                       json:"object_id"`
	ActorID   string          `gorm:"type:uuid"                                      json:"actor_id,omitempty"`
	ActorType string          `gorm:"default:'user'"                                 json:"actor_type"`
	EventType string          `gorm:"not null;index"                                 json:"event_type"`
	Payload   json.RawMessage `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt time.Time       `gorm:"index"                                          json:"created_at"`
}
