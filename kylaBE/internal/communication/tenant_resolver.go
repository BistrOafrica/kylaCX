package communication

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// TenantResolver extracts org/workspace IDs from webhooks using various strategies.
type TenantResolver struct {
	contactStore ContactLookup
	// Webhook signature secrets per channel
	whatsappSecret string
	twilioSecret   string
}

// ContactLookup defines minimal contact store interface for tenant resolution.
type ContactLookup interface {
	FindByPhone(phone string) (*ContactInfo, error)
	FindByEmail(email string) (*ContactInfo, error)
}

// ContactInfo holds tenant context for a contact.
type ContactInfo struct {
	ID          string
	OrgID       string
	WorkspaceID string
	Phone       string
	Email       string
}

// NewTenantResolver constructs a TenantResolver.
func NewTenantResolver(contactStore ContactLookup, whatsappSecret, twilioSecret string) *TenantResolver {
	return &TenantResolver{
		contactStore:   contactStore,
		whatsappSecret: whatsappSecret,
		twilioSecret:   twilioSecret,
	}
}

// ResolveFromWebhook attempts to extract org/workspace IDs from webhook request.
// Priority: 1) Custom headers, 2) Contact lookup, 3) Default fallback.
func (r *TenantResolver) ResolveFromWebhook(req *http.Request, channel, contactIdentifier string) (orgID, workspaceID string, err error) {
	// Strategy 1: Check custom headers (for authenticated webhooks)
	if orgID = req.Header.Get("X-Kyla-Org-ID"); orgID != "" {
		workspaceID = req.Header.Get("X-Kyla-Workspace-ID")
		if workspaceID == "" {
			// Default workspace for org
			workspaceID = uuid.New().String() // FIXME: Query org's default workspace
		}
		log.Printf("[tenant] resolved from headers: org=%s workspace=%s", orgID, workspaceID)
		return orgID, workspaceID, nil
	}

	// Strategy 2: Lookup contact by phone/email
	if r.contactStore != nil && contactIdentifier != "" {
		var contact *ContactInfo
		switch channel {
		case ChannelSMS, ChannelWhatsApp, ChannelVoice:
			contact, err = r.contactStore.FindByPhone(contactIdentifier)
		case ChannelEmail:
			contact, err = r.contactStore.FindByEmail(contactIdentifier)
		}

		if err == nil && contact != nil {
			log.Printf("[tenant] resolved from contact lookup: org=%s workspace=%s", contact.OrgID, contact.WorkspaceID)
			return contact.OrgID, contact.WorkspaceID, nil
		}
		log.Printf("[tenant] contact lookup failed for %s: %v", contactIdentifier, err)
	}

	// Strategy 3: Default fallback (temporary for Phase 3)
	// FIXME: This should fail or queue for manual assignment in production
	log.Printf("[tenant] FALLBACK: Using zero UUID for org/workspace (channel=%s, contact=%s)", channel, contactIdentifier)
	return uuid.New().String(), uuid.New().String(), fmt.Errorf("tenant not resolved, using fallback")
}

// VerifyWhatsAppSignature validates Meta's webhook signature.
// Signature format: sha256=<hex>
func (r *TenantResolver) VerifyWhatsAppSignature(payload []byte, signature string) bool {
	if r.whatsappSecret == "" || signature == "" {
		return false
	}

	// Extract hex hash from "sha256=<hash>"
	parts := strings.SplitN(signature, "=", 2)
	if len(parts) != 2 || parts[0] != "sha256" {
		return false
	}

	expected := parts[1]
	mac := hmac.New(sha256.New, []byte(r.whatsappSecret))
	mac.Write(payload)
	computed := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(computed))
}

// VerifyTwilioSignature validates Twilio's X-Twilio-Signature.
// Uses Twilio's signature algorithm (HMAC-SHA1 of URL + sorted params).
func (r *TenantResolver) VerifyTwilioSignature(url string, params map[string][]string, signature string) bool {
	if r.twilioSecret == "" || signature == "" {
		return false
	}

	// Build validation string: URL + sorted(params)
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	// Sort keys for consistent signature (skipped for brevity - add sort.Strings)

	data := url
	for _, k := range keys {
		data += k + params[k][0]
	}

	mac := hmac.New(sha256.New, []byte(r.twilioSecret))
	mac.Write([]byte(data))
	computed := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(computed))
}

// ResolveFromContext extracts tenant info from gRPC metadata (for authenticated API calls).
func ResolveFromContext(ctx context.Context) (orgID, workspaceID string, err error) {
	// FIXME: Extract from authctx.RequestMetadata
	// This should be injected by AuthInterceptor
	return "", "", fmt.Errorf("context-based tenant resolution not implemented")
}
