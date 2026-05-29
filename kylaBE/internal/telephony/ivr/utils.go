package ivr

import (
	"encoding/json"

	"kyla-be/pkg/pb"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ── Flow <-> proto ───────────────────────────────────────────────────────────

func FlowToPb(f *Flow) *pb.IVRFlow {
	if f == nil {
		return nil
	}
	return &pb.IVRFlow{
		Id:          f.ID,
		OrgId:       f.OrgID,
		WorkspaceId: f.WorkspaceID,
		Name:        f.Name,
		Description: f.Description,
		Definition:  definitionToPb(f.Definition),
		IsActive:    f.IsActive,
		Version:     int32(f.Version),
		CreatedBy:   f.CreatedBy,
		CreatedAt:   timestamppb.New(f.CreatedAt),
		UpdatedAt:   timestamppb.New(f.UpdatedAt),
	}
}

// FlowFromPb maps the writable fields from a proto. ID, version, audit
// fields are managed server-side.
func FlowFromPb(p *pb.IVRFlow) *Flow {
	if p == nil {
		return nil
	}
	return &Flow{
		ID:          p.GetId(),
		OrgID:       p.GetOrgId(),
		WorkspaceID: p.GetWorkspaceId(),
		Name:        p.GetName(),
		Description: p.GetDescription(),
		Definition:  definitionFromPb(p.GetDefinition()),
		IsActive:    p.GetIsActive(),
	}
}

func definitionToPb(raw json.RawMessage) *pb.IVRDefinition {
	d, err := DecodeDefinition(raw)
	if err != nil {
		return &pb.IVRDefinition{}
	}
	out := &pb.IVRDefinition{StartNodeId: d.StartNodeID}
	for _, n := range d.Nodes {
		var cfg *structpb.Struct
		if len(n.Config) > 0 {
			if s, err := structpb.NewStruct(n.Config); err == nil {
				cfg = s
			}
		}
		out.Nodes = append(out.Nodes, &pb.IVRNode{
			Id:         n.ID,
			Type:       string(n.Type),
			Config:     cfg,
			NextNodeId: n.NextNodeID,
			Branches:   n.Branches,
		})
	}
	return out
}

func definitionFromPb(p *pb.IVRDefinition) json.RawMessage {
	if p == nil {
		return json.RawMessage(`{}`)
	}
	d := Definition{StartNodeID: p.GetStartNodeId()}
	for _, n := range p.GetNodes() {
		node := Node{
			ID:         n.GetId(),
			Type:       NodeType(n.GetType()),
			NextNodeID: n.GetNextNodeId(),
			Branches:   n.GetBranches(),
		}
		if n.GetConfig() != nil {
			node.Config = n.GetConfig().AsMap()
		}
		d.Nodes = append(d.Nodes, node)
	}
	b, _ := json.Marshal(d)
	return b
}

// ── DIDMapping <-> proto ────────────────────────────────────────────────────

func DIDMappingToPb(m *DIDMapping) *pb.IVRDIDMapping {
	if m == nil {
		return nil
	}
	return &pb.IVRDIDMapping{
		Id:          m.ID,
		OrgId:       m.OrgID,
		WorkspaceId: m.WorkspaceID,
		Did:         m.DID,
		FlowId:      m.FlowID,
		CreatedAt:   timestamppb.New(m.CreatedAt),
	}
}

func DIDMappingFromPb(p *pb.IVRDIDMapping) *DIDMapping {
	if p == nil {
		return nil
	}
	return &DIDMapping{
		ID:          p.GetId(),
		OrgID:       p.GetOrgId(),
		WorkspaceID: p.GetWorkspaceId(),
		DID:         p.GetDid(),
		FlowID:      p.GetFlowId(),
	}
}

// ── Run <-> proto ───────────────────────────────────────────────────────────

func RunToPb(r *Run) *pb.IVRRun {
	if r == nil {
		return nil
	}
	out := &pb.IVRRun{
		Id:            r.ID,
		FlowId:        r.FlowID,
		CallId:        r.CallID,
		OrgId:         r.OrgID,
		WorkspaceId:   r.WorkspaceID,
		Status:        r.Status,
		CurrentNodeId: r.CurrentNodeID,
		StartedAt:     timestamppb.New(r.StartedAt),
		EndReason:     r.EndReason,
	}
	if r.EndedAt != nil {
		out.EndedAt = timestamppb.New(*r.EndedAt)
	}
	var steps []RunStep
	if len(r.VisitedNodes) > 0 {
		_ = json.Unmarshal(r.VisitedNodes, &steps)
	}
	for _, s := range steps {
		out.VisitedNodes = append(out.VisitedNodes, &pb.IVRRunStep{
			NodeId:    s.NodeID,
			EnteredAt: timestamppb.New(s.EnteredAt),
			Input:     s.Input,
		})
	}
	return out
}
