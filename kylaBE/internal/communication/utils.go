package communication

import (
	"time"

	"kyla-be/internal/objectcore"
	"kyla-be/pkg/pb"
)

// ── Enum converters: string ↔ pb ─────────────────────────────────────────────

func channelToPb(s string) pb.Channel {
	switch s {
	case ChannelWhatsApp:
		return pb.Channel_CHANNEL_WHATSAPP
	case ChannelEmail:
		return pb.Channel_CHANNEL_EMAIL
	case ChannelSMS:
		return pb.Channel_CHANNEL_SMS
	case ChannelVoice:
		return pb.Channel_CHANNEL_VOICE
	case ChannelWebChat:
		return pb.Channel_CHANNEL_WEBCHAT
	case ChannelInstagram:
		return pb.Channel_CHANNEL_INSTAGRAM
	case ChannelMessenger:
		return pb.Channel_CHANNEL_MESSENGER
	default:
		return pb.Channel_CHANNEL_UNSPECIFIED
	}
}

func channelFromPb(c pb.Channel) string {
	switch c {
	case pb.Channel_CHANNEL_WHATSAPP:
		return ChannelWhatsApp
	case pb.Channel_CHANNEL_EMAIL:
		return ChannelEmail
	case pb.Channel_CHANNEL_SMS:
		return ChannelSMS
	case pb.Channel_CHANNEL_VOICE:
		return ChannelVoice
	case pb.Channel_CHANNEL_WEBCHAT:
		return ChannelWebChat
	case pb.Channel_CHANNEL_INSTAGRAM:
		return ChannelInstagram
	case pb.Channel_CHANNEL_MESSENGER:
		return ChannelMessenger
	default:
		return ChannelWhatsApp
	}
}

func statusToPb(s string) pb.ConversationStatus {
	switch s {
	case StatusOpen:
		return pb.ConversationStatus_CONVERSATION_STATUS_OPEN
	case StatusPending:
		return pb.ConversationStatus_CONVERSATION_STATUS_PENDING
	case StatusResolved:
		return pb.ConversationStatus_CONVERSATION_STATUS_RESOLVED
	case StatusSnoozed:
		return pb.ConversationStatus_CONVERSATION_STATUS_SNOOZED
	default:
		return pb.ConversationStatus_CONVERSATION_STATUS_UNSPECIFIED
	}
}

func statusFromPb(s pb.ConversationStatus) string {
	switch s {
	case pb.ConversationStatus_CONVERSATION_STATUS_OPEN:
		return StatusOpen
	case pb.ConversationStatus_CONVERSATION_STATUS_PENDING:
		return StatusPending
	case pb.ConversationStatus_CONVERSATION_STATUS_RESOLVED:
		return StatusResolved
	case pb.ConversationStatus_CONVERSATION_STATUS_SNOOZED:
		return StatusSnoozed
	default:
		return StatusOpen
	}
}

func priorityToPb(s string) pb.ConversationPriority {
	switch s {
	case PriorityLow:
		return pb.ConversationPriority_CONVERSATION_PRIORITY_LOW
	case PriorityNormal:
		return pb.ConversationPriority_CONVERSATION_PRIORITY_NORMAL
	case PriorityHigh:
		return pb.ConversationPriority_CONVERSATION_PRIORITY_HIGH
	case PriorityUrgent:
		return pb.ConversationPriority_CONVERSATION_PRIORITY_URGENT
	default:
		return pb.ConversationPriority_CONVERSATION_PRIORITY_NORMAL
	}
}

func priorityFromPb(p pb.ConversationPriority) string {
	switch p {
	case pb.ConversationPriority_CONVERSATION_PRIORITY_LOW:
		return PriorityLow
	case pb.ConversationPriority_CONVERSATION_PRIORITY_HIGH:
		return PriorityHigh
	case pb.ConversationPriority_CONVERSATION_PRIORITY_URGENT:
		return PriorityUrgent
	default:
		return PriorityNormal
	}
}

func senderTypeToPb(s string) pb.SenderType {
	switch s {
	case SenderAgent:
		return pb.SenderType_SENDER_TYPE_AGENT
	case SenderContact:
		return pb.SenderType_SENDER_TYPE_CONTACT
	case SenderBot:
		return pb.SenderType_SENDER_TYPE_BOT
	case SenderSystem:
		return pb.SenderType_SENDER_TYPE_SYSTEM
	default:
		return pb.SenderType_SENDER_TYPE_AGENT
	}
}

func msgStatusToPb(s string) pb.MessageStatus {
	switch s {
	case MsgStatusPending:
		return pb.MessageStatus_MESSAGE_STATUS_PENDING
	case MsgStatusSent:
		return pb.MessageStatus_MESSAGE_STATUS_SENT
	case MsgStatusDelivered:
		return pb.MessageStatus_MESSAGE_STATUS_DELIVERED
	case MsgStatusRead:
		return pb.MessageStatus_MESSAGE_STATUS_READ
	case MsgStatusFailed:
		return pb.MessageStatus_MESSAGE_STATUS_FAILED
	default:
		return pb.MessageStatus_MESSAGE_STATUS_SENT
	}
}

func contentTypeToPb(s string) pb.ContentType {
	switch s {
	case ContentImage:
		return pb.ContentType_CONTENT_TYPE_IMAGE
	case ContentAudio:
		return pb.ContentType_CONTENT_TYPE_AUDIO
	case ContentVideo:
		return pb.ContentType_CONTENT_TYPE_VIDEO
	case ContentFile:
		return pb.ContentType_CONTENT_TYPE_FILE
	case ContentTemplate:
		return pb.ContentType_CONTENT_TYPE_TEMPLATE
	case ContentInteractive:
		return pb.ContentType_CONTENT_TYPE_INTERACTIVE
	default:
		return pb.ContentType_CONTENT_TYPE_TEXT
	}
}

func contentTypeFromPb(c pb.ContentType) string {
	switch c {
	case pb.ContentType_CONTENT_TYPE_IMAGE:
		return ContentImage
	case pb.ContentType_CONTENT_TYPE_AUDIO:
		return ContentAudio
	case pb.ContentType_CONTENT_TYPE_VIDEO:
		return ContentVideo
	case pb.ContentType_CONTENT_TYPE_FILE:
		return ContentFile
	case pb.ContentType_CONTENT_TYPE_TEMPLATE:
		return ContentTemplate
	case pb.ContentType_CONTENT_TYPE_INTERACTIVE:
		return ContentInteractive
	default:
		return ContentText
	}
}

// ── Model → proto converters ──────────────────────────────────────────────────

// ConversationToPb converts a Conversation model to its proto representation.
func ConversationToPb(c *Conversation) *pb.Conversation {
	p := &pb.Conversation{
		Id:          c.ID,
		OrgId:       c.OrgID,
		WorkspaceId: c.WorkspaceID,
		Channel:     channelToPb(c.Channel),
		ChannelRef:  c.ChannelRef,
		ContactId:   c.ContactID,
		AssignedTo:  c.AssignedTo,
		TeamId:      c.TeamID,
		Status:      statusToPb(c.Status),
		Priority:    priorityToPb(c.Priority),
		Subject:     c.Subject,
		Meta:        string(c.Meta),
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
	}
	if c.SLADeadline != nil {
		s := c.SLADeadline.Format(time.RFC3339)
		p.SlaDeadline = &s
	}
	if c.SnoozedUntil != nil {
		s := c.SnoozedUntil.Format(time.RFC3339)
		p.SnoozedUntil = &s
	}
	if c.ResolvedAt != nil {
		s := c.ResolvedAt.Format(time.RFC3339)
		p.ResolvedAt = &s
	}
	return p
}

// ConversationsToPb converts a slice of Conversation models.
func ConversationsToPb(convs []*Conversation) []*pb.Conversation {
	out := make([]*pb.Conversation, len(convs))
	for i, c := range convs {
		out[i] = ConversationToPb(c)
	}
	return out
}

// MessageToPb converts a Message model to its proto representation.
func MessageToPb(m *Message) *pb.Message {
	return &pb.Message{
		Id:             m.ID,
		ConversationId: m.ConversationID,
		SenderId:       m.SenderID,
		SenderType:     senderTypeToPb(m.SenderType),
		Channel:        channelToPb(m.Channel),
		ContentType:    contentTypeToPb(m.ContentType),
		Content:        string(m.Content),
		Status:         msgStatusToPb(m.Status),
		ExternalId:     m.ExternalID,
		CreatedAt:      m.CreatedAt.Format(time.RFC3339),
	}
}

// MessagesToPb converts a slice of Message models.
func MessagesToPb(msgs []*Message) []*pb.Message {
	out := make([]*pb.Message, len(msgs))
	for i, m := range msgs {
		out[i] = MessageToPb(m)
	}
	return out
}

// TimelineEventToPb converts an ObjectEvent to a pb.TimelineEvent.
func TimelineEventToPb(e *objectcore.ObjectEvent) *pb.TimelineEvent {
	return &pb.TimelineEvent{
		Id:        e.ID,
		ObjectId:  e.ObjectID,
		ActorId:   e.ActorID,
		ActorType: e.ActorType,
		EventType: e.EventType,
		Payload:   string(e.Payload),
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
	}
}

// RoutingRuleToPb converts a RoutingRule model to its proto representation.
func RoutingRuleToPb(r *RoutingRule) *pb.RoutingRule {
	return &pb.RoutingRule{
		Id:          r.ID,
		OrgId:       r.OrgID,
		WorkspaceId: r.WorkspaceID,
		Name:        r.Name,
		Priority:    int32(r.Priority),
		Conditions:  string(r.Conditions),
		Actions:     string(r.Actions),
		Strategy:    r.Strategy,
		IsActive:    r.IsActive,
		CreatedAt:   r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   r.UpdatedAt.Format(time.RFC3339),
	}
}

// RoutingRulesToPb converts a slice of RoutingRule models.
func RoutingRulesToPb(rules []*RoutingRule) []*pb.RoutingRule {
	out := make([]*pb.RoutingRule, len(rules))
	for i, r := range rules {
		out[i] = RoutingRuleToPb(r)
	}
	return out
}

// SLAPolicyToPb converts an SLAPolicy model to its proto representation.
func SLAPolicyToPb(p *SLAPolicy) *pb.SLAPolicy {
	return &pb.SLAPolicy{
		Id:          p.ID,
		OrgId:       p.OrgID,
		WorkspaceId: p.WorkspaceID,
		Name:        p.Name,
		Conditions:  string(p.Conditions),
		Metrics:     string(p.Metrics),
		IsDefault:   p.IsDefault,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}

// SLAPoliciesToPb converts a slice of SLAPolicy models.
func SLAPoliciesToPb(policies []*SLAPolicy) []*pb.SLAPolicy {
	out := make([]*pb.SLAPolicy, len(policies))
	for i, p := range policies {
		out[i] = SLAPolicyToPb(p)
	}
	return out
}

// SLARecordToPb converts an SLARecord model to its proto representation.
func SLARecordToPb(r *SLARecord) *pb.SLARecord {
	rec := &pb.SLARecord{
		Id:                    r.ID,
		ConversationId:        r.ConversationID,
		PolicyId:              r.PolicyID,
		OrgId:                 r.OrgID,
		StartedAt:             r.StartedAt.Format(time.RFC3339),
		FirstResponseBreached: r.FirstResponseBreached,
		ResolutionBreached:    r.ResolutionBreached,
		CreatedAt:             r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:             r.UpdatedAt.Format(time.RFC3339),
	}

	if r.FirstResponseDeadline != nil {
		s := r.FirstResponseDeadline.Format(time.RFC3339)
		rec.FirstResponseDeadline = &s
	}
	if r.FirstRespondedAt != nil {
		s := r.FirstRespondedAt.Format(time.RFC3339)
		rec.FirstRespondedAt = &s
	}
	if r.ResolutionDeadline != nil {
		s := r.ResolutionDeadline.Format(time.RFC3339)
		rec.ResolutionDeadline = &s
	}
	if r.ResolvedAt != nil {
		s := r.ResolvedAt.Format(time.RFC3339)
		rec.ResolvedAt = &s
	}

	return rec
}
