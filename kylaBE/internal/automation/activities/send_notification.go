package activities

import (
	"context"
	"errors"
	"fmt"
	"log"

	"kyla-be/shared/events"
)

// SendNotificationActivity delivers an in-app / push / email notification.
// The actual delivery channel is owned by the Notifier implementation passed
// in via Deps. When no Notifier is wired (notification service not yet built),
// the activity logs and returns success so workflows aren't blocked.
//
// Expected node.Config keys:
//
//	user_id  string — required; recipient
//	kind     string — optional; "in_app" | "push" | "email" (default "in_app")
//	title    string — required
//	body     string — required
//	payload  map    — optional; arbitrary structured payload for deep-link / templating
type SendNotificationActivity struct {
	Deps Deps
}

func (a *SendNotificationActivity) SendNotification(ctx context.Context, params map[string]interface{}, event events.DomainEvent) (string, error) {
	userID, _ := params["user_id"].(string)
	title, _ := params["title"].(string)
	body, _ := params["body"].(string)
	if userID == "" || title == "" || body == "" {
		return "", errors.New("send_notification: user_id, title, and body are required")
	}
	kind, _ := params["kind"].(string)
	if kind == "" {
		kind = "in_app"
	}
	payload, _ := params["payload"].(map[string]interface{})

	if a.Deps.Notifier == nil {
		log.Printf("[automation] send_notification (no notifier wired) org=%s user=%s kind=%s title=%q",
			event.OrgID, userID, kind, title)
		return "logged", nil
	}
	if err := a.Deps.Notifier.Send(ctx, event.OrgID, userID, kind, title, body, payload); err != nil {
		return "", fmt.Errorf("send_notification: %w", err)
	}
	return "sent", nil
}
