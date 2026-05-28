package activities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"kyla-be/shared/events"
)

// AssignUserActivity sets an assignee on an Object Core record by writing
// `assigned_to` into the object's data JSONB.
//
// Expected node.Config keys:
//
//	object_id  string — required (or falls back to event.EntityID)
//	user_id    string — required
//	field      string — optional; the JSONB key to write to (default "assigned_to")
//	actor_id   string — optional; defaults to event actor
//
// This activity is intentionally a thin wrapper over ObjectStore.UpdateObject
// rather than reaching into a separate "assignment" table — assignments live
// inside the object's data, consistent with how MoveDeal patches `stage_id`.
type AssignUserActivity struct {
	Deps Deps
}

func (a *AssignUserActivity) AssignUser(ctx context.Context, params map[string]interface{}, event events.DomainEvent) (string, error) {
	if a.Deps.ObjectStore == nil {
		return "", errors.New("assign_user: ObjectStore dependency missing")
	}
	objectID, _ := params["object_id"].(string)
	if objectID == "" {
		objectID = event.EntityID
	}
	userID, _ := params["user_id"].(string)
	if objectID == "" || userID == "" {
		return "", errors.New("assign_user: object_id and user_id are required")
	}
	field, _ := params["field"].(string)
	if field == "" {
		field = "assigned_to"
	}
	actorID, _ := params["actor_id"].(string)
	if actorID == "" {
		actorID = event.ActorID
	}

	patch, err := json.Marshal(map[string]interface{}{field: userID})
	if err != nil {
		return "", fmt.Errorf("assign_user: marshal patch: %w", err)
	}
	obj, err := a.Deps.ObjectStore.UpdateObject(objectID, event.OrgID, actorID, patch)
	if err != nil {
		return "", fmt.Errorf("assign_user: %w", err)
	}
	return obj.ID, nil
}
