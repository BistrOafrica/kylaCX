package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"kyla-be/internal/objectcore"
	"kyla-be/shared/events"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: restrict origin in production
	},
}

// WebChatAdapter manages WebSocket connections for web chat.
type WebChatAdapter struct {
	registry  *ChatSessionRegistry
	convStore *ConversationStore
	msgStore  *MessageStore
	ocStore   ObjectCoreGateway
	eventBus  events.Publisher
	router    RouterInterface
}

// NewWebChatAdapter constructs a WebChatAdapter.
func NewWebChatAdapter(
	convStore *ConversationStore,
	msgStore *MessageStore,
	ocStore ObjectCoreGateway,
	eventBus events.Publisher,
	router RouterInterface,
) *WebChatAdapter {
	return &WebChatAdapter{
		registry:  NewChatSessionRegistry(),
		convStore: convStore,
		msgStore:  msgStore,
		ocStore:   ocStore,
		eventBus:  eventBus,
		router:    router,
	}
}

// Channel returns "webchat".
func (a *WebChatAdapter) Channel() string {
	return ChannelWebChat
}

// Send pushes a message to the WebSocket connection for the conversation.
func (a *WebChatAdapter) Send(ctx context.Context, conv *Conversation, msg *Message) error {
	conn := a.registry.Get(conv.ID)
	if conn == nil {
		return fmt.Errorf("no active websocket for conversation %s", conv.ID)
	}

	payload := map[string]interface{}{
		"type":    "message",
		"message": MessageToPb(msg),
	}

	if err := conn.WriteJSON(payload); err != nil {
		log.Printf("[webchat] send error conv=%s: %v", conv.ID, err)
		a.registry.Remove(conv.ID)
		return err
	}

	log.Printf("[webchat] sent msg=%s conv=%s", msg.ID, conv.ID)
	return nil
}

// HandleWebSocket handles WebSocket upgrade and message processing.
func (a *WebChatAdapter) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[webchat] upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Wait for handshake message with org_id, workspace_id, contact_id
	var handshake struct {
		OrgID       string `json:"org_id"`
		WorkspaceID string `json:"workspace_id"`
		ContactID   string `json:"contact_id"`
	}

	if err := conn.ReadJSON(&handshake); err != nil {
		log.Printf("[webchat] handshake error: %v", err)
		return
	}

	if handshake.OrgID == "" || handshake.WorkspaceID == "" {
		log.Println("[webchat] handshake missing required fields")
		return
	}

	// Find or create conversation
	conv, isNew, err := a.findOrCreateConversation(handshake.OrgID, handshake.WorkspaceID, handshake.ContactID)
	if err != nil {
		log.Printf("[webchat] find/create conversation error: %v", err)
		return
	}

	// Register connection
	a.registry.Set(conv.ID, conn)
	defer a.registry.Remove(conv.ID)

	// Send handshake response
	_ = conn.WriteJSON(map[string]interface{}{
		"type":            "handshake",
		"conversation_id": conv.ID,
	})

	// Route if new
	if isNew && a.router != nil {
		go a.router.Route(context.Background(), conv)
	}

	// Read loop
	for {
		var inbound struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		}

		if err := conn.ReadJSON(&inbound); err != nil {
			log.Printf("[webchat] read error conv=%s: %v", conv.ID, err)
			break
		}

		if inbound.Type != "message" {
			continue
		}

		// Create message
		content, _ := json.Marshal(map[string]interface{}{
			"text": inbound.Content,
		})

		msg := &Message{
			ConversationID: conv.ID,
			SenderType:     SenderContact,
			Channel:        ChannelWebChat,
			ContentType:    ContentTypeText,
			Content:        content,
			Status:         MsgStatusReceived,
		}

		created, err := a.msgStore.Create(msg)
		if err != nil {
			log.Printf("[webchat] create message error: %v", err)
			continue
		}

		a.publishEvent(conv.OrgID, conv.WorkspaceID, conv.ID, "message_received", "", created)
		log.Printf("[webchat] inbound msg=%s conv=%s", created.ID, conv.ID)
	}
}

func (a *WebChatAdapter) findOrCreateConversation(orgID, workspaceID, contactID string) (*Conversation, bool, error) {
	// TODO: Try to find existing open conversation for this contact
	// For now, we'll just create a new conversation every time

	// Create new conversation
	conv := &Conversation{
		OrgID:       orgID,
		WorkspaceID: workspaceID,
		Channel:     ChannelWebChat,
		ChannelRef:  "",
		ContactID:   contactID,
		Status:      StatusOpen,
		Priority:    PriorityNormal,
		Subject:     "Web Chat",
		Meta:        []byte(`{}`),
	}

	created, err := a.convStore.Create(conv)
	if err != nil {
		return nil, false, err
	}

	ocObj := &objectcore.Object{
		OrgID:       created.OrgID,
		WorkspaceID: created.WorkspaceID,
		TypeSlug:    "conversation",
		Data:        []byte(`{"channel":"webchat"}`),
		CreatedBy:   "system",
	}
	ocObj.ID = created.ID
	_, _ = a.ocStore.CreateObject(ocObj, "system")

	a.publishEvent(created.OrgID, created.WorkspaceID, created.ID, "created", "", created)

	return created, true, nil
}

func (a *WebChatAdapter) publishEvent(orgID, workspaceID, entityID, action, actorID string, payload interface{}) {
	ev, err := events.NewEvent(orgID, workspaceID, "conversation", action, entityID, actorID, payload)
	if err != nil {
		log.Printf("[webchat] event build error (action=%s entity=%s): %v", action, entityID, err)
		return
	}
	if err := a.eventBus.Publish(ev); err != nil {
		log.Printf("[webchat] event publish error (action=%s entity=%s): %v", action, entityID, err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ChatSessionRegistry
// ─────────────────────────────────────────────────────────────────────────────

// ChatSessionRegistry manages active WebSocket connections by conversation ID.
type ChatSessionRegistry struct {
	mu    sync.RWMutex
	conns map[string]*websocket.Conn
}

// NewChatSessionRegistry constructs a ChatSessionRegistry.
func NewChatSessionRegistry() *ChatSessionRegistry {
	return &ChatSessionRegistry{
		conns: make(map[string]*websocket.Conn),
	}
}

// Set registers a connection for a conversation.
func (r *ChatSessionRegistry) Set(conversationID string, conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conns[conversationID] = conn
}

// Get retrieves a connection for a conversation.
func (r *ChatSessionRegistry) Get(conversationID string) *websocket.Conn {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.conns[conversationID]
}

// Remove removes a connection for a conversation.
func (r *ChatSessionRegistry) Remove(conversationID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.conns, conversationID)
}
