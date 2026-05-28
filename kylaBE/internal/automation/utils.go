package automation

import (
	"encoding/json"

	"kyla-be/pkg/pb"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ── proto.Struct ↔ Go types ──────────────────────────────────────────────────

func structToMap(s *structpb.Struct) map[string]interface{} {
	if s == nil {
		return nil
	}
	return s.AsMap()
}

func mapToStruct(m map[string]interface{}) *structpb.Struct {
	if m == nil {
		return nil
	}
	s, err := structpb.NewStruct(m)
	if err != nil {
		return nil
	}
	return s
}

func triggerFromStruct(s *structpb.Struct) TriggerConfig {
	var t TriggerConfig
	if s == nil {
		return t
	}
	raw, err := json.Marshal(s.AsMap())
	if err != nil {
		return t
	}
	_ = json.Unmarshal(raw, &t)
	return t
}

func triggerToStruct(t TriggerConfig) *structpb.Struct {
	raw, err := json.Marshal(t)
	if err != nil {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	return mapToStruct(m)
}

func conditionsFromStructs(ss []*structpb.Struct) []ConditionGroup {
	if len(ss) == 0 {
		return nil
	}
	out := make([]ConditionGroup, 0, len(ss))
	for _, s := range ss {
		raw, err := json.Marshal(s.AsMap())
		if err != nil {
			continue
		}
		var g ConditionGroup
		if err := json.Unmarshal(raw, &g); err != nil {
			continue
		}
		out = append(out, g)
	}
	return out
}

func conditionsToStructs(groups []ConditionGroup) []*structpb.Struct {
	if len(groups) == 0 {
		return nil
	}
	out := make([]*structpb.Struct, 0, len(groups))
	for _, g := range groups {
		raw, err := json.Marshal(g)
		if err != nil {
			continue
		}
		var m map[string]interface{}
		if err := json.Unmarshal(raw, &m); err != nil {
			continue
		}
		if s := mapToStruct(m); s != nil {
			out = append(out, s)
		}
	}
	return out
}

func actionsFromStructs(ss []*structpb.Struct) []ActionNode {
	if len(ss) == 0 {
		return nil
	}
	out := make([]ActionNode, 0, len(ss))
	for _, s := range ss {
		raw, err := json.Marshal(s.AsMap())
		if err != nil {
			continue
		}
		var a ActionNode
		if err := json.Unmarshal(raw, &a); err != nil {
			continue
		}
		out = append(out, a)
	}
	return out
}

func actionsToStructs(actions []ActionNode) []*structpb.Struct {
	if len(actions) == 0 {
		return nil
	}
	out := make([]*structpb.Struct, 0, len(actions))
	for _, a := range actions {
		raw, err := json.Marshal(a)
		if err != nil {
			continue
		}
		var m map[string]interface{}
		if err := json.Unmarshal(raw, &m); err != nil {
			continue
		}
		if s := mapToStruct(m); s != nil {
			out = append(out, s)
		}
	}
	return out
}

// ── model ↔ pb ────────────────────────────────────────────────────────────────

func WorkflowToPb(w *Workflow) *pb.Workflow {
	if w == nil {
		return nil
	}
	return &pb.Workflow{
		Id:          w.ID,
		OrgId:       w.OrgID,
		WorkspaceId: w.WorkspaceID,
		Name:        w.Name,
		Description: w.Description,
		Trigger:     triggerToStruct(w.Trigger),
		Conditions:  conditionsToStructs(w.Conditions),
		Actions:     actionsToStructs(w.Actions),
		Status:      string(w.Status),
		CreatedAt:   timestamppb.New(w.CreatedAt),
		UpdatedAt:   timestamppb.New(w.UpdatedAt),
		CreatedBy:   w.CreatedBy,
	}
}

func WorkflowRunToPb(r *WorkflowRun) *pb.WorkflowRun {
	if r == nil {
		return nil
	}
	out := &pb.WorkflowRun{
		Id:             r.ID,
		WorkflowId:     r.WorkflowID,
		TemporalRunId:  r.TemporalRunID,
		TriggerEventId: r.TriggerEventID,
		Status:         string(r.Status),
		Error:          r.Error,
	}
	if r.StartedAt != nil {
		out.StartedAt = timestamppb.New(*r.StartedAt)
	}
	if r.FinishedAt != nil {
		out.FinishedAt = timestamppb.New(*r.FinishedAt)
	}
	if len(r.Context) > 0 {
		var m map[string]interface{}
		if err := json.Unmarshal(r.Context, &m); err == nil {
			out.Context = mapToStruct(m)
		}
	}
	return out
}
