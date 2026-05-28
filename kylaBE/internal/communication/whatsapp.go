package communication

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"kyla-be/internal/objectcore"
	"kyla-be/shared/events"

	"github.com/gin-gonic/gin"
)

// ContactGateway is the subset of objectcore.ObjectCoreStore used by WhatsAppHandler.
type ContactGateway interface {
	FindObjectByDataField(orgID, workspaceID, typeSlug, key, value string) (*objectcore.Object, error)
	CreateObject(obj *objectcore.Object, actorID string) (*objectcore.Object, error)
}

// WhatsAppHandler processes inbound WhatsApp Cloud API webhooks.
type WhatsAppHandler struct {
	convStore     *ConversationStore
	msgStore      *MessageStore
	contactGW     ContactGateway
	eventBus      events.Publisher
	webhookSecret string // used for X-Hub-Signature-256 HMAC verification
	verifyToken   string // used for hub.verify_token GET verification
}

// NewWhatsAppHandler constructs a WhatsAppHandler.
func NewWhatsAppHandler(
	convStore *ConversationStore,
	msgStore *MessageStore,
	contactGW ContactGateway,
	eventBus events.Publisher,
	webhookSecret, verifyToken string,
) *WhatsAppHandler {
	return &WhatsAppHandler{
		convStore:     convStore,
		msgStore:      msgStore,
		contactGW:     contactGW,
		eventBus:      eventBus,
		webhookSecret: webhookSecret,
		verifyToken:   verifyToken,
	}
}

// ── WhatsApp payload structs ──────────────────────────────────────────────────
// These mirror the WhatsApp Cloud API webhook schema.

type waWebhookBody struct {
	Object string    `json:"object"`
	Entry  []waEntry `json:"entry"`
}

type waEntry struct {
	ID      string     `json:"id"` // WABA ID
	Changes []waChange `json:"changes"`
}

type waChange struct {
	Value waValue `json:"value"`
	Field string  `json:"field"`
}

type waValue struct {
	MessagingProduct string      `json:"messaging_product"`
	Metadata         waMetadata  `json:"metadata"`
	Contacts         []waContact `json:"contacts"`
	Messages         []waMessage `json:"messages"`
	Statuses         []waStatus  `json:"statuses"`
}

type waMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

type waContact struct {
	Profile waProfile `json:"profile"`
	WaID    string    `json:"wa_id"` // sender's WhatsApp number
}

type waProfile struct {
	Name string `json:"name"`
}

type waMessage struct {
	From      string    `json:"from"`      // sender's WA ID / phone number
	ID        string    `json:"id"`        // provider message ID
	Timestamp string    `json:"timestamp"`
	Type      string    `json:"type"`      // text | image | audio | video | document | sticker
	Text      *waText   `json:"text,omitempty"`
	Image     *waMedia  `json:"image,omitempty"`
	Audio     *waMedia  `json:"audio,omitempty"`
	Video     *waMedia  `json:"video,omitempty"`
	Document  *waMedia  `json:"document,omitempty"`
}

type waText struct {
	Body string `json:"body"`
}

type waMedia struct {
	ID       string `json:"id"`
	MimeType string `json:"mime_type"`
	Caption  string `json:"caption,omitempty"`
}

type waStatus struct {
	ID          string `json:"id"`     // provider message ID
	Status      string `json:"status"` // sent | delivered | read | failed
	RecipientID string `json:"recipient_id"`
}

// ── Gin handlers ──────────────────────────────────────────────────────────────

// Verify handles GET /webhooks/whatsapp — responds to Meta's hub challenge.
func (h *WhatsAppHandler) Verify(c *gin.Context) {
	mode      := c.Query("hub.mode")
	challenge := c.Query("hub.challenge")
	token     := c.Query("hub.verify_token")

	if mode == "subscribe" && token == h.verifyToken {
		c.String(http.StatusOK, challenge)
		return
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "invalid verify token"})
}

// Receive handles POST /webhooks/whatsapp — processes inbound messages.
func (h *WhatsAppHandler) Receive(c *gin.Context) {
	// Read body for HMAC verification + parsing.
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}

	// Verify HMAC-SHA256 signature when secret is configured.
	if h.webhookSecret != "" {
		sig := c.GetHeader("X-Hub-Signature-256")
		if !h.verifySignature(body, sig) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
	}

	var payload waWebhookBody
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// Acknowledge immediately — WhatsApp requires a 200 within 5 s.
	c.JSON(http.StatusOK, gin.H{"status": "ok"})

	// Process asynchronously so we never miss the 5-second window.
	go h.processPayload(payload)
}

// ── Async processing ──────────────────────────────────────────────────────────

func (h *WhatsAppHandler) processPayload(payload waWebhookBody) {
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			if change.Field != "messages" {
				continue
			}
			v := change.Value

			// Handle delivery status updates.
			for _, st := range v.Statuses {
				h.handleStatusUpdate(st)
			}

			// Handle inbound messages.
			for i, msg := range v.Messages {
				var contactName string
				if i < len(v.Contacts) {
					contactName = v.Contacts[i].Profile.Name
				}
				// orgID and workspaceID are resolved from the WABA phone_number_id.
				// In production these are looked up from an Apps/Integration record.
				// For now we embed them in the webhook URL as query params or use
				// a placeholder resolved from meta.
				h.handleInboundMessage(entry.ID, v.Metadata.PhoneNumberID, msg, contactName)
			}
		}
	}
}

// handleInboundMessage processes a single inbound WhatsApp message.
// wabaID is the WhatsApp Business Account ID (entry.ID).
// phoneNumberID identifies which WA number received the message.
func (h *WhatsAppHandler) handleInboundMessage(wabaID, phoneNumberID string, msg waMessage, contactName string) {
	// Resolve orgID + workspaceID from the phone_number_id.
	// Production: look up an "integration" record keyed by phone_number_id.
	// Development stub: use wabaID as orgID and a fixed workspaceID placeholder.
	orgID       := wabaID
	workspaceID := wabaID // replace with real lookup in integration store

	// Idempotency: skip if we already stored this provider message ID.
	existing, _ := h.msgStore.FindByExternalID(msg.ID)
	if existing != nil {
		return
	}

	// Find or create the contact Object Core record.
	contact, err := h.findOrCreateContact(orgID, workspaceID, msg.From, contactName)
	if err != nil {
		log.Printf("[whatsapp] find/create contact error (from=%s): %v", msg.From, err)
		return
	}

	// Find or create the conversation (keyed by org + channel + channel_ref=from number).
	conv, isNew, err := h.findOrCreateConversation(orgID, workspaceID, contact.ID, msg.From)
	if err != nil {
		log.Printf("[whatsapp] find/create conversation error (from=%s): %v", msg.From, err)
		return
	}

	// Build message content JSON.
	contentType, content := h.buildContent(msg)

	// Persist the message.
	newMsg := &Message{
		ConversationID: conv.ID,
		SenderID:       contact.ID,
		SenderType:     SenderContact,
		Channel:        ChannelWhatsApp,
		ContentType:    contentType,
		Content:        content,
		Status:         MsgStatusSent,
		ExternalID:     msg.ID,
	}
	if _, err := h.msgStore.Create(newMsg); err != nil {
		log.Printf("[whatsapp] store message error (conv=%s): %v", conv.ID, err)
		return
	}

	// Re-open snoozed/resolved conversation on new inbound.
	if conv.Status == StatusResolved || conv.Status == StatusSnoozed {
		_, _ = h.convStore.SetStatus(conv.ID, orgID, StatusOpen, nil)
	}

	// Publish NATS events.
	if isNew {
		h.emitEvent(orgID, workspaceID, conv.ID, "created", contact.ID, conv)
	}
	h.emitEvent(orgID, workspaceID, conv.ID, "message_received", contact.ID, newMsg)

	log.Printf("[whatsapp] inbound msg stored conv=%s msg=%s contact=%s", conv.ID, newMsg.ID, contact.ID)
}

// handleStatusUpdate processes a WhatsApp delivery/read receipt.
func (h *WhatsAppHandler) handleStatusUpdate(st waStatus) {
	msg, _ := h.msgStore.FindByExternalID(st.ID)
	if msg == nil {
		return
	}
	var newStatus string
	switch st.Status {
	case "delivered":
		newStatus = MsgStatusDelivered
	case "read":
		newStatus = MsgStatusRead
	case "failed":
		newStatus = MsgStatusFailed
	default:
		return
	}
	_ = h.msgStore.UpdateStatus(msg.ID, newStatus)
}

// findOrCreateContact returns the Object Core contact for the given phone number,
// creating a new one if none exists.
func (h *WhatsAppHandler) findOrCreateContact(orgID, workspaceID, phone, name string) (*objectcore.Object, error) {
	obj, err := h.contactGW.FindObjectByDataField(orgID, workspaceID, "contact", "phone", phone)
	if err != nil {
		return nil, err
	}
	if obj != nil {
		return obj, nil
	}

	// Create a new contact Object.
	data, _ := json.Marshal(map[string]string{
		"phone":   phone,
		"name":    name,
		"channel": ChannelWhatsApp,
	})
	newObj := &objectcore.Object{
		OrgID:       orgID,
		WorkspaceID: workspaceID,
		TypeSlug:    "contact",
		Data:        data,
		CreatedBy:   "system",
	}
	return h.contactGW.CreateObject(newObj, "system")
}

// findOrCreateConversation returns the open conversation for the given contact/phone,
// or creates a new one. Returns (conversation, isNew, error).
func (h *WhatsAppHandler) findOrCreateConversation(orgID, workspaceID, contactID, channelRef string) (*Conversation, bool, error) {
	existing, err := h.convStore.FindByChannelRef(orgID, ChannelWhatsApp, channelRef)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}

	// Create Object Core record for the conversation.
	ocData, _ := json.Marshal(map[string]string{
		"channel":     ChannelWhatsApp,
		"channel_ref": channelRef,
	})
	ocObj := &objectcore.Object{
		OrgID:       orgID,
		WorkspaceID: workspaceID,
		TypeSlug:    "conversation",
		Data:        ocData,
		CreatedBy:   "system",
	}
	created, err := h.contactGW.CreateObject(ocObj, "system")
	if err != nil {
		return nil, false, fmt.Errorf("create conversation object: %w", err)
	}

	conv := &Conversation{
		ID:          created.ID,
		OrgID:       orgID,
		WorkspaceID: workspaceID,
		Channel:     ChannelWhatsApp,
		ChannelRef:  channelRef,
		ContactID:   contactID,
		Status:      StatusOpen,
		Priority:    PriorityNormal,
		Meta:        json.RawMessage(`{}`),
	}
	saved, err := h.convStore.Create(conv)
	if err != nil {
		return nil, false, fmt.Errorf("create conversation: %w", err)
	}
	return saved, true, nil
}

// buildContent maps a WhatsApp message type to a ContentType string + JSONB content.
func (h *WhatsAppHandler) buildContent(msg waMessage) (string, json.RawMessage) {
	switch msg.Type {
	case "text":
		body := ""
		if msg.Text != nil {
			body = msg.Text.Body
		}
		b, _ := json.Marshal(map[string]string{"text": body})
		return ContentText, b
	case "image":
		b, _ := json.Marshal(msg.Image)
		return ContentImage, b
	case "audio":
		b, _ := json.Marshal(msg.Audio)
		return ContentAudio, b
	case "video":
		b, _ := json.Marshal(msg.Video)
		return ContentVideo, b
	case "document":
		b, _ := json.Marshal(msg.Document)
		return ContentFile, b
	default:
		b, _ := json.Marshal(map[string]string{"raw_type": msg.Type})
		return ContentText, b
	}
}

// verifySignature validates the X-Hub-Signature-256 header.
func (h *WhatsAppHandler) verifySignature(body []byte, sigHeader string) bool {
	const prefix = "sha256="
	if !strings.HasPrefix(sigHeader, prefix) {
		return false
	}
	expected := sigHeader[len(prefix):]
	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(body)
	actual := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(actual), []byte(expected))
}

// emitEvent publishes a domain event to the event bus, ignoring publish errors.
func (h *WhatsAppHandler) emitEvent(orgID, workspaceID, entityID, action, actorID string, payload interface{}) {
	ev, err := events.NewEvent(orgID, workspaceID, "conversation", action, entityID, actorID, payload)
	if err != nil {
		return
	}
	_ = h.eventBus.Publish(ev)
}

// Ensure time is used (imported via objectcore).
var _ = time.Now
