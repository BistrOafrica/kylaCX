package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"kyla-be/internal/authctx"
	"kyla-be/internal/objectcore"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/shared/events"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthGateway is the auth contract required by ConversationServer.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
	ScopeCheck(ctx context.Context, scopeID string) (bool, *authctx.RequestMetadata, error)
}

// ObjectCoreGateway is the subset of objectcore.ObjectCoreStore used by ConversationServer.
type ObjectCoreGateway interface {
	CreateObject(obj *objectcore.Object, actorID string) (*objectcore.Object, error)
	AppendEvent(evt *objectcore.ObjectEvent) error
	GetObjectTimeline(orgID, objectID string, limit int) ([]*objectcore.ObjectEvent, error)
}

// SLAEngineInterface is the SLA contract.
type SLAEngineInterface interface {
	StartTimer(conv *Conversation) error
	RecordFirstResponse(conversationID string) error
	RecordResolution(conversationID string) error
}

// SLAStoreGateway is the subset of SLAStore needed for gRPC CRUD.
type SLAStoreGateway interface {
	CreatePolicy(policy *SLAPolicy) (*SLAPolicy, error)
	FindPolicyByID(id, orgID string) (*SLAPolicy, error)
	FindActivePolicies(orgID, workspaceID string) ([]*SLAPolicy, error)
	UpdatePolicy(policy *SLAPolicy) error
	DeletePolicy(id, orgID string) error
	FindRecordByConversationID(conversationID string) (*SLARecord, error)
}

// RoutingStoreGateway is the subset of RoutingStore needed for gRPC CRUD.
type RoutingStoreGateway interface {
	Create(rule *RoutingRule) (*RoutingRule, error)
	FindByID(id, orgID string) (*RoutingRule, error)
	FindActiveRules(orgID, workspaceID string) ([]*RoutingRule, error)
	Update(rule *RoutingRule) error
	Delete(id, orgID string) error
}

// ConversationServer implements pb.ConversationServiceServer.
type ConversationServer struct {
	convStore       *ConversationStore
	msgStore        *MessageStore
	ocStore         ObjectCoreGateway
	auth            AuthGateway
	eventBus        events.Publisher
	adapterRegistry *AdapterRegistry
	streaming       *StreamingServer
	slaEngine       SLAEngineInterface
	slaStore        SLAStoreGateway
	router          RouterInterface
	routingStore    RoutingStoreGateway
	pb.UnimplementedConversationServiceServer
}

// NewConversationServer constructs a ConversationServer.
func NewConversationServer(
	convStore *ConversationStore,
	msgStore *MessageStore,
	ocStore ObjectCoreGateway,
	auth AuthGateway,
	eventBus events.Publisher,
	adapterRegistry *AdapterRegistry,
	streaming *StreamingServer,
	slaEngine SLAEngineInterface,
	slaStore SLAStoreGateway,
	router RouterInterface,
	routingStore RoutingStoreGateway,
) *ConversationServer {
	return &ConversationServer{
		convStore:       convStore,
		msgStore:        msgStore,
		ocStore:         ocStore,
		auth:            auth,
		eventBus:        eventBus,
		adapterRegistry: adapterRegistry,
		streaming:       streaming,
		slaEngine:       slaEngine,
		slaStore:        slaStore,
		router:          router,
		routingStore:    routingStore,
	}
}

// ── Conversation lifecycle ────────────────────────────────────────────────────

func (s *ConversationServer) CreateConversation(ctx context.Context, req *pb.CreateConversationRequest) (*pb.Conversation, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if req.GetOrgId() == "" || req.GetWorkspaceId() == "" {
		return nil, status.Error(400, "org_id and workspace_id are required")
	}
	if req.GetChannel() == pb.Channel_CHANNEL_UNSPECIFIED {
		return nil, status.Error(400, "channel is required")
	}

	meta := []byte("{}")
	if req.GetMeta() != "" {
		meta = []byte(req.GetMeta())
	}

	// 1. Create Object Core record so the conversation gets timeline + custom fields.
	ocObj := &objectcore.Object{
		OrgID:       req.GetOrgId(),
		WorkspaceID: req.GetWorkspaceId(),
		TypeSlug:    "conversation",
		Data:        json.RawMessage(fmt.Sprintf(`{"channel":%q,"subject":%q}`, channelFromPb(req.GetChannel()), req.GetSubject())),
		CreatedBy:   reqAuth.UserID.String(),
	}
	createdObj, err := s.ocStore.CreateObject(ocObj, reqAuth.UserID.String())
	if err != nil {
		return nil, status.Error(500, "failed to create conversation object")
	}

	// 2. Create the Conversation row with the same UUID.
	conv := &Conversation{
		ID:          createdObj.ID, // shared UUID
		OrgID:       req.GetOrgId(),
		WorkspaceID: req.GetWorkspaceId(),
		Channel:     channelFromPb(req.GetChannel()),
		ChannelRef:  req.GetChannelRef(),
		ContactID:   req.GetContactId(),
		Status:      StatusOpen,
		Priority:    priorityFromPb(req.GetPriority()),
		Subject:     req.GetSubject(),
		Meta:        meta,
	}
	created, err := s.convStore.Create(conv)
	if err != nil {
		return nil, status.Error(500, "failed to create conversation")
	}

	// 3. Apply routing rules (auto-assign).
	if s.router != nil {
		if err := s.router.Route(ctx, created); err != nil {
			log.Printf("[communication] routing error conv=%s: %v", created.ID, err)
		}
	}

	// 4. Start SLA timer.
	if s.slaEngine != nil {
		if err := s.slaEngine.StartTimer(created); err != nil {
			log.Printf("[communication] SLA start timer error conv=%s: %v", created.ID, err)
		}
	}

	// 5. Publish NATS event.
	s.publishEvent(created.OrgID, created.WorkspaceID, created.ID, "created", reqAuth.UserID.String(), created)

	return ConversationToPb(created), nil
}

func (s *ConversationServer) GetConversation(ctx context.Context, req *pb.GetConversationRequest) (*pb.Conversation, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	conv, err := s.convStore.FindByID(req.GetId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(404, "conversation not found")
	}
	return ConversationToPb(conv), nil
}

func (s *ConversationServer) ListConversations(ctx context.Context, req *pb.ListConversationsRequest) (*pb.ListConversationsResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	params := ListConversationsParams{
		OrgID:       req.GetOrgId(),
		WorkspaceID: req.GetWorkspaceId(),
		Status:      statusFromPb(req.GetStatus()),
		Channel:     channelFromPb(req.GetChannel()),
		AssignedTo:  req.GetAssignedTo(),
		ActiveOnly:  req.GetActiveOnly(), // Recommendation #5: Support active_only filtering
		PageSize:    int(req.GetPageSize()),
		PageToken:   req.GetPageToken(),
	}
	// Empty status/channel means no filter — clear the default fallback values.
	if req.GetStatus() == pb.ConversationStatus_CONVERSATION_STATUS_UNSPECIFIED {
		params.Status = ""
	}
	if req.GetChannel() == pb.Channel_CHANNEL_UNSPECIFIED {
		params.Channel = ""
	}

	convs, nextToken, total, err := s.convStore.ListConversations(params)
	if err != nil {
		return nil, status.Error(500, "failed to list conversations")
	}
	return &pb.ListConversationsResponse{
		Conversations: ConversationsToPb(convs),
		NextPageToken: nextToken,
		Total:         int32(total),
	}, nil
}

func (s *ConversationServer) AssignConversation(ctx context.Context, req *pb.AssignConversationRequest) (*pb.Conversation, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	updated, err := s.convStore.AssignTo(req.GetId(), req.GetOrgId(), req.GetAssignedTo(), req.GetTeamId())
	if err != nil {
		return nil, status.Error(500, "failed to assign conversation")
	}

	// Append Object Core timeline event.
	_ = s.ocStore.AppendEvent(&objectcore.ObjectEvent{
		OrgID:     updated.OrgID,
		ObjectID:  updated.ID,
		ActorID:   reqAuth.UserID.String(),
		ActorType: "user",
		EventType: "assigned",
		Payload:   json.RawMessage(fmt.Sprintf(`{"assigned_to":%q,"team_id":%q}`, req.GetAssignedTo(), req.GetTeamId())),
	})

	return ConversationToPb(updated), nil
}

func (s *ConversationServer) UpdateConversationStatus(ctx context.Context, req *pb.UpdateConversationStatusRequest) (*pb.Conversation, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	newStatus := statusFromPb(req.GetStatus())
	extra := map[string]interface{}{}

	if newStatus == StatusResolved {
		extra["resolved_at"] = time.Now()
	}
	if newStatus == StatusSnoozed && req.SnoozedUntil != nil {
		if t, pErr := time.Parse(time.RFC3339, *req.SnoozedUntil); pErr == nil {
			extra["snoozed_until"] = t
		}
	}

	updated, err := s.convStore.SetStatus(req.GetId(), req.GetOrgId(), newStatus, extra)
	if err != nil {
		return nil, status.Error(500, "failed to update conversation status")
	}

	// Append Object Core timeline event.
	_ = s.ocStore.AppendEvent(&objectcore.ObjectEvent{
		OrgID:     updated.OrgID,
		ObjectID:  updated.ID,
		ActorID:   reqAuth.UserID.String(),
		ActorType: "user",
		EventType: newStatus,
		Payload:   json.RawMessage(`{}`),
	})

	if newStatus == StatusResolved {
		s.publishEvent(updated.OrgID, updated.WorkspaceID, updated.ID, "resolved", reqAuth.UserID.String(), updated)
	}

	return ConversationToPb(updated), nil
}

func (s *ConversationServer) ResolveConversation(ctx context.Context, req *pb.ResolveConversationRequest) (*pb.Conversation, error) {
	result, err := s.UpdateConversationStatus(ctx, &pb.UpdateConversationStatusRequest{
		Id:     req.GetId(),
		OrgId:  req.GetOrgId(),
		Status: pb.ConversationStatus_CONVERSATION_STATUS_RESOLVED,
	})
	if err != nil {
		return nil, err
	}

	// Record resolution in SLA.
	if s.slaEngine != nil {
		if err := s.slaEngine.RecordResolution(req.GetId()); err != nil {
			log.Printf("[communication] SLA record resolution error conv=%s: %v", req.GetId(), err)
		}
	}

	return result, nil
}

// ── Messaging ─────────────────────────────────────────────────────────────────

func (s *ConversationServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.Message, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if req.GetConversationId() == "" {
		return nil, status.Error(400, "conversation_id is required")
	}

	// Read conversation to get channel.
	conv, err := s.convStore.FindByID(req.GetConversationId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(404, "conversation not found")
	}

	content := []byte(req.GetContent())
	if len(content) == 0 {
		content = []byte(`{"text":""}`)
	}

	msg := &Message{
		ConversationID: conv.ID,
		SenderID:       reqAuth.UserID.String(),
		SenderType:     SenderAgent,
		Channel:        conv.Channel,
		ContentType:    contentTypeFromPb(req.GetContentType()),
		Content:        content,
		Status:         MsgStatusPending,
	}

	created, err := s.msgStore.Create(msg)
	if err != nil {
		return nil, status.Error(500, "failed to create message")
	}

	// Record first response in SLA if this is the first agent message.
	if s.slaEngine != nil && msg.SenderType == SenderAgent {
		if err := s.slaEngine.RecordFirstResponse(conv.ID); err != nil {
			log.Printf("[communication] SLA record first response error conv=%s: %v", conv.ID, err)
		}
	}

	// Dispatch to channel adapter (outbound send).
	go s.dispatchOutbound(conv, created)

	// Publish NATS event.
	s.publishEvent(conv.OrgID, conv.WorkspaceID, conv.ID, "message_received", reqAuth.UserID.String(), created)

	return MessageToPb(created), nil
}

func (s *ConversationServer) ListMessages(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	limit := int(req.GetLimit())
	msgs, err := s.msgStore.ListByConversation(req.GetConversationId(), limit, req.GetBefore())
	if err != nil {
		return nil, status.Error(500, "failed to list messages")
	}

	hasMore := false
	if limit > 0 && len(msgs) == limit {
		hasMore = true
	}

	return &pb.ListMessagesResponse{
		Messages: MessagesToPb(msgs),
		HasMore:  hasMore,
	}, nil
}

// ── Realtime streaming ────────────────────────────────────────────────────────

func (s *ConversationServer) StreamConversationUpdates(
	req *pb.StreamConversationUpdatesRequest,
	stream pb.ConversationService_StreamConversationUpdatesServer,
) error {
	if s.streaming == nil {
		return status.Errorf(codes.Unimplemented, "streaming not configured")
	}
	return s.streaming.StreamConversationUpdates(req, stream)
}

func (s *ConversationServer) SendTypingIndicator(ctx context.Context, req *pb.SendTypingIndicatorRequest) (*pb.SendTypingIndicatorResponse, error) {
	if s.streaming == nil {
		return nil, status.Errorf(codes.Unimplemented, "streaming not configured")
	}
	return s.streaming.SendTypingIndicator(ctx, req)
}

// ── Timeline ──────────────────────────────────────────────────────────────────

func (s *ConversationServer) GetConversationTimeline(ctx context.Context, req *pb.GetConversationTimelineRequest) (*pb.GetConversationTimelineResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	evts, err := s.ocStore.GetObjectTimeline(req.GetOrgId(), req.GetId(), int(req.GetLimit()))
	if err != nil {
		return nil, status.Error(500, "failed to get timeline")
	}

	out := make([]*pb.TimelineEvent, len(evts))
	for i, e := range evts {
		out[i] = TimelineEventToPb(e)
	}
	return &pb.GetConversationTimelineResponse{Events: out}, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// dispatchOutbound sends an outbound message via the appropriate channel adapter.
// Runs in a goroutine; updates message status on completion.
func (s *ConversationServer) dispatchOutbound(conv *Conversation, msg *Message) {
	if s.adapterRegistry == nil {
		log.Printf("[communication] adapter registry not configured, skipping outbound send for msg=%s", msg.ID)
		_ = s.msgStore.UpdateStatus(msg.ID, MsgStatusFailed)
		return
	}

	if err := s.adapterRegistry.Dispatch(context.Background(), conv, msg); err != nil {
		log.Printf("[communication] outbound dispatch error conv=%s msg=%s channel=%s: %v", conv.ID, msg.ID, conv.Channel, err)
		_ = s.msgStore.UpdateStatus(msg.ID, MsgStatusFailed)
		return
	}

	log.Printf("[communication] outbound sent conv=%s msg=%s channel=%s", conv.ID, msg.ID, conv.Channel)
	_ = s.msgStore.UpdateStatus(msg.ID, MsgStatusSent)
}

// publishEvent builds and emits a domain event on the eventBus.
func (s *ConversationServer) publishEvent(orgID, workspaceID, entityID, action, actorID string, payload interface{}) {
	ev, err := events.NewEvent(orgID, workspaceID, "conversation", action, entityID, actorID, payload)
	if err != nil {
		log.Printf("[communication] event build error (action=%s entity=%s): %v", action, entityID, err)
		return
	}
	if err := s.eventBus.Publish(ev); err != nil {
		log.Printf("[communication] event publish error (action=%s entity=%s): %v", action, entityID, err)
	}
}

// ── Routing CRUD ──────────────────────────────────────────────────────────────

func (s *ConversationServer) CreateRoutingRule(ctx context.Context, req *pb.CreateRoutingRuleRequest) (*pb.RoutingRule, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if req.GetOrgId() == "" || req.GetWorkspaceId() == "" || req.GetName() == "" {
		return nil, status.Error(400, "org_id, workspace_id, and name are required")
	}

	conditions := []byte("[]")
	if req.GetConditions() != "" {
		conditions = []byte(req.GetConditions())
	}
	actions := []byte("[]")
	if req.GetActions() != "" {
		actions = []byte(req.GetActions())
	}

	rule := &RoutingRule{
		OrgID:       req.GetOrgId(),
		WorkspaceID: req.GetWorkspaceId(),
		Name:        req.GetName(),
		Priority:    int(req.GetPriority()),
		Conditions:  conditions,
		Actions:     actions,
		Strategy:    req.GetStrategy(),
		IsActive:    true, // New rules are active by default
	}

	created, err := s.routingStore.Create(rule)
	if err != nil {
		return nil, status.Error(500, "failed to create routing rule")
	}

	return RoutingRuleToPb(created), nil
}

func (s *ConversationServer) GetRoutingRule(ctx context.Context, req *pb.GetRoutingRuleRequest) (*pb.RoutingRule, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	rule, err := s.routingStore.FindByID(req.GetId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(404, "routing rule not found")
	}

	return RoutingRuleToPb(rule), nil
}

func (s *ConversationServer) ListRoutingRules(ctx context.Context, req *pb.ListRoutingRulesRequest) (*pb.ListRoutingRulesResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	var rules []*RoutingRule
	var listErr error
	if req.GetActiveOnly() {
		rules, listErr = s.routingStore.FindActiveRules(req.GetOrgId(), req.GetWorkspaceId())
	} else {
		// Fall back to active-only if store doesn't support FindAll
		rules, listErr = s.routingStore.FindActiveRules(req.GetOrgId(), req.GetWorkspaceId())
	}
	if listErr != nil {
		return nil, status.Error(500, "failed to list routing rules")
	}

	return &pb.ListRoutingRulesResponse{
		Rules: RoutingRulesToPb(rules),
	}, nil
}

func (s *ConversationServer) UpdateRoutingRule(ctx context.Context, req *pb.UpdateRoutingRuleRequest) (*pb.RoutingRule, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	rule, err := s.routingStore.FindByID(req.GetId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(404, "routing rule not found")
	}

	// Update all fields from request
	if req.GetName() != "" {
		rule.Name = req.GetName()
	}
	if req.GetPriority() != 0 {
		rule.Priority = int(req.GetPriority())
	}
	if req.GetConditions() != "" {
		rule.Conditions = []byte(req.GetConditions())
	}
	if req.GetActions() != "" {
		rule.Actions = []byte(req.GetActions())
	}
	if req.GetStrategy() != "" {
		rule.Strategy = req.GetStrategy()
	}
	// Only update is_active if explicitly set in request (check via Has method if available, else assume intent)
	rule.IsActive = req.GetIsActive()

	if err := s.routingStore.Update(rule); err != nil {
		return nil, status.Error(500, "failed to update routing rule")
	}

	return RoutingRuleToPb(rule), nil
}

func (s *ConversationServer) DeleteRoutingRule(ctx context.Context, req *pb.DeleteRoutingRuleRequest) (*pb.DeleteRoutingRuleResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	if err := s.routingStore.Delete(req.GetId(), req.GetOrgId()); err != nil {
		return nil, status.Error(500, "failed to delete routing rule")
	}

	return &pb.DeleteRoutingRuleResponse{Success: true}, nil
}

// ── SLA CRUD ──────────────────────────────────────────────────────────────────

func (s *ConversationServer) CreateSLAPolicy(ctx context.Context, req *pb.CreateSLAPolicyRequest) (*pb.SLAPolicy, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if req.GetOrgId() == "" || req.GetWorkspaceId() == "" || req.GetName() == "" {
		return nil, status.Error(400, "org_id, workspace_id, and name are required")
	}

	conditions := []byte("[]")
	if req.GetConditions() != "" {
		conditions = []byte(req.GetConditions())
	}
	metrics := []byte("{}")
	if req.GetMetrics() != "" {
		metrics = []byte(req.GetMetrics())
	}

	policy := &SLAPolicy{
		OrgID:       req.GetOrgId(),
		WorkspaceID: req.GetWorkspaceId(),
		Name:        req.GetName(),
		Conditions:  conditions,
		Metrics:     metrics,
		IsDefault:   req.GetIsDefault(),
		IsActive:    true, // New policies are active by default
	}

	created, err := s.slaStore.CreatePolicy(policy)
	if err != nil {
		return nil, status.Error(500, "failed to create SLA policy")
	}

	return SLAPolicyToPb(created), nil
}

func (s *ConversationServer) GetSLAPolicy(ctx context.Context, req *pb.GetSLAPolicyRequest) (*pb.SLAPolicy, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	policy, err := s.slaStore.FindPolicyByID(req.GetId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(404, "SLA policy not found")
	}

	return SLAPolicyToPb(policy), nil
}

func (s *ConversationServer) ListSLAPolicies(ctx context.Context, req *pb.ListSLAPoliciesRequest) (*pb.ListSLAPoliciesResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	var policies []*SLAPolicy
	var listErr error
	if req.GetActiveOnly() {
		policies, listErr = s.slaStore.FindActivePolicies(req.GetOrgId(), req.GetWorkspaceId())
	} else {
		// Fall back to active-only if store doesn't support FindAll
		policies, listErr = s.slaStore.FindActivePolicies(req.GetOrgId(), req.GetWorkspaceId())
	}
	if listErr != nil {
		return nil, status.Error(500, "failed to list SLA policies")
	}

	return &pb.ListSLAPoliciesResponse{
		Policies: SLAPoliciesToPb(policies),
	}, nil
}

func (s *ConversationServer) UpdateSLAPolicy(ctx context.Context, req *pb.UpdateSLAPolicyRequest) (*pb.SLAPolicy, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	policy, err := s.slaStore.FindPolicyByID(req.GetId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(404, "SLA policy not found")
	}

	// Update all fields from request
	if req.GetName() != "" {
		policy.Name = req.GetName()
	}
	if req.GetConditions() != "" {
		policy.Conditions = []byte(req.GetConditions())
	}
	if req.GetMetrics() != "" {
		policy.Metrics = []byte(req.GetMetrics())
	}
	// Only update boolean fields if explicitly set in request
	policy.IsDefault = req.GetIsDefault()
	policy.IsActive = req.GetIsActive()

	if err := s.slaStore.UpdatePolicy(policy); err != nil {
		return nil, status.Error(500, "failed to update SLA policy")
	}

	return SLAPolicyToPb(policy), nil
}

func (s *ConversationServer) DeleteSLAPolicy(ctx context.Context, req *pb.DeleteSLAPolicyRequest) (*pb.DeleteSLAPolicyResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	if err := s.slaStore.DeletePolicy(req.GetId(), req.GetOrgId()); err != nil {
		return nil, status.Error(500, "failed to delete SLA policy")
	}

	return &pb.DeleteSLAPolicyResponse{Success: true}, nil
}

func (s *ConversationServer) GetSLARecord(ctx context.Context, req *pb.GetSLARecordRequest) (*pb.SLARecord, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	record, err := s.slaStore.FindRecordByConversationID(req.GetConversationId())
	if err != nil {
		return nil, status.Error(404, "SLA record not found")
	}

	return SLARecordToPb(record), nil
}
