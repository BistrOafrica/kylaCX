package campaigns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"kyla-be/internal/communication"
	"kyla-be/internal/objectcore"
)

// ActivityDeps bundles the dependencies activities need. Mirrors the pattern
// used by internal/automation/activities/deps.go.
type ActivityDeps struct {
	Store             *Store
	ObjectStore       *objectcore.ObjectCoreStore
	ConversationStore *communication.ConversationStore
	MessageStore      *communication.MessageStore
	AdapterRegistry   *communication.AdapterRegistry
}

// ResolveAudienceActivity walks the campaign's audience definition and
// materialises one campaign_recipients row per matched record.
//
// The activity is idempotent — re-running it on the same campaign just
// re-inserts under the unique (campaign_id, object_id) constraint with
// ON CONFLICT DO NOTHING. That matches Temporal's at-least-once semantics.
type ResolveAudienceActivity struct {
	Deps ActivityDeps
}

func (a *ResolveAudienceActivity) ResolveAudience(ctx context.Context, campaignID string) (int, error) {
	if a.Deps.Store == nil || a.Deps.ObjectStore == nil {
		return 0, errors.New("resolve_audience: dependencies missing")
	}
	c, err := a.Deps.Store.lookupByIDOnly(campaignID)
	if err != nil {
		return 0, fmt.Errorf("resolve_audience: load campaign: %w", err)
	}
	aud, err := DecodeAudience(c.Audience)
	if err != nil {
		return 0, fmt.Errorf("resolve_audience: decode audience: %w", err)
	}

	var rows []*CampaignRecipient

	switch aud.Kind {
	case AudienceExplicit:
		for _, objID := range aud.ObjectIDs {
			ref, err := contactRefForObject(a.Deps.ObjectStore, c.OrgID, objID, c.Channel)
			if err != nil {
				// Skip unresolvable rows rather than aborting the whole campaign.
				continue
			}
			rows = append(rows, &CampaignRecipient{
				CampaignID: c.ID,
				OrgID:      c.OrgID,
				ObjectID:   objID,
				ContactRef: ref,
				Status:     string(RecipientQueued),
			})
		}

	case AudienceObjectQuery, "":
		typeSlug := aud.TypeSlug
		if typeSlug == "" {
			typeSlug = "contact"
		}
		// For v1, no advanced filter support — list all objects of the type
		// in the campaign's workspace. The Filter map will drive richer
		// matching once the object_query language is fleshed out.
		objs, _, _, err := a.Deps.ObjectStore.ListObjects(objectcore.ListObjectsParams{
			OrgID:       c.OrgID,
			WorkspaceID: c.WorkspaceID,
			TypeSlug:    typeSlug,
			PageSize:    5000, // see roadmap: pagination for very large audiences lands later
		})
		if err != nil {
			return 0, fmt.Errorf("resolve_audience: list objects: %w", err)
		}
		for _, obj := range objs {
			ref, ok := contactRefFromObjectData(obj.Data, c.Channel)
			if !ok {
				continue
			}
			rows = append(rows, &CampaignRecipient{
				CampaignID: c.ID,
				OrgID:      c.OrgID,
				ObjectID:   obj.ID,
				ContactRef: ref,
				Status:     string(RecipientQueued),
			})
		}

	default:
		return 0, fmt.Errorf("resolve_audience: unsupported audience kind %q", aud.Kind)
	}

	if err := a.Deps.Store.BulkInsertRecipients(rows); err != nil {
		return 0, fmt.Errorf("resolve_audience: persist: %w", err)
	}
	if err := a.Deps.Store.RecomputeStats(c.ID); err != nil {
		return len(rows), fmt.Errorf("resolve_audience: recompute stats: %w", err)
	}
	return len(rows), nil
}

// SendRecipientActivity sends one campaign message and updates that recipient's
// row. The Temporal activity retry policy handles transient send failures; a
// permanent failure (4xx from provider) flips the row to "failed" and returns.
type SendRecipientActivity struct {
	Deps ActivityDeps
}

// SendRecipientInput is the fan-out payload. Keeping it explicit (rather than
// re-deriving from DB inside the activity) means the workflow body controls
// exactly what goes over the wire and per-recipient retry doesn't re-fetch.
type SendRecipientInput struct {
	RecipientID string                 `json:"recipient_id"`
	CampaignID  string                 `json:"campaign_id"`
	OrgID       string                 `json:"org_id"`
	WorkspaceID string                 `json:"workspace_id"`
	Channel     string                 `json:"channel"`
	ContactRef  string                 `json:"contact_ref"`
	Payload     map[string]interface{} `json:"payload"`
}

func (a *SendRecipientActivity) SendRecipient(ctx context.Context, in SendRecipientInput) error {
	if a.Deps.AdapterRegistry == nil || a.Deps.ConversationStore == nil || a.Deps.MessageStore == nil {
		return errors.New("send_recipient: communication dependencies missing")
	}
	if in.ContactRef == "" {
		_ = a.Deps.Store.MarkRecipientFailed(in.RecipientID, "empty contact_ref")
		return nil
	}

	// One conversation per (channel, contact_ref) — find or create. Reuses
	// the existing inbound-webhook conversation if there is one so replies to
	// the campaign thread back to the same conversation.
	conv, err := a.Deps.ConversationStore.FindByChannelRef(in.OrgID, in.Channel, in.ContactRef)
	if err != nil || conv == nil {
		conv = &communication.Conversation{
			OrgID:       in.OrgID,
			WorkspaceID: in.WorkspaceID,
			Channel:     in.Channel,
			ChannelRef:  in.ContactRef,
			Status:      "open",
			Priority:    "normal",
		}
		conv, err = a.Deps.ConversationStore.Create(conv)
		if err != nil {
			_ = a.Deps.Store.MarkRecipientFailed(in.RecipientID, fmt.Sprintf("create conversation: %v", err))
			return fmt.Errorf("create conversation: %w", err)
		}
	}

	contentBytes, err := json.Marshal(in.Payload)
	if err != nil {
		_ = a.Deps.Store.MarkRecipientFailed(in.RecipientID, fmt.Sprintf("marshal payload: %v", err))
		return nil
	}

	msg := &communication.Message{
		ConversationID: conv.ID,
		SenderType:     "campaign",
		Channel:        in.Channel,
		ContentType:    contentTypeForChannel(in.Channel, in.Payload),
		Content:        contentBytes,
		Status:         "pending",
	}
	created, err := a.Deps.MessageStore.Create(msg)
	if err != nil {
		_ = a.Deps.Store.MarkRecipientFailed(in.RecipientID, fmt.Sprintf("persist message: %v", err))
		return fmt.Errorf("persist message: %w", err)
	}

	if err := a.Deps.AdapterRegistry.Dispatch(ctx, conv, created); err != nil {
		_ = a.Deps.MessageStore.UpdateStatus(created.ID, "failed")
		_ = a.Deps.Store.MarkRecipientFailed(in.RecipientID, err.Error())
		// Return the error so Temporal retries — provider transient failures
		// are common. The activity's retry policy will cap attempts.
		return fmt.Errorf("dispatch %s: %w", in.Channel, err)
	}

	_ = a.Deps.MessageStore.UpdateStatus(created.ID, "sent")
	if err := a.Deps.Store.MarkRecipientSent(in.RecipientID, created.ID); err != nil {
		// Logging only — the send succeeded.
		fmt.Printf("[campaigns] mark recipient sent %s: %v\n", in.RecipientID, err)
	}
	return nil
}

// FinaliseCampaignActivity refreshes the stats projection and moves the
// campaign to its terminal state. Called once the per-recipient fan-out drains.
type FinaliseCampaignActivity struct {
	Deps ActivityDeps
}

func (a *FinaliseCampaignActivity) FinaliseCampaign(ctx context.Context, campaignID, terminalStatus string) error {
	if err := a.Deps.Store.RecomputeStats(campaignID); err != nil {
		return fmt.Errorf("finalise: recompute stats: %w", err)
	}
	c, err := a.Deps.Store.lookupByIDOnly(campaignID)
	if err != nil {
		return fmt.Errorf("finalise: lookup: %w", err)
	}
	if terminalStatus == "" {
		terminalStatus = string(StatusCompleted)
	}
	if _, err := a.Deps.Store.SetStatus(c.ID, c.OrgID, CampaignStatus(terminalStatus)); err != nil {
		return fmt.Errorf("finalise: set status: %w", err)
	}
	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

// contactRefForObject extracts the channel-appropriate contact reference (phone
// for WA/SMS, email for email) from an Object Core record by ID.
func contactRefForObject(store *objectcore.ObjectCoreStore, orgID, objectID, channel string) (string, error) {
	obj, err := store.FindObjectByID(objectID, orgID)
	if err != nil {
		return "", err
	}
	ref, ok := contactRefFromObjectData(obj.Data, channel)
	if !ok {
		return "", fmt.Errorf("no %s contact ref on object %s", channelToRefField(channel), objectID)
	}
	return ref, nil
}

// contactRefFromObjectData inspects an Object's data JSONB for the appropriate
// contact reference. Looks at known field names — keep this conservative so we
// don't accidentally send to fields that aren't truly contact endpoints.
func contactRefFromObjectData(raw json.RawMessage, channel string) (string, bool) {
	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return "", false
	}
	field := channelToRefField(channel)
	if v, ok := data[field]; ok {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return s, true
		}
	}
	// Common alternatives.
	switch channel {
	case "whatsapp", "sms":
		for _, alt := range []string{"phone", "phone_number", "mobile", "msisdn"} {
			if v, ok := data[alt]; ok {
				if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
					return s, true
				}
			}
		}
	case "email":
		for _, alt := range []string{"email_address", "email_addr"} {
			if v, ok := data[alt]; ok {
				if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
					return s, true
				}
			}
		}
	}
	return "", false
}

func channelToRefField(channel string) string {
	switch channel {
	case "email":
		return "email"
	default:
		return "phone"
	}
}

func contentTypeForChannel(channel string, payload map[string]interface{}) string {
	switch channel {
	case "whatsapp":
		// If the payload looks like a WhatsApp template send, mark it as such
		// so the adapter knows to use the Templates API.
		if _, ok := payload["template"]; ok {
			return "template"
		}
		if _, ok := payload["template_name"]; ok {
			return "template"
		}
	}
	return "text"
}

// lookupByIDOnly is an internal helper that fetches by ID without an org filter
// — used inside activities where the activity input is already trusted to
// belong to the campaign's org. Defined here to keep store.go's public API
// org-scoped by default.
func (s *Store) lookupByIDOnly(id string) (*Campaign, error) {
	var c Campaign
	err := s.db.Where("id = ?", id).First(&c).Error
	return &c, err
}
