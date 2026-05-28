package communication

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/smtp"
	"time"

	"kyla-be/internal/objectcore"
	"kyla-be/shared/events"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/google/uuid"
)

// EmailAdapter sends emails via SMTP and polls IMAP for inbound messages.
type EmailAdapter struct {
	smtpHost     string
	smtpPort     string
	smtpUser     string
	smtpPassword string

	imapHost     string
	imapPort     string
	imapUser     string
	imapPassword string

	pollInterval time.Duration

	convStore      *ConversationStore
	msgStore       *MessageStore
	ocStore        ObjectCoreGateway
	eventBus       events.Publisher
	router         RouterInterface
	tenantResolver *TenantResolver
}

// RouterInterface will be implemented by Router in routing.go
type RouterInterface interface {
	Route(ctx context.Context, conv *Conversation) error
}

// NewEmailAdapter constructs an EmailAdapter.
func NewEmailAdapter(
	smtpHost, smtpPort, smtpUser, smtpPassword string,
	imapHost, imapPort, imapUser, imapPassword string,
	pollIntervalSecs int,
	convStore *ConversationStore,
	msgStore *MessageStore,
	ocStore ObjectCoreGateway,
	eventBus events.Publisher,
	router RouterInterface,
	tenantResolver *TenantResolver,
) *EmailAdapter {
	interval := time.Duration(pollIntervalSecs) * time.Second
	if pollIntervalSecs == 0 {
		interval = 60 * time.Second
	}
	return &EmailAdapter{
		smtpHost:       smtpHost,
		smtpPort:       smtpPort,
		smtpUser:       smtpUser,
		smtpPassword:   smtpPassword,
		imapHost:       imapHost,
		imapPort:       imapPort,
		imapUser:       imapUser,
		imapPassword:   imapPassword,
		pollInterval:   interval,
		convStore:      convStore,
		msgStore:       msgStore,
		ocStore:        ocStore,
		eventBus:       eventBus,
		router:         router,
		tenantResolver: tenantResolver,
	}
}

// Channel returns "email".
func (a *EmailAdapter) Channel() string {
	return ChannelEmail
}

// Send sends an email via SMTP.
func (a *EmailAdapter) Send(ctx context.Context, conv *Conversation, msg *Message) error {
	var content map[string]interface{}
	if err := json.Unmarshal(msg.Content, &content); err != nil {
		return fmt.Errorf("unmarshal email content: %w", err)
	}

	to, _ := content["to"].(string)
	subject, _ := content["subject"].(string)
	body, _ := content["body"].(string)

	if to == "" {
		to = conv.ChannelRef // fallback to conversation's email address
	}
	if to == "" {
		return fmt.Errorf("email missing recipient")
	}

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		a.smtpUser, to, subject, body)

	auth := smtp.PlainAuth("", a.smtpUser, a.smtpPassword, a.smtpHost)
	addr := fmt.Sprintf("%s:%s", a.smtpHost, a.smtpPort)

	tlsConfig := &tls.Config{ServerName: a.smtpHost}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("smtp tls dial: %w", err)
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, a.smtpHost)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Quit()

	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := c.Mail(a.smtpUser); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write([]byte(message)); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}

	log.Printf("[email] sent msg=%s to=%s", msg.ID, to)
	return nil
}

// StartIMAPPoller runs the IMAP polling loop in a goroutine.
func (a *EmailAdapter) StartIMAPPoller(ctx context.Context) {
	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[email] IMAP poller stopped")
			return
		case <-ticker.C:
			a.pollIMAP(ctx)
		}
	}
}

func (a *EmailAdapter) pollIMAP(ctx context.Context) {
	c, err := client.DialTLS(fmt.Sprintf("%s:%s", a.imapHost, a.imapPort), &tls.Config{ServerName: a.imapHost})
	if err != nil {
		log.Printf("[email] IMAP dial error: %v", err)
		return
	}
	defer c.Logout()

	if err := c.Login(a.imapUser, a.imapPassword); err != nil {
		log.Printf("[email] IMAP login error: %v", err)
		return
	}

	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Printf("[email] IMAP select INBOX error: %v", err)
		return
	}

	if mbox.Messages == 0 {
		return
	}

	seqset := new(imap.SeqSet)
	seqset.AddRange(1, mbox.Messages)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchRFC822}, messages)
	}()

	for msg := range messages {
		if hasFlag(msg.Flags, imap.SeenFlag) {
			continue
		}

		a.processInboundEmail(ctx, msg, c)
	}

	if err := <-done; err != nil {
		log.Printf("[email] IMAP fetch error: %v", err)
	}
}

func (a *EmailAdapter) processInboundEmail(ctx context.Context, imapMsg *imap.Message, c *client.Client) {
	from := ""
	if len(imapMsg.Envelope.From) > 0 {
		from = imapMsg.Envelope.From[0].Address()
	}
	subject := imapMsg.Envelope.Subject
	messageID := imapMsg.Envelope.MessageId

	// Read body
	section := &imap.BodySectionName{}
	r := imapMsg.GetBody(section)
	if r == nil {
		log.Printf("[email] no body for message %s", messageID)
		return
	}
	body, _ := io.ReadAll(r)

	// Check for duplicate (idempotency via Message-ID header)
	if messageID != "" {
		if existing, _ := a.msgStore.FindByExternalID(messageID); existing != nil {
			log.Printf("[email] duplicate message %s, skipping", messageID)
			return
		}
	}

	// TODO: Find or create contact from `from` email address
	// For now, use a placeholder contact_id
	contactID := uuid.New().String()

	// Find or create conversation using Message-ID as channel_ref
	conv, isNew, err := a.findOrCreateConversation(from, subject, messageID, contactID)
	if err != nil {
		log.Printf("[email] find/create conversation error: %v", err)
		return
	}

	// Create message
	content, _ := json.Marshal(map[string]interface{}{
		"from":    from,
		"subject": subject,
		"body":    string(body),
	})

	msg := &Message{
		ConversationID: conv.ID,
		SenderType:     SenderContact,
		Channel:        ChannelEmail,
		ContentType:    ContentTypeText,
		Content:        content,
		Status:         MsgStatusReceived,
		ExternalID:     messageID,
	}

	created, err := a.msgStore.Create(msg)
	if err != nil {
		log.Printf("[email] create message error: %v", err)
		return
	}

	// Publish NATS event
	a.publishEvent(conv.OrgID, conv.WorkspaceID, conv.ID, "message_received", "", created)

	// Mark as SEEN in IMAP
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(imapMsg.SeqNum)
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.SeenFlag}
	_ = c.Store(seqSet, item, flags, nil)

	// Route if new conversation
	if isNew && a.router != nil {
		go a.router.Route(context.Background(), conv)
	}

	log.Printf("[email] processed inbound email from=%s conv=%s msg=%s", from, conv.ID, created.ID)
}

func (a *EmailAdapter) findOrCreateConversation(from, subject, messageID, contactID string) (*Conversation, bool, error) {
	// Recommendation #1: Tenant extraction from webhook
	orgID := "00000000-0000-0000-0000-000000000000"       // fallback
	workspaceID := "00000000-0000-0000-0000-000000000000" // fallback

	if a.tenantResolver != nil {
		resolvedOrg, resolvedWs, err := a.tenantResolver.ResolveFromWebhook(nil, ChannelEmail, from)
		if err == nil {
			orgID = resolvedOrg
			workspaceID = resolvedWs
			log.Printf("[email] resolved tenant: org=%s workspace=%s", orgID, workspaceID)
		} else {
			log.Printf("[email] tenant resolution failed: %v, using fallback", err)
		}
	}

	// Try to find existing conversation by channel_ref (Message-ID)
	if messageID != "" {
		conv, err := a.convStore.FindByChannelRef(orgID, ChannelEmail, messageID)
		if err != nil {
			return nil, false, fmt.Errorf("lookup conversation: %w", err)
		}
		if conv != nil {
			return conv, false, nil
		}
	}

	// Create new conversation

	conv := &Conversation{
		OrgID:       orgID,
		WorkspaceID: workspaceID,
		Channel:     ChannelEmail,
		ChannelRef:  messageID,
		ContactID:   contactID,
		Status:      StatusOpen,
		Priority:    PriorityNormal,
		Subject:     subject,
		Meta:        []byte(`{}`),
	}

	created, err := a.convStore.Create(conv)
	if err != nil {
		return nil, false, err
	}

	// Create Object Core record
	ocObj := &objectcore.Object{
		OrgID:       created.OrgID,
		WorkspaceID: created.WorkspaceID,
		TypeSlug:    "conversation",
		Data:        []byte(fmt.Sprintf(`{"channel":"email","subject":%q}`, subject)),
		CreatedBy:   "system",
	}
	ocObj.ID = created.ID
	_, _ = a.ocStore.CreateObject(ocObj, "system")

	// Publish created event
	a.publishEvent(created.OrgID, created.WorkspaceID, created.ID, "created", "", created)

	return created, true, nil
}

func (a *EmailAdapter) publishEvent(orgID, workspaceID, entityID, action, actorID string, payload interface{}) {
	ev, err := events.NewEvent(orgID, workspaceID, "conversation", action, entityID, actorID, payload)
	if err != nil {
		log.Printf("[email] event build error (action=%s entity=%s): %v", action, entityID, err)
		return
	}
	if err := a.eventBus.Publish(ev); err != nil {
		log.Printf("[email] event publish error (action=%s entity=%s): %v", action, entityID, err)
	}
}

func hasFlag(flags []string, target string) bool {
	for _, f := range flags {
		if f == target {
			return true
		}
	}
	return false
}
