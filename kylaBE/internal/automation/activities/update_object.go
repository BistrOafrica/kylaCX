package activities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"kyla-be/shared/events"
)

// UpdateObjectActivity patches arbitrary fields on an Object Core record.
//
// Expected node.Config keys:
//
//	object_id  string — required (may also come from the trigger event)
//	patch      map    — required, JSON-mergeable into Object.data
//	actor_id   string — optional; defaults to the trigger event's actor
//
// If object_id is omitted, the activity falls back to event.EntityID — this is
// the common case when a workflow reacts to "object.updated" and wants to
// mutate the object that just changed.
type UpdateObjectActivity struct {
	Deps Deps
}

// UpdateObject is the registered Temporal activity. Returning the patched
// object's ID lets downstream activities reference it via workflow state.
func (a *UpdateObjectActivity) UpdateObject(ctx context.Context, params map[string]interface{}, event events.DomainEvent) (string, error) {
	if a.Deps.ObjectStore == nil {
		return "", errors.New("update_object: ObjectStore dependency missing")
	}
	objectID, _ := params["object_id"].(string)
	if objectID == "" {
		objectID = event.EntityID
	}
	if objectID == "" {
		return "", errors.New("update_object: object_id is required (no fallback in event)")
	}
	patchRaw, ok := params["patch"]
	if !ok {
		return "", errors.New("update_object: patch is required")
	}
	patchBytes, err := json.Marshal(patchRaw)
	if err != nil {
		return "", fmt.Errorf("update_object: marshal patch: %w", err)
	}
	actorID, _ := params["actor_id"].(string)
	if actorID == "" {
		actorID = event.ActorID
	}

	obj, err := a.Deps.ObjectStore.UpdateObject(objectID, event.OrgID, actorID, patchBytes)
	if err != nil {
		return "", fmt.Errorf("update_object: %w", err)
	}
	return obj.ID, nil
}
