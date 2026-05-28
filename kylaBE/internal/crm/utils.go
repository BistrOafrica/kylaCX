package crm

import (
	"encoding/json"

	"kyla-be/pkg/pb"
)

// PipelineToPb converts a Pipeline model to its proto representation.
func PipelineToPb(p *Pipeline, stagesCount int) *pb.Pipeline {
	return &pb.Pipeline{
		Id:          p.ID,
		OrgId:       p.OrgID,
		WorkspaceId: p.WorkspaceID,
		Name:        p.Name,
		Description: p.Description,
		Type:        typeStringToPb(p.Type),
		Color:       p.Color,
		StagesCount: int32(stagesCount),
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// StageToPb converts a PipelineStage model to its proto representation.
func StageToPb(s *PipelineStage) *pb.PipelineStage {
	return &pb.PipelineStage{
		Id:          s.ID,
		PipelineId:  s.PipelineID,
		OrgId:       s.OrgID,
		Name:        s.Name,
		Color:       s.Color,
		Index:       int32(s.Index),
		Probability: int32(s.Probability),
		CreatedAt:   s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// StagesToPb converts a slice of PipelineStage models to proto.
func StagesToPb(stages []*PipelineStage) []*pb.PipelineStage {
	out := make([]*pb.PipelineStage, len(stages))
	for i, s := range stages {
		out[i] = StageToPb(s)
	}
	return out
}

// typeStringToPb maps a pipeline type string to the proto enum.
func typeStringToPb(t string) pb.PipelineType {
	switch t {
	case "sales":
		return pb.PipelineType_PIPELINE_TYPE_SALES
	case "operations":
		return pb.PipelineType_PIPELINE_TYPE_OPERATIONS
	case "custom":
		return pb.PipelineType_PIPELINE_TYPE_CUSTOM
	default:
		return pb.PipelineType_PIPELINE_TYPE_UNSPECIFIED
	}
}

// typeFromPb maps the proto pipeline type to a string.
func typeFromPb(t pb.PipelineType) string {
	switch t {
	case pb.PipelineType_PIPELINE_TYPE_SALES:
		return "sales"
	case pb.PipelineType_PIPELINE_TYPE_OPERATIONS:
		return "operations"
	case pb.PipelineType_PIPELINE_TYPE_CUSTOM:
		return "custom"
	default:
		return "sales"
	}
}

// DealCardFromRow converts a DealRow into a pb.DealCard.
func DealCardFromRow(row DealRow) *pb.DealCard {
	card := &pb.DealCard{ObjectId: row.ID}
	if len(row.Data) == 0 {
		return card
	}
	var m map[string]interface{}
	if err := json.Unmarshal(row.Data, &m); err != nil {
		return card
	}
	if v, ok := m["name"].(string); ok {
		card.Name = v
	}
	if v, ok := m["value"].(string); ok {
		card.Value = v
	}
	if v, ok := m["assignee"].(string); ok {
		card.AssigneeId = v
	}
	if v, ok := m["close_date"].(string); ok {
		card.CloseDate = v
	}
	if v, ok := m["probability"].(float64); ok {
		card.Probability = int32(v)
	}
	return card
}
