package queues

import (
	"kyla-be/pkg/pb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ── Queue <-> proto ─────────────────────────────────────────────────────────

func QueueToPb(q *Queue) *pb.Queue {
	if q == nil {
		return nil
	}
	return &pb.Queue{
		Id:              q.ID,
		OrgId:           q.OrgID,
		WorkspaceId:     q.WorkspaceID,
		Name:            q.Name,
		Description:     q.Description,
		Strategy:        q.Strategy,
		MohPath:         q.MOHPath,
		MaxWaitSeconds:  int32(q.MaxWaitSeconds),
		OverflowAction:  q.OverflowAction,
		OverflowTarget:  q.OverflowTarget,
		IsActive:        q.IsActive,
		CreatedAt:       timestamppb.New(q.CreatedAt),
		UpdatedAt:       timestamppb.New(q.UpdatedAt),
	}
}

func QueueFromPb(p *pb.Queue) *Queue {
	if p == nil {
		return nil
	}
	return &Queue{
		ID:              p.GetId(),
		OrgID:           p.GetOrgId(),
		WorkspaceID:     p.GetWorkspaceId(),
		Name:            p.GetName(),
		Description:     p.GetDescription(),
		Strategy:        p.GetStrategy(),
		MOHPath:         p.GetMohPath(),
		MaxWaitSeconds:  int(p.GetMaxWaitSeconds()),
		OverflowAction:  p.GetOverflowAction(),
		OverflowTarget:  p.GetOverflowTarget(),
		IsActive:        p.GetIsActive(),
	}
}

// ── Membership <-> proto ────────────────────────────────────────────────────

func MembershipToPb(m *Membership) *pb.QueueMembership {
	if m == nil {
		return nil
	}
	out := &pb.QueueMembership{
		Id:        m.ID,
		QueueId:   m.QueueID,
		OrgId:     m.OrgID,
		UserId:    m.UserID,
		Priority:  int32(m.Priority),
		IsActive:  m.IsActive,
		CreatedAt: timestamppb.New(m.CreatedAt),
		UpdatedAt: timestamppb.New(m.UpdatedAt),
	}
	if m.LastCallEndedAt != nil {
		out.LastCallEndedAt = timestamppb.New(*m.LastCallEndedAt)
	}
	return out
}

func MembershipFromPb(p *pb.QueueMembership) *Membership {
	if p == nil {
		return nil
	}
	return &Membership{
		ID:       p.GetId(),
		QueueID:  p.GetQueueId(),
		OrgID:    p.GetOrgId(),
		UserID:   p.GetUserId(),
		Priority: int(p.GetPriority()),
		IsActive: p.GetIsActive(),
	}
}

// ── Entry <-> proto ─────────────────────────────────────────────────────────

func EntryToPb(e *Entry) *pb.QueueEntry {
	if e == nil {
		return nil
	}
	out := &pb.QueueEntry{
		Id:              e.ID,
		QueueId:         e.QueueID,
		CallId:          e.CallID,
		OrgId:           e.OrgID,
		WorkspaceId:     e.WorkspaceID,
		Priority:        int32(e.Priority),
		Status:          e.Status,
		AssignedAgentId: e.AssignedAgentID,
		EnteredAt:       timestamppb.New(e.EnteredAt),
		EndedReason:     e.EndedReason,
	}
	if e.AssignedAt != nil {
		out.AssignedAt = timestamppb.New(*e.AssignedAt)
	}
	if e.EndedAt != nil {
		out.EndedAt = timestamppb.New(*e.EndedAt)
	}
	return out
}
