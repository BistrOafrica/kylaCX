package automation

import (
	"encoding/json"
	"time"
)

// WorkflowStatus indicates whether a workflow is active.
type WorkflowStatus string

const (
	WorkflowStatusActive   WorkflowStatus = "active"
	WorkflowStatusInactive WorkflowStatus = "inactive"
	WorkflowStatusDraft    WorkflowStatus = "draft"
)

// RunStatus represents the execution state of a workflow run.
type RunStatus string

const (
	RunStatusPending RunStatus = "pending"
	RunStatusRunning RunStatus = "running"
	RunStatusSuccess RunStatus = "success"
	RunStatusFailed  RunStatus = "failed"
	RunStatusSkipped RunStatus = "skipped"
)

// TriggerConfig defines what fires a workflow.
type TriggerConfig struct {
	Type                string `json:"type"`
	EventSubjectPattern string `json:"event_subject_pattern,omitempty"`
	ObjectType          string `json:"object_type,omitempty"`
	CronExpression      string `json:"cron_expression,omitempty"`
}

// ConditionGroup joins conditions with a logical operator.
type ConditionGroup struct {
	Operator   string      `json:"operator"`
	Conditions []Condition `json:"conditions"`
}

// Condition evaluates a field value in the event context.
type Condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// ActionNode is a single step in a workflow execution.
type ActionNode struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Config       map[string]interface{} `json:"config"`
	OnSuccess    string                 `json:"on_success,omitempty"`
	OnFailure    string                 `json:"on_failure,omitempty"`
	DelaySeconds int                    `json:"delay_seconds,omitempty"`
}

// Workflow is a user-defined automation rule.
type Workflow struct {
	ID          string           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID       string           `gorm:"type:uuid;not null;index"                       json:"org_id"`
	WorkspaceID string           `gorm:"type:uuid;index"                                json:"workspace_id"`
	Name        string           `gorm:"not null"                                       json:"name"`
	Description string           `gorm:"type:text;default:''"                           json:"description,omitempty"`
	Status      WorkflowStatus   `gorm:"not null;default:'draft'"                       json:"status"`
	Trigger     TriggerConfig    `gorm:"type:jsonb;serializer:json"                     json:"trigger"`
	Conditions  []ConditionGroup `gorm:"type:jsonb;serializer:json"                     json:"conditions,omitempty"`
	Actions     []ActionNode     `gorm:"type:jsonb;serializer:json"                     json:"actions"`
	RunCount    int              `gorm:"default:0"                                      json:"run_count"`
	CreatedBy   string           `gorm:"type:uuid"                                      json:"created_by,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

func (Workflow) TableName() string { return "workflows" }

// WorkflowRun records a single execution of a Workflow.
// Temporal owns the canonical execution state; this row is a projection.
type WorkflowRun struct {
	ID             string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID     string          `gorm:"type:uuid;not null;index"                       json:"workflow_id"`
	TemporalRunID  string          `gorm:"not null;index"                                 json:"temporal_run_id"`
	OrgID          string          `gorm:"type:uuid;not null;index"                       json:"org_id"`
	TriggerEventID string          `gorm:"index"                                          json:"trigger_event_id,omitempty"`
	Status         RunStatus       `gorm:"not null;default:'pending'"                     json:"status"`
	Context        json.RawMessage `gorm:"type:jsonb"                                     json:"context,omitempty"`
	Error          string          `json:"error,omitempty"`
	StartedAt      *time.Time      `json:"started_at,omitempty"`
	FinishedAt     *time.Time      `json:"finished_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

func (WorkflowRun) TableName() string { return "workflow_runs" }
