package label

import (
	"kyla-be/internal/authctx"
	"kyla-be/pkg/pb"

	"github.com/google/uuid"
)

func PbLabelToLabel(label *pb.Label) *Label {
	id, err := uuid.Parse(label.GetId())
	if err != nil {
		id = uuid.New()
	}
	return &Label{
		ID:        id,
		Name:      label.GetName(),
		CreatedBy: label.GetCreatedBy(),
		OwnerType: authctx.OwnerType(pb.OwnerType_name[int32(label.GetOwnerType())]),
		OwnerID:   uuid.MustParse(label.GetOwnerId()),
	}
}

func LabelToPbLabel(label *Label) *pb.Label {
	return &pb.Label{
		Id:        label.ID.String(),
		Name:      label.Name,
		CreatedBy: label.CreatedBy,
		CreatedAt: label.CreatedAt.String(),
		UpdatedAt: label.UpdatedAt.String(),
		OwnerType: pb.OwnerType(pb.OwnerType_value[string(label.OwnerType)]),
		OwnerId:   label.OwnerID.String(),
	}
}

func PbLabelsToLabels(labels []*pb.Label) []*Label {
	var ls []*Label
	for _, label := range labels {
		ls = append(ls, PbLabelToLabel(label))
	}
	return ls
}

func LabelsToPbLabels(labels []*Label) []*pb.Label {
	var ls []*pb.Label
	for _, label := range labels {
		ls = append(ls, LabelToPbLabel(label))
	}
	return ls
}
