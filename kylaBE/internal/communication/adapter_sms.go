package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"kyla-be/internal/objectcore"
	"kyla-be/shared/events"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SMSProvider defines the interface for SMS sending.
type SMSProvider interface {
	Send(to, from, body string) error
	Name() string
}

// SMSAdapter routes outbound SMS to the configured provider.
type SMSAdapter struct {
	provider SMSProvider
}

// NewSMSAdapter constructs an SMSAdapter from a provider.
func NewSMSAdapter(provider SMSProvider) *SMSAdapter {
	return &SMSAdapter{provider: provider}
}

// Channel returns "sms".
func (a *SMSAdapter) Channel() string {
	return ChannelSMS
}

// Send dispatches SMS via the configured provider.
func (a *SMSAdapter) Send(ctx context.Context, conv *Conversation, msg *Message) error {
	var content map[string]interface{}
	if err := json.Unmarshal(msg.Content, &content); err != nil {
		return fmt.Errorf("unmarshal sms content: %w", err)
	}

	to := conv.ChannelRef // recipient phone number
	body, _ := content["text"].(string)
	from, _ := content["from"].(string)

	if to == "" {
		return fmt.Errorf("sms missing recipient")
	}
	if body == "" {
		return fmt.Errorf("sms missing text")
	}

	if err := a.provider.Send(to, from, body); err != nil {
		return fmt.Errorf("sms provider %s: %w", a.provider.Name(), err)
	}

	log.Printf("[sms] sent via %s msg=%s to=%s", a.provider.Name(), msg.ID, to)
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Twilio Provider
// ─────────────────────────────────────────────────────────────────────────────

// TwilioProvider sends SMS via Twilio.
type TwilioProvider struct {
	accountSID string
	authToken  string
	fromNumber string
	httpClient *http.Client
}

// NewTwilioProvider constructs a TwilioProvider.
func NewTwilioProvider(accountSID, authToken, fromNumber string) *TwilioProvider {
	return &TwilioProvider{
		accountSID: accountSID,
		authToken:  authToken,
		fromNumber: fromNumber,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// Name returns "twilio".
func (p *TwilioProvider) Name() string {
	return "twilio"
}

// Send sends an SMS via Twilio.
func (p *TwilioProvider) Send(to, from, body string) error {
	if from == "" {
		from = p.fromNumber
	}

	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", p.accountSID)
	data := url.Values{}
	data.Set("To", to)
	data.Set("From", from)
	data.Set("Body", body)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.SetBasicAuth(p.accountSID, p.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("twilio api call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("twilio error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Africa's Talking Provider
// ─────────────────────────────────────────────────────────────────────────────

// AfricasTalkingProvider sends SMS via Africa's Talking.
type AfricasTalkingProvider struct {
	apiKey     string
	username   string
	fromNumber string
	httpClient *http.Client
}

// NewAfricasTalkingProvider constructs an AfricasTalkingProvider.
func NewAfricasTalkingProvider(apiKey, username, fromNumber string) *AfricasTalkingProvider {
	return &AfricasTalkingProvider{
		apiKey:     apiKey,
		username:   username,
		fromNumber: fromNumber,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// Name returns "africastalking".
func (p *AfricasTalkingProvider) Name() string {
	return "africastalking"
}

// Send sends an SMS via Africa's Talking.
func (p *AfricasTalkingProvider) Send(to, from, body string) error {
	if from == "" {
		from = p.fromNumber
	}

	apiURL := "https://api.africastalking.com/version1/messaging"
	data := url.Values{}
	data.Set("username", p.username)
	data.Set("to", to)
	data.Set("message", body)
	if from != "" {
		data.Set("from", from)
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("apiKey", p.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("africastalking api call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("africastalking error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// SMS Webhook Handlers (Inbound)
// ─────────────────────────────────────────────────────────────────────────────

// SMSWebhookHandler processes inbound SMS from Twilio and Africa's Talking.
type SMSWebhookHandler struct {
	convStore      *ConversationStore
	msgStore       *MessageStore
	ocStore        ObjectCoreGateway
	eventBus       events.Publisher
	router         RouterInterface
	tenantResolver *TenantResolver
}

// NewSMSWebhookHandler constructs an SMSWebhookHandler.
func NewSMSWebhookHandler(
	convStore *ConversationStore,
	msgStore *MessageStore,
	ocStore ObjectCoreGateway,
	eventBus events.Publisher,
	router RouterInterface,
	tenantResolver *TenantResolver,
) *SMSWebhookHandler {
	return &SMSWebhookHandler{
		convStore:      convStore,
		msgStore:       msgStore,
		ocStore:        ocStore,
		eventBus:       eventBus,
		router:         router,
		tenantResolver: tenantResolver,
	}
}

// ReceiveTwilio handles Twilio inbound SMS webhook.
func (h *SMSWebhookHandler) ReceiveTwilio(c *gin.Context) {
	from := c.PostForm("From")
	to := c.PostForm("To")
	body := c.PostForm("Body")
	messageSid := c.PostForm("MessageSid")

	if from == "" || body == "" {
		c.JSON(400, gin.H{"error": "missing required fields"})
		return
	}

	// Check for duplicate
	if messageSid != "" {
		if existing, _ := h.msgStore.FindByExternalID(messageSid); existing != nil {
			c.String(200, "OK")
			return
		}
	}

	// TODO: Find or create contact from phone number
	contactID := uuid.New().String()

	conv, isNew, err := h.findOrCreateConversation(from, to, contactID)
	if err != nil {
		log.Printf("[sms] twilio find/create conversation error: %v", err)
		c.JSON(500, gin.H{"error": "internal error"})
		return
	}

	content, _ := json.Marshal(map[string]interface{}{
		"text": body,
		"from": from,
		"to":   to,
	})

	msg := &Message{
		ConversationID: conv.ID,
		SenderType:     SenderContact,
		Channel:        ChannelSMS,
		ContentType:    ContentTypeText,
		Content:        content,
		Status:         MsgStatusReceived,
		ExternalID:     messageSid,
	}

	created, err := h.msgStore.Create(msg)
	if err != nil {
		log.Printf("[sms] twilio create message error: %v", err)
		c.JSON(500, gin.H{"error": "internal error"})
		return
	}

	h.publishEvent(conv.OrgID, conv.WorkspaceID, conv.ID, "message_received", "", created)

	if isNew && h.router != nil {
		go h.router.Route(context.Background(), conv)
	}

	log.Printf("[sms] twilio inbound from=%s conv=%s msg=%s", from, conv.ID, created.ID)
	c.String(200, "OK")
}

// ReceiveAT handles Africa's Talking inbound SMS webhook.
func (h *SMSWebhookHandler) ReceiveAT(c *gin.Context) {
	from := c.PostForm("from")
	to := c.PostForm("to")
	text := c.PostForm("text")
	messageID := c.PostForm("id")

	if from == "" || text == "" {
		c.JSON(400, gin.H{"error": "missing required fields"})
		return
	}

	// Check for duplicate
	if messageID != "" {
		if existing, _ := h.msgStore.FindByExternalID(messageID); existing != nil {
			c.String(200, "OK")
			return
		}
	}

	// TODO: Find or create contact
	contactID := uuid.New().String()

	conv, isNew, err := h.findOrCreateConversation(from, to, contactID)
	if err != nil {
		log.Printf("[sms] at find/create conversation error: %v", err)
		c.JSON(500, gin.H{"error": "internal error"})
		return
	}

	content, _ := json.Marshal(map[string]interface{}{
		"text": text,
		"from": from,
		"to":   to,
	})

	msg := &Message{
		ConversationID: conv.ID,
		SenderType:     SenderContact,
		Channel:        ChannelSMS,
		ContentType:    ContentTypeText,
		Content:        content,
		Status:         MsgStatusReceived,
		ExternalID:     messageID,
	}

	created, err := h.msgStore.Create(msg)
	if err != nil {
		log.Printf("[sms] at create message error: %v", err)
		c.JSON(500, gin.H{"error": "internal error"})
		return
	}

	h.publishEvent(conv.OrgID, conv.WorkspaceID, conv.ID, "message_received", "", created)

	if isNew && h.router != nil {
		go h.router.Route(context.Background(), conv)
	}

	log.Printf("[sms] africastalking inbound from=%s conv=%s msg=%s", from, conv.ID, created.ID)
	c.String(200, "OK")
}

func (h *SMSWebhookHandler) findOrCreateConversation(from, to, contactID string) (*Conversation, bool, error) {
	// Recommendation #1: Tenant extraction from webhook
	orgID := "00000000-0000-0000-0000-000000000000"       // fallback
	workspaceID := "00000000-0000-0000-0000-000000000000" // fallback

	if h.tenantResolver != nil {
		resolvedOrg, resolvedWs, err := h.tenantResolver.ResolveFromWebhook(nil, ChannelSMS, from)
		if err == nil {
			orgID = resolvedOrg
			workspaceID = resolvedWs
			log.Printf("[sms] resolved tenant: org=%s workspace=%s", orgID, workspaceID)
		} else {
			log.Printf("[sms] tenant resolution failed: %v, using fallback", err)
		}
	}

	conv, err := h.convStore.FindByChannelRef(orgID, ChannelSMS, from)
	if err != nil {
		return nil, false, fmt.Errorf("lookup conversation: %w", err)
	}
	if conv != nil {
		return conv, false, nil
	}

	// Create new conversation
	conv = &Conversation{
		OrgID:       orgID,
		WorkspaceID: workspaceID,
		Channel:     ChannelSMS,
		ChannelRef:  from,
		ContactID:   contactID,
		Status:      StatusOpen,
		Priority:    PriorityNormal,
		Subject:     fmt.Sprintf("SMS from %s", from),
		Meta:        []byte(`{}`),
	}

	created, err := h.convStore.Create(conv)
	if err != nil {
		return nil, false, err
	}

	ocObj := &objectcore.Object{
		OrgID:       created.OrgID,
		WorkspaceID: created.WorkspaceID,
		TypeSlug:    "conversation",
		Data:        []byte(fmt.Sprintf(`{"channel":"sms","from":%q}`, from)),
		CreatedBy:   "system",
	}
	ocObj.ID = created.ID
	_, _ = h.ocStore.CreateObject(ocObj, "system")

	h.publishEvent(created.OrgID, created.WorkspaceID, created.ID, "created", "", created)

	return created, true, nil
}

func (h *SMSWebhookHandler) publishEvent(orgID, workspaceID, entityID, action, actorID string, payload interface{}) {
	ev, err := events.NewEvent(orgID, workspaceID, "conversation", action, entityID, actorID, payload)
	if err != nil {
		log.Printf("[sms] event build error (action=%s entity=%s): %v", action, entityID, err)
		return
	}
	if err := h.eventBus.Publish(ev); err != nil {
		log.Printf("[sms] event publish error (action=%s entity=%s): %v", action, entityID, err)
	}
}
