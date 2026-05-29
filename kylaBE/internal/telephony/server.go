package telephony

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"kyla-be/internal/authctx"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/shared/events"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthGateway is the subset of the auth stack the telephony server needs.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
}

// TokenIssuer issues short-lived JWTs for browser softphones to authenticate
// SIP-over-WSS. Defined as an interface so main.go can inject any signer
// without the server caring about the underlying secret.
type TokenIssuer interface {
	IssueSoftphoneToken(orgID, userID, extension string, ttl time.Duration) (string, error)
}

// Server implements pb.TelephonyServiceServer.
type Server struct {
	store     *Store
	auth      AuthGateway
	pbx       PBXController
	publisher events.Publisher
	issuer    TokenIssuer
	wsURL     string
	sipRealm  string
	turnURL   string
	turnUser  string
	turnPass  string
	pb.UnimplementedTelephonyServiceServer
}

// ServerConfig groups the static runtime knobs the server reads at construction
// time. Avoids a long constructor signature.
type ServerConfig struct {
	WssURL       string
	SipRealm     string
	TurnURL      string
	TurnUsername string
	TurnPassword string
}

func NewServer(store *Store, auth AuthGateway, pbx PBXController, publisher events.Publisher, issuer TokenIssuer, cfg ServerConfig) *Server {
	if pbx == nil {
		pbx = NoopPBX{}
	}
	return &Server{
		store:     store,
		auth:      auth,
		pbx:       pbx,
		publisher: publisher,
		issuer:    issuer,
		wsURL:     cfg.WssURL,
		sipRealm:  cfg.SipRealm,
		turnURL:   cfg.TurnURL,
		turnUser:  cfg.TurnUsername,
		turnPass:  cfg.TurnPassword,
	}
}

// ── Call control ─────────────────────────────────────────────────────────────

func (s *Server) OriginateCall(ctx context.Context, req *pb.OriginateCallRequest) (*pb.OriginateCallResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !s.pbx.Enabled() {
		return nil, status.Error(codes.FailedPrecondition, "telephony PBX not configured")
	}
	if req.GetToNumber() == "" {
		return nil, status.Error(codes.InvalidArgument, "to_number is required")
	}
	agentID := req.GetAgentId()
	if agentID == "" {
		agentID = md.UserID.String()
	}
	ext, err := s.store.GetExtensionByUserID(agentID)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "no SIP extension registered for agent")
	}

	// Resolve outbound trunk. Either explicit (req.trunk_id) or the org's
	// first active trunk.
	var trunk *SipTrunk
	if req.GetTrunkId() != "" {
		trunk, err = s.store.GetTrunk(req.GetTrunkId(), md.OrganisationID.String())
		if err != nil {
			return nil, status.Error(codes.NotFound, "trunk not found")
		}
	} else {
		trunks, _ := s.store.ListTrunks(md.OrganisationID.String())
		for _, t := range trunks {
			if t.IsActive {
				trunk = t
				break
			}
		}
		if trunk == nil {
			return nil, status.Error(codes.FailedPrecondition, "no active SIP trunk configured")
		}
	}

	callUUID, err := s.pbx.Originate(ctx, OriginateRequest{
		AgentID:          agentID,
		AgentExtension:   ext.Extension,
		ToNumber:         req.GetToNumber(),
		FromNumber:       trunk.FromURI,
		TrunkGateway:     trunk.GatewayName,
		OrgID:            md.OrganisationID.String(),
		WorkspaceID:      req.GetWorkspaceId(),
		ContactID:        req.GetContactId(),
		RecordingEnabled: req.GetRecordingEnabled(),
	})
	if err != nil {
		log.Printf("[telephony] originate failed: %v", err)
		return nil, status.Error(codes.Internal, "failed to start call")
	}

	call := &Call{
		ID:               callUUID,
		OrgID:            md.OrganisationID.String(),
		WorkspaceID:      req.GetWorkspaceId(),
		Direction:        string(DirectionOutbound),
		Status:           string(StatusRinging),
		FromNumber:       trunk.FromURI,
		ToNumber:         req.GetToNumber(),
		AgentID:          agentID,
		ContactID:        req.GetContactId(),
		TrunkID:          trunk.ID,
		RecordingEnabled: req.GetRecordingEnabled(),
		StartedAt:        time.Now().UTC(),
	}
	created, err := s.store.CreateCall(call)
	if err != nil {
		log.Printf("[telephony] persist call %s: %v", callUUID, err)
		// Don't fail the RPC — the call is live in the PBX; we just lost the
		// projection. Operator can recover via the ESL event stream.
	} else {
		s.publishCallEvent(created, "call.started")
	}
	return &pb.OriginateCallResponse{Session: CallToPb(call)}, nil
}

func (s *Server) HangupCall(ctx context.Context, req *pb.HangupRequest) (*pb.HangupResponse, error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	if err := s.pbx.Hangup(ctx, req.GetId(), req.GetReason()); err != nil {
		log.Printf("[telephony] hangup %s: %v", req.GetId(), err)
		return nil, status.Error(codes.Internal, "hangup failed")
	}
	return &pb.HangupResponse{Id: req.GetId(), Status: "hangup_requested"}, nil
}

func (s *Server) TransferCall(ctx context.Context, req *pb.TransferRequest) (*pb.TransferResponse, error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	target := req.GetTargetExtension()
	if target == "" {
		target = req.GetTargetNumber()
	}
	if target == "" {
		return nil, status.Error(codes.InvalidArgument, "target_extension or target_number is required")
	}
	if err := s.pbx.Transfer(ctx, req.GetId(), target, req.GetBlind()); err != nil {
		log.Printf("[telephony] transfer %s: %v", req.GetId(), err)
		return nil, status.Error(codes.Internal, "transfer failed")
	}
	return &pb.TransferResponse{Id: req.GetId(), Status: "transfer_requested"}, nil
}

func (s *Server) HoldCall(ctx context.Context, req *pb.HoldRequest) (*pb.HoldResponse, error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	if err := s.pbx.Hold(ctx, req.GetId()); err != nil {
		log.Printf("[telephony] hold %s: %v", req.GetId(), err)
		return nil, status.Error(codes.Internal, "hold failed")
	}
	return &pb.HoldResponse{Id: req.GetId(), Status: "held"}, nil
}

func (s *Server) ResumeCall(ctx context.Context, req *pb.ResumeRequest) (*pb.ResumeResponse, error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	if err := s.pbx.Resume(ctx, req.GetId()); err != nil {
		log.Printf("[telephony] resume %s: %v", req.GetId(), err)
		return nil, status.Error(codes.Internal, "resume failed")
	}
	return &pb.ResumeResponse{Id: req.GetId(), Status: "resumed"}, nil
}

// ── Call history ─────────────────────────────────────────────────────────────

func (s *Server) GetCallSession(ctx context.Context, req *pb.GetCallRequest) (*pb.Call, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	c, err := s.store.GetCall(req.GetId(), md.OrganisationID.String())
	if err != nil {
		return nil, status.Error(codes.NotFound, "call not found")
	}
	return CallToPb(c), nil
}

func (s *Server) ListCallSessions(ctx context.Context, req *pb.ListCallsRequest) (*pb.ListCallsResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	rows, total, err := s.store.ListCalls(ListCallsParams{
		OrgID:       md.OrganisationID.String(),
		WorkspaceID: req.GetWorkspaceId(),
		Direction:   req.GetDirection(),
		Status:      req.GetStatus(),
		AgentID:     req.GetAgentId(),
		PageSize:    int(req.GetPageSize()),
		PageToken:   req.GetPageToken(),
	})
	if err != nil {
		log.Printf("[telephony] list calls: %v", err)
		return nil, status.Error(codes.Internal, "failed to list calls")
	}
	out := make([]*pb.Call, 0, len(rows))
	for _, r := range rows {
		out = append(out, CallToPb(r))
	}
	return &pb.ListCallsResponse{Sessions: out, Total: total}, nil
}

func (s *Server) AppendCallLog(ctx context.Context, req *pb.AppendCallEventRequest) (*pb.AppendCallEventResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	// Trust the caller's call_session_id but scope by org via the call row.
	if _, err := s.store.GetCall(req.GetCallSessionId(), md.OrganisationID.String()); err != nil {
		return nil, status.Error(codes.NotFound, "call not found")
	}
	detail := json.RawMessage(req.GetDetail())
	if len(detail) == 0 {
		detail = json.RawMessage(`{}`)
	}
	evt, err := s.store.AppendEvent(&CallEvent{
		CallID:    req.GetCallSessionId(),
		OrgID:     md.OrganisationID.String(),
		EventType: req.GetEventType(),
		Detail:    detail,
		At:        time.Now().UTC(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to append event")
	}
	return &pb.AppendCallEventResponse{Log: CallEventToPb(evt)}, nil
}

func (s *Server) ListCallLogs(ctx context.Context, req *pb.ListCallEventsRequest) (*pb.ListCallEventsResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.store.GetCall(req.GetCallSessionId(), md.OrganisationID.String()); err != nil {
		return nil, status.Error(codes.NotFound, "call not found")
	}
	rows, err := s.store.ListEvents(req.GetCallSessionId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list events")
	}
	out := make([]*pb.CallEvent, 0, len(rows))
	for _, r := range rows {
		out = append(out, CallEventToPb(r))
	}
	return &pb.ListCallEventsResponse{Logs: out}, nil
}

// ── SIP extensions ───────────────────────────────────────────────────────────

func (s *Server) CreateSipExtension(ctx context.Context, req *pb.CreateSipExtensionRequest) (*pb.SipExtension, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	in := ExtensionFromPb(req.GetExtension())
	if in == nil || in.Extension == "" || in.UserID == "" {
		return nil, status.Error(codes.InvalidArgument, "extension and user_id are required")
	}
	in.OrgID = md.OrganisationID.String()

	// Generate a SIP password and persist its bcrypt hash. The plaintext is
	// pushed to the PBX via ProvisionExtension and never returned over gRPC
	// — operators retrieve it out-of-band from the PBX config the first time.
	plaintext := newSipPassword()
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "hash password")
	}
	in.PasswordHash = string(hash)

	created, err := s.store.CreateExtension(in)
	if err != nil {
		log.Printf("[telephony] create extension: %v", err)
		return nil, status.Error(codes.Internal, "failed to create extension")
	}
	// Best-effort PBX provisioning — failure here doesn't fail the gRPC call
	// since operators can re-provision later.
	if err := s.pbx.ProvisionExtension(ctx, *created, plaintext); err != nil {
		log.Printf("[telephony] provision extension on PBX: %v", err)
	}
	return ExtensionToPb(created), nil
}

func (s *Server) GetSipExtension(ctx context.Context, req *pb.GetSipExtensionRequest) (*pb.SipExtension, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	e, err := s.store.GetExtension(req.GetId(), md.OrganisationID.String())
	if err != nil {
		return nil, status.Error(codes.NotFound, "extension not found")
	}
	return ExtensionToPb(e), nil
}

func (s *Server) ListSipExtensions(ctx context.Context, req *pb.ListSipExtensionsRequest) (*pb.ListSipExtensionsResponse, error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	rows, err := s.store.ListExtensions(req.GetWorkspaceId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list extensions")
	}
	out := make([]*pb.SipExtension, 0, len(rows))
	for _, r := range rows {
		out = append(out, ExtensionToPb(r))
	}
	return &pb.ListSipExtensionsResponse{Extensions: out}, nil
}

func (s *Server) DeleteSipExtension(ctx context.Context, req *pb.DeleteSipExtensionRequest) (*pb.DeleteSipExtensionResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.store.DeleteExtension(req.GetId(), md.OrganisationID.String()); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete extension")
	}
	return &pb.DeleteSipExtensionResponse{Ok: true}, nil
}

// ── SIP trunks ──────────────────────────────────────────────────────────────

func (s *Server) CreateSipTrunk(ctx context.Context, req *pb.CreateSipTrunkRequest) (*pb.SipTrunk, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	in := TrunkFromPb(req.GetTrunk())
	if in == nil || in.Name == "" || in.GatewayName == "" {
		return nil, status.Error(codes.InvalidArgument, "name and gateway_name are required")
	}
	in.OrgID = md.OrganisationID.String()
	created, err := s.store.CreateTrunk(in)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create trunk")
	}
	if err := s.pbx.ProvisionTrunk(ctx, *created); err != nil {
		log.Printf("[telephony] provision trunk on PBX: %v", err)
	}
	return TrunkToPb(created), nil
}

func (s *Server) GetSipTrunk(ctx context.Context, req *pb.GetSipTrunkRequest) (*pb.SipTrunk, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	t, err := s.store.GetTrunk(req.GetId(), md.OrganisationID.String())
	if err != nil {
		return nil, status.Error(codes.NotFound, "trunk not found")
	}
	return TrunkToPb(t), nil
}

func (s *Server) ListSipTrunks(ctx context.Context, req *pb.ListSipTrunksRequest) (*pb.ListSipTrunksResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.store.ListTrunks(md.OrganisationID.String())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list trunks")
	}
	out := make([]*pb.SipTrunk, 0, len(rows))
	for _, r := range rows {
		out = append(out, TrunkToPb(r))
	}
	return &pb.ListSipTrunksResponse{Trunks: out}, nil
}

func (s *Server) UpdateSipTrunk(ctx context.Context, req *pb.UpdateSipTrunkRequest) (*pb.SipTrunk, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	in := TrunkFromPb(req.GetTrunk())
	if in == nil || in.ID == "" {
		return nil, status.Error(codes.InvalidArgument, "trunk.id is required")
	}
	in.OrgID = md.OrganisationID.String()
	updated, err := s.store.UpdateTrunk(in)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update trunk")
	}
	return TrunkToPb(updated), nil
}

func (s *Server) DeleteSipTrunk(ctx context.Context, req *pb.DeleteSipTrunkRequest) (*pb.DeleteSipTrunkResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.store.DeleteTrunk(req.GetId(), md.OrganisationID.String()); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete trunk")
	}
	return &pb.DeleteSipTrunkResponse{Ok: true}, nil
}

// ── SIP domains ─────────────────────────────────────────────────────────────

func (s *Server) CreateSipDomain(ctx context.Context, req *pb.CreateSipDomainRequest) (*pb.SipDomain, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	in := DomainFromPb(req.GetDomain())
	if in == nil || in.Domain == "" {
		return nil, status.Error(codes.InvalidArgument, "domain is required")
	}
	in.OrgID = md.OrganisationID.String()
	created, err := s.store.CreateDomain(in)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create domain")
	}
	return DomainToPb(created), nil
}

func (s *Server) ListSipDomains(ctx context.Context, req *pb.ListSipDomainsRequest) (*pb.ListSipDomainsResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.store.ListDomains(md.OrganisationID.String())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list domains")
	}
	out := make([]*pb.SipDomain, 0, len(rows))
	for _, r := range rows {
		out = append(out, DomainToPb(r))
	}
	return &pb.ListSipDomainsResponse{Domains: out}, nil
}

func (s *Server) DeleteSipDomain(ctx context.Context, req *pb.DeleteSipDomainRequest) (*pb.DeleteSipDomainResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.store.DeleteDomain(req.GetId(), md.OrganisationID.String()); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete domain")
	}
	return &pb.DeleteSipDomainResponse{Ok: true}, nil
}

// ── Softphone token ──────────────────────────────────────────────────────────

func (s *Server) IssueSoftphoneToken(ctx context.Context, req *pb.IssueSoftphoneTokenRequest) (*pb.IssueSoftphoneTokenResponse, error) {
	md, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if s.issuer == nil {
		return nil, status.Error(codes.FailedPrecondition, "softphone token issuer not configured")
	}
	userID := req.GetAgentId()
	if userID == "" {
		userID = md.UserID.String()
	}
	ext, err := s.store.GetExtensionByUserID(userID)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "no SIP extension registered for user")
	}
	const ttl = 30 * time.Minute
	tok, err := s.issuer.IssueSoftphoneToken(md.OrganisationID.String(), userID, ext.Extension, ttl)
	if err != nil {
		return nil, status.Error(codes.Internal, "issue token")
	}
	ice := []*pb.IceServer{
		{Urls: "stun:" + s.turnHost()},
	}
	if s.turnURL != "" {
		ice = append(ice, &pb.IceServer{
			Urls:       s.turnURL,
			Username:   s.turnUser,
			Credential: s.turnPass,
		})
	}
	return &pb.IssueSoftphoneTokenResponse{
		Token: &pb.SoftphoneToken{
			Token:        tok,
			WsUrl:        s.wsURL,
			SipRealm:     s.sipRealm,
			SipExtension: ext.Extension,
			IceServers:   ice,
			ExpiresIn:    int32(ttl.Seconds()),
		},
	}, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func (s *Server) requireAuth(ctx context.Context) (*authctx.RequestMetadata, error) {
	md, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || md.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}
	return md, nil
}

func (s *Server) publishCallEvent(c *Call, subject string) {
	if s.publisher == nil {
		return
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"call_id":         c.ID,
		"direction":       c.Direction,
		"from_number":     c.FromNumber,
		"to_number":       c.ToNumber,
		"agent_id":        c.AgentID,
		"contact_id":      c.ContactID,
		"conversation_id": c.ConversationID,
	})
	evt, err := events.NewEvent(c.OrgID, c.WorkspaceID, "call",
		subjectAction(subject), c.ID, c.AgentID, payload)
	if err != nil {
		return
	}
	evt.Subject = subject
	_ = s.publisher.Publish(evt)
}

// subjectAction extracts the action from a subject like "call.started" -> "started".
func subjectAction(subject string) string {
	for i := len(subject) - 1; i >= 0; i-- {
		if subject[i] == '.' {
			return subject[i+1:]
		}
	}
	return subject
}

// turnHost extracts the host part from the configured TURN URL for use as a
// STUN URL fallback (turn://host:port -> stun:host:port).
func (s *Server) turnHost() string {
	if s.turnURL == "" {
		return "stun.l.google.com:19302" // public fallback
	}
	const prefix = "turn:"
	if len(s.turnURL) > len(prefix) && s.turnURL[:len(prefix)] == prefix {
		return s.turnURL[len(prefix):]
	}
	return s.turnURL
}

// newSipPassword generates a 32-char URL-safe SIP password.
func newSipPassword() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	now := time.Now().UnixNano()
	for i := range b {
		now = now*1103515245 + 12345
		b[i] = chars[uint(now>>16)%uint(len(chars))]
	}
	return string(b)
}
