package ticketing

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"kyla-be/internal/authctx"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/shared/events"

	"google.golang.org/grpc/status"
)

// AuthGateway is the subset of the auth stack TicketingServer needs.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
	ScopeCheck(ctx context.Context, scopeID string) (bool, *authctx.RequestMetadata, error)
}

// TicketingServer implements pb.TicketingServiceServer.
type TicketingServer struct {
	store    *TicketingStore
	auth     AuthGateway
	eventBus events.Publisher
	pb.UnimplementedTicketingServiceServer
}

// NewTicketingServer constructs a TicketingServer.
func NewTicketingServer(store *TicketingStore, auth AuthGateway, eventBus events.Publisher) *TicketingServer {
	return &TicketingServer{store: store, auth: auth, eventBus: eventBus}
}

// ── Ticket Room RPCs ──────────────────────────────────────────────────────────

func (s *TicketingServer) CreateTicketRoom(ctx context.Context, req *pb.CreateTicketRoomRequest) (*pb.TicketRoom, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	if req.GetTicketId() == "" || req.GetName() == "" {
		return nil, status.Error(400, "ticket_id and name are required")
	}
	room := &TicketRoom{
		TicketID: req.GetTicketId(),
		OrgID:    req.GetOrgId(),
		Name:     req.GetName(),
		Type:     roomTypeFromPb(req.GetType()),
	}
	created, err := s.store.CreateRoom(room)
	if err != nil {
		return nil, status.Error(500, "failed to create ticket room")
	}
	s.publishEvent(req.GetOrgId(), "room.created", created.ID, reqAuth.UserID.String(), map[string]string{"ticket_id": created.TicketID})
	return RoomToPb(created), nil
}

func (s *TicketingServer) GetTicketRoom(ctx context.Context, req *pb.GetTicketRoomRequest) (*pb.TicketRoom, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	room, err := s.store.FindRoomByID(req.GetId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(404, "ticket room not found")
	}
	return RoomToPb(room), nil
}

func (s *TicketingServer) ListTicketRooms(ctx context.Context, req *pb.ListTicketRoomsRequest) (*pb.ListTicketRoomsResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	rooms, err := s.store.ListRooms(req.GetTicketId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(500, "failed to list ticket rooms")
	}
	out := make([]*pb.TicketRoom, len(rooms))
	for i, r := range rooms {
		out[i] = RoomToPb(r)
	}
	return &pb.ListTicketRoomsResponse{Rooms: out}, nil
}

func (s *TicketingServer) AddRoomMessage(ctx context.Context, req *pb.AddRoomMessageRequest) (*pb.TicketRoomMessage, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	if req.GetContent() == "" {
		return nil, status.Error(400, "content is required")
	}
	msg := &TicketRoomMessage{
		RoomID:    req.GetRoomId(),
		OrgID:     req.GetOrgId(),
		AuthorID:  reqAuth.UserID.String(),
		Content:   req.GetContent(),
		IsPrivate: req.GetIsPrivate(),
	}
	created, err := s.store.AddMessage(msg)
	if err != nil {
		return nil, status.Error(500, "failed to add message")
	}
	s.publishEvent(req.GetOrgId(), "room.message_added", created.ID, reqAuth.UserID.String(), map[string]string{"room_id": created.RoomID})
	return MessageToPb(created), nil
}

func (s *TicketingServer) ListRoomMessages(ctx context.Context, req *pb.ListRoomMessagesRequest) (*pb.ListRoomMessagesResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	msgs, hasMore, err := s.store.ListMessages(req.GetRoomId(), req.GetOrgId(), req.GetBefore(), int(req.GetLimit()))
	if err != nil {
		return nil, status.Error(500, "failed to list messages")
	}
	out := make([]*pb.TicketRoomMessage, len(msgs))
	for i, m := range msgs {
		out[i] = MessageToPb(m)
	}
	return &pb.ListRoomMessagesResponse{Messages: out, HasMore: hasMore}, nil
}

// ── Macro RPCs ────────────────────────────────────────────────────────────────

func (s *TicketingServer) CreateMacro(ctx context.Context, req *pb.CreateMacroRequest) (*pb.Macro, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId(), req.GetWorkspaceId()); err != nil {
		return nil, err
	}
	m := &Macro{
		OrgID:       req.GetOrgId(),
		WorkspaceID: req.GetWorkspaceId(),
		Name:        req.GetName(),
		Content:     req.GetContent(),
		Actions:     req.GetActions(),
		Visibility:  macroVisibilityFromPb(req.GetVisibility()),
		CreatedBy:   reqAuth.UserID.String(),
	}
	if m.Actions == "" {
		m.Actions = "[]"
	}
	created, err := s.store.CreateMacro(m)
	if err != nil {
		return nil, status.Error(500, "failed to create macro")
	}
	s.publishEvent(req.GetOrgId(), "macro.created", created.ID, reqAuth.UserID.String(), map[string]string{"workspace_id": req.GetWorkspaceId()})
	return MacroToPb(created), nil
}

func (s *TicketingServer) GetMacro(ctx context.Context, req *pb.GetMacroRequest) (*pb.Macro, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	m, err := s.store.FindMacroByID(req.GetId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(404, "macro not found")
	}
	return MacroToPb(m), nil
}

func (s *TicketingServer) ListMacros(ctx context.Context, req *pb.ListMacrosRequest) (*pb.ListMacrosResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId(), req.GetWorkspaceId()); err != nil {
		return nil, err
	}
	macros, err := s.store.ListMacros(req.GetOrgId(), req.GetWorkspaceId())
	if err != nil {
		return nil, status.Error(500, "failed to list macros")
	}
	out := make([]*pb.Macro, len(macros))
	for i, m := range macros {
		out[i] = MacroToPb(m)
	}
	return &pb.ListMacrosResponse{Macros: out}, nil
}

func (s *TicketingServer) UpdateMacro(ctx context.Context, req *pb.UpdateMacroRequest) (*pb.Macro, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if req.GetName() != "" {
		updates["name"] = req.GetName()
	}
	if req.GetContent() != "" {
		updates["content"] = req.GetContent()
	}
	if req.GetActions() != "" {
		updates["actions"] = req.GetActions()
	}
	updates["visibility"] = macroVisibilityFromPb(req.GetVisibility())
	updated, err := s.store.UpdateMacro(req.GetId(), req.GetOrgId(), updates)
	if err != nil {
		return nil, status.Error(500, "failed to update macro")
	}
	s.publishEvent(req.GetOrgId(), "macro.updated", updated.ID, reqAuth.UserID.String(), nil)
	return MacroToPb(updated), nil
}

func (s *TicketingServer) DeleteMacro(ctx context.Context, req *pb.DeleteMacroRequest) (*pb.DeleteMacroResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}
	if err := s.store.DeleteMacro(req.GetId(), req.GetOrgId()); err != nil {
		return nil, status.Error(500, "failed to delete macro")
	}
	s.publishEvent(req.GetOrgId(), "macro.deleted", req.GetId(), reqAuth.UserID.String(), nil)
	return &pb.DeleteMacroResponse{Success: true}, nil
}

func (s *TicketingServer) ApplyMacro(ctx context.Context, req *pb.ApplyMacroRequest) (*pb.ApplyMacroResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if err := s.authorizeScope(ctx, req.GetOrgId()); err != nil {
		return nil, err
	}

	macro, err := s.store.FindMacroByID(req.GetMacroId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(404, "macro not found")
	}

	// Parse actions JSON and apply field patches to the ticket object
	var actions []map[string]interface{}
	if err := json.Unmarshal([]byte(macro.Actions), &actions); err == nil && len(actions) > 0 {
		if err := s.applyFieldPatches(req.GetTicketId(), req.GetOrgId(), actions); err != nil {
			return nil, status.Error(500, fmt.Sprintf("apply macro patches: %v", err))
		}
	}

	// Post macro content into the specified room (if room_id provided)
	if req.GetRoomId() != "" && macro.Content != "" {
		msg := &TicketRoomMessage{
			RoomID:    req.GetRoomId(),
			OrgID:     req.GetOrgId(),
			AuthorID:  reqAuth.UserID.String(),
			Content:   macro.Content,
			IsPrivate: false,
		}
		if _, err := s.store.AddMessage(msg); err != nil {
			return nil, status.Error(500, "failed to post macro message")
		}
	}
	s.publishEvent(req.GetOrgId(), "macro.applied", req.GetTicketId(), reqAuth.UserID.String(), map[string]string{"macro_id": req.GetMacroId()})

	return &pb.ApplyMacroResponse{Success: true, TicketId: req.GetTicketId()}, nil
}

// applyFieldPatches applies field-patch actions to a ticket object via raw SQL.
func (s *TicketingServer) applyFieldPatches(ticketID, orgID string, actions []map[string]interface{}) error {
	existing, err := s.store.GetTicketData(ticketID, orgID)
	if err != nil {
		return err
	}
	for _, action := range actions {
		if field, ok := action["field"].(string); ok {
			existing[field] = action["value"]
		}
	}
	return s.store.PatchTicketData(ticketID, orgID, existing)
}

func (s *TicketingServer) authorizeScope(ctx context.Context, scopeIDs ...string) error {
	for _, scopeID := range scopeIDs {
		if scopeID == "" {
			continue
		}
		ok, _, err := s.auth.ScopeCheck(ctx, scopeID)
		if err != nil || !ok {
			return status.Error(403, "Forbidden")
		}
	}
	return nil
}

func (s *TicketingServer) publishEvent(orgID, action, entityID, actorID string, payload interface{}) {
	ev, err := events.NewEvent(orgID, "", "ticket", action, entityID, actorID, payload)
	if err != nil {
		return
	}
	ev.Subject = fmt.Sprintf("kyla.%s.ticket.%s", orgID, action)
	if err := s.eventBus.Publish(ev); err != nil {
		log.Printf("[ticketing] event publish error (%s): %v", action, err)
	}
}
