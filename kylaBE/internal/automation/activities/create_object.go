package activities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"kyla-be/internal/objectcore"
	"kyla-be/shared/events"
)

// CreateObjectActivity inserts a new Object Core record.
//
// Expected node.Config keys:
//
//	type_slug    string — required; e.g. "task", "ticket", "deal"
//	workspace_id string — optional; defaults to event.WorkspaceID
//	data         map    — required; serialised into Object.data JSONB
//	actor_id     string — optional; defaults to event actor
//
// Returns the new object's ID so downstream actions can reference it.
type CreateObjectActivity struct {
	Deps Deps
}

func (a *CreateObjectActivity) CreateObject(ctx context.Context, params map[string]interface{}, event events.DomainEvent) (string, error) {
	if a.Deps.ObjectStore == nil {
		return "", errors.New("create_object: ObjectStore dependency missing")
	}
	typeSlug, _ := params["type_slug"].(string)
	if typeSlug == "" {
		return "", errors.New("create_object: type_slug is required")
	}
	workspaceID, _ := params["workspace_id"].(string)
	if workspaceID == "" {
		workspaceID = event.WorkspaceID
	}
	dataRaw, ok := params["data"]
	if !ok {
		return "", errors.New("create_object: data is required")
	}
	dataBytes, err := json.Marshal(dataRaw)
	if err != nil {
		return "", fmt.Errorf("create_object: marshal data: %w", err)
	}
	actorID, _ := params["actor_id"].(string)
	if actorID == "" {
		actorID = event.ActorID
	}

	obj := &objectcore.Object{
		OrgID:       event.OrgID,
		WorkspaceID: workspaceID,
		TypeSlug:    typeSlug,
		Data:        dataBytes,
	}
	created, err := a.Deps.ObjectStore.CreateObject(obj, actorID)
	if err != nil {
		return "", fmt.Errorf("create_object: %w", err)
	}
	return created.ID, nil
}

// CreateTaskActivity is a thin wrapper that always creates an Object Core
// record with type_slug="task". Tasks live in Object Core per the Phase 4
// design — there is no dedicated tasks table.
type CreateTaskActivity struct {
	Deps Deps
}

func (a *CreateTaskActivity) CreateTask(ctx context.Context, params map[string]interface{}, event events.DomainEvent) (string, error) {
	if params == nil {
		params = map[string]interface{}{}
	}
	params["type_slug"] = "task"
	delegate := &CreateObjectActivity{Deps: a.Deps}
	return delegate.CreateObject(ctx, params, event)
}
