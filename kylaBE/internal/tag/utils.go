package tag

import (
	"kyla-be/internal/authctx"
	"kyla-be/pkg/pb"

	"github.com/google/uuid"
)

func PbTagToTag(tag *pb.Tag) *Tag {
	id, err := uuid.Parse(tag.GetId())
	if err != nil {
		id = uuid.New()
	}
	return &Tag{
		ID:        id,
		ColorCode: tag.GetColorCode(),
		Name:      tag.GetName(),
		CreatedBy: tag.GetCreatedBy(),
		OwnerType: authctx.OwnerType(tag.GetOwnerType()),
		OwnerID:   uuid.MustParse(tag.GetOwnerId()),
	}
}

func TagToPbTag(tag *Tag) *pb.Tag {
	return &pb.Tag{
		Id:        tag.ID.String(),
		ColorCode: tag.ColorCode,
		Name:      tag.Name,
		CreatedBy: tag.CreatedBy,
		CreatedAt: tag.CreatedAt.String(),
		UpdatedAt: tag.UpdatedAt.String(),
		OwnerType: pb.OwnerType(pb.OwnerType_value[string(tag.OwnerType)]),
		OwnerId:   tag.OwnerID.String(),
	}
}

func PbTagsToTags(tags []*pb.Tag) []*Tag {
	var ts []*Tag
	for _, tag := range tags {
		ts = append(ts, PbTagToTag(tag))
	}
	return ts
}

func TagsToPbTags(tags []*Tag) []*pb.Tag {
	var ts []*pb.Tag
	for _, tag := range tags {
		ts = append(ts, TagToPbTag(tag))
	}
	return ts
}
