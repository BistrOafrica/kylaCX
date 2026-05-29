package telephony

import (
	"time"

	"kyla-be/pkg/pb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ── Call <-> proto ───────────────────────────────────────────────────────────

func CallToPb(c *Call) *pb.Call {
	if c == nil {
		return nil
	}
	out := &pb.Call{
		Id:               c.ID,
		OrgId:            c.OrgID,
		WorkspaceId:      c.WorkspaceID,
		Direction:        c.Direction,
		Status:           c.Status,
		FromNumber:       c.FromNumber,
		ToNumber:         c.ToNumber,
		AgentId:          c.AgentID,
		ContactId:        c.ContactID,
		QueueId:          c.QueueID,
		TrunkId:          c.TrunkID,
		IvrFlowId:        c.IvrFlowID,
		ConversationId:   c.ConversationID,
		DealId:           c.DealID,
		TicketId:         c.TicketID,
		RecordingEnabled: c.RecordingEnabled,
		RecordingUrl:     c.RecordingURL,
		RingSeconds:      int32(c.RingSeconds),
		TalkSeconds:      int32(c.TalkSeconds),
		HangupCause:      c.HangupCause,
		Disposition:      c.Disposition,
		StartedAt:        timestamppb.New(c.StartedAt),
		CreatedAt:        timestamppb.New(c.CreatedAt),
		UpdatedAt:        timestamppb.New(c.UpdatedAt),
	}
	if c.AnsweredAt != nil {
		out.AnsweredAt = timestamppb.New(*c.AnsweredAt)
	}
	if c.EndedAt != nil {
		out.EndedAt = timestamppb.New(*c.EndedAt)
	}
	return out
}

// ── CallEvent <-> proto ──────────────────────────────────────────────────────

func CallEventToPb(e *CallEvent) *pb.CallEvent {
	if e == nil {
		return nil
	}
	return &pb.CallEvent{
		Id:          e.ID,
		CallSessionId: e.CallID,
		OrgId:       e.OrgID,
		EventType:   e.EventType,
		Detail:      string(e.Detail),
		At:          timestamppb.New(e.At),
	}
}

// ── SipExtension <-> proto ──────────────────────────────────────────────────

func ExtensionToPb(e *SipExtension) *pb.SipExtension {
	if e == nil {
		return nil
	}
	out := &pb.SipExtension{
		Id:          e.ID,
		OrgId:       e.OrgID,
		WorkspaceId: e.WorkspaceID,
		UserId:      e.UserID,
		Extension:   e.Extension,
		DisplayName: e.DisplayName,
		Status:      e.Status,
		CreatedAt:   timestamppb.New(e.CreatedAt),
		UpdatedAt:   timestamppb.New(e.UpdatedAt),
	}
	if e.LastRegistration != nil {
		out.LastRegistration = timestamppb.New(*e.LastRegistration)
	}
	return out
}

func ExtensionFromPb(p *pb.SipExtension) *SipExtension {
	if p == nil {
		return nil
	}
	return &SipExtension{
		ID:          p.GetId(),
		OrgID:       p.GetOrgId(),
		WorkspaceID: p.GetWorkspaceId(),
		UserID:      p.GetUserId(),
		Extension:   p.GetExtension(),
		DisplayName: p.GetDisplayName(),
		Status:      p.GetStatus(),
	}
}

// ── SipTrunk <-> proto ──────────────────────────────────────────────────────

// TrunkToPb returns the trunk with the password field zeroed — we never
// expose the SIP trunk credential over gRPC reads.
func TrunkToPb(t *SipTrunk) *pb.SipTrunk {
	if t == nil {
		return nil
	}
	return &pb.SipTrunk{
		Id:          t.ID,
		OrgId:       t.OrgID,
		Name:        t.Name,
		GatewayName: t.GatewayName,
		Provider:    t.Provider,
		SipServer:   t.SipServer,
		Username:    t.Username,
		Password:    "", // never expose
		FromUri:     t.FromURI,
		IsActive:    t.IsActive,
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}
}

func TrunkFromPb(p *pb.SipTrunk) *SipTrunk {
	if p == nil {
		return nil
	}
	return &SipTrunk{
		ID:          p.GetId(),
		OrgID:       p.GetOrgId(),
		Name:        p.GetName(),
		GatewayName: p.GetGatewayName(),
		Provider:    p.GetProvider(),
		SipServer:   p.GetSipServer(),
		Username:    p.GetUsername(),
		Password:    p.GetPassword(),
		FromURI:     p.GetFromUri(),
		IsActive:    p.GetIsActive(),
	}
}

// ── SipDomain <-> proto ─────────────────────────────────────────────────────

func DomainToPb(d *SipDomain) *pb.SipDomain {
	if d == nil {
		return nil
	}
	return &pb.SipDomain{
		Id:        d.ID,
		OrgId:     d.OrgID,
		Domain:    d.Domain,
		IsDefault: d.IsDefault,
		CreatedAt: timestamppb.New(d.CreatedAt),
	}
}

func DomainFromPb(p *pb.SipDomain) *SipDomain {
	if p == nil {
		return nil
	}
	return &SipDomain{
		ID:        p.GetId(),
		OrgID:     p.GetOrgId(),
		Domain:    p.GetDomain(),
		IsDefault: p.GetIsDefault(),
	}
}

// computeRingSeconds returns the time between started_at and answered_at.
// Used by the ESL hangup handler when the PBX doesn't supply ring duration.
func computeRingSeconds(started time.Time, answered *time.Time, ended time.Time) int {
	if answered != nil {
		return int(answered.Sub(started).Seconds())
	}
	return int(ended.Sub(started).Seconds())
}

// computeTalkSeconds returns the duration between answered_at and ended_at.
func computeTalkSeconds(answered *time.Time, ended time.Time) int {
	if answered == nil {
		return 0
	}
	return int(ended.Sub(*answered).Seconds())
}
