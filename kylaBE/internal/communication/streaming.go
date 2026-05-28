package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"kyla-be/pkg/pb"
	"kyla-be/shared/events"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StreamingServer handles realtime gRPC streaming.
type StreamingServer struct {
	streamBus events.StreamBus
	convStore *ConversationStore
	msgStore  *MessageStore
	eventBus  events.Publisher
}

// NewStreamingServer constructs a StreamingServer.
func NewStreamingServer(
	streamBus events.StreamBus,
	convStore *ConversationStore,
	msgStore *MessageStore,
	eventBus events.Publisher,
) *StreamingServer {
	return &StreamingServer{
		streamBus: streamBus,
		convStore: convStore,
		msgStore:  msgStore,
		eventBus:  eventBus,
	}
}

// StreamConversationUpdates streams live updates to connected gRPC clients.
func (s *StreamingServer) StreamConversationUpdates(
	req *pb.StreamConversationUpdatesRequest,
	stream pb.ConversationService_StreamConversationUpdatesServer,
) error {
	subject := fmt.Sprintf("kyla.%s.conversation.>", req.GetOrgId())
	ctx := stream.Context()
	ch := make(chan *events.DomainEvent, 64)

	// Subscribe to NATS
	sub, err := s.streamBus.SubscribeRaw(subject, func(ev *events.DomainEvent) error {
		select {
		case ch <- ev:
		default:
			// Slow client — drop event to avoid blocking NATS
			log.Printf("[streaming] dropped event for slow client org=%s", req.GetOrgId())
		}
		return nil
	})
	if err != nil {
		return status.Errorf(codes.Internal, "subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	log.Printf("[streaming] client connected org=%s workspace=%s", req.GetOrgId(), req.GetWorkspaceId())

	for {
		select {
		case <-ctx.Done():
			log.Printf("[streaming] client disconnected org=%s", req.GetOrgId())
			return ctx.Err()

		case ev := <-ch:
			// Filter by workspace if specified
			if req.GetWorkspaceId() != "" && ev.WorkspaceID != req.GetWorkspaceId() {
				continue
			}

			update, err := s.domainEventToConversationUpdate(ev)
			if err != nil {
				log.Printf("[streaming] convert event error: %v", err)
				continue
			}
			if update == nil {
				continue
			}

			if err := stream.Send(update); err != nil {
				log.Printf("[streaming] send error: %v", err)
				return err
			}
		}
	}
}

func (s *StreamingServer) domainEventToConversationUpdate(ev *events.DomainEvent) (*pb.ConversationUpdate, error) {
	update := &pb.ConversationUpdate{
		EventType:      "",
		ConversationId: ev.EntityID,
		Timestamp:      ev.OccurredAt.Unix(),
	}

	switch ev.Action {
	case "created", "updated", "assigned", "resolved":
		conv, err := s.convStore.FindByID(ev.EntityID, ev.OrgID)
		if err != nil {
			return nil, fmt.Errorf("fetch conversation: %w", err)
		}
		update.EventType = "conversation.updated"
		update.Conversation = ConversationToPb(conv)

	case "message_received":
		// Extract message from payload
		var msgData *Message
		if err := json.Unmarshal(ev.Payload, &msgData); err != nil {
			return nil, fmt.Errorf("unmarshal message payload: %w", err)
		}
		update.EventType = "message.received"
		update.Message = MessageToPb(msgData)

	case "typing":
		// Pass through typing indicator
		update.EventType = "typing"
		payloadStr := string(ev.Payload)
		update.Payload = &payloadStr

	case "presence_update":
		update.EventType = "presence_update"
		payloadStr := string(ev.Payload)
		update.Payload = &payloadStr

	default:
		// Unknown action, skip
		return nil, nil
	}

	return update, nil
}

// SendTypingIndicator publishes a typing indicator event (ephemeral NATS).
func (s *StreamingServer) SendTypingIndicator(
	ctx context.Context,
	req *pb.SendTypingIndicatorRequest,
) (*pb.SendTypingIndicatorResponse, error) {
	payload := map[string]interface{}{
		"user_id":         req.GetUserId(),
		"conversation_id": req.GetConversationId(),
		"is_typing":       req.GetIsTyping(),
	}

	ev, err := events.NewEvent(
		req.GetOrgId(),
		req.GetWorkspaceId(),
		"conversation",
		"typing",
		req.GetConversationId(),
		req.GetUserId(),
		payload,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "build typing event: %v", err)
	}

	// Publish typing event via event bus
	if err := s.eventBus.Publish(ev); err != nil {
		return nil, status.Errorf(codes.Internal, "publish typing event: %v", err)
	}

	log.Printf("[streaming] typing indicator conv=%s user=%s is_typing=%v",
		req.GetConversationId(), req.GetUserId(), req.GetIsTyping())

	return &pb.SendTypingIndicatorResponse{Success: true}, nil
}
