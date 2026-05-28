package tag

import (
	"context"
	"fmt"
	"kyla-be/internal/authctx"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthGateway defines the auth methods the TagService needs.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
}

type TagService struct {
	pb.UnimplementedTagServiceServer
	tagStore  *TagStore
	AuthStore AuthGateway
}

func NewTagService(tagStore *TagStore, authStore AuthGateway) *TagService {
	return &TagService{
		tagStore:  tagStore,
		AuthStore: authStore,
	}
}

func (s *TagService) CreateTag(ctx context.Context, request *pb.CreateTagRequest) (*pb.CreateTagResponse, error) {
	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to create a tag")
	}

	scope := authctx.PbScopeToOpScope(request.GetScope())
	tag := PbTagToTag(request.GetTag())

	// Validate required fields
	if err := utils.ValidateRequiredFields(k.NewConsts().TagsRequiredFields, tag); err != nil {
		return nil, err
	}

	tagID := uuid.New()
	tag.ID = tagID
	tag.CreatedBy = contextData.UserID.String()
	tag.OwnerID = uuid.MustParse(scope.ID)
	tag.OwnerType = scope.Owner

	err := s.tagStore.Save(tag)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create tag: %v", err))
	}

	return &pb.CreateTagResponse{
		Tag: TagToPbTag(tag),
	}, nil
}

func (s *TagService) ReadTag(ctx context.Context, request *pb.ReadTagRequest) (*pb.ReadTagResponse, error) {
	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to create a tag")
	}

	tagID := request.GetId()
	organisationId := contextData.OrganisationID

	tag, err := s.tagStore.FindById(tagID, organisationId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("tag not found with ID %s", tagID))
	}

	return &pb.ReadTagResponse{
		Tag: TagToPbTag(tag),
	}, nil
}

func (s *TagService) ReadTags(ctx context.Context, request *pb.ReadTagsRequest) (*pb.ReadTagsResponse, error) {
	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to create a tag")
	}

	organisationId := contextData.OrganisationID

	tags, err := s.tagStore.FindByOrganisationID(organisationId)
	if err != nil {
		return nil, fmt.Errorf("failed to read tags: %v", err)
	}

	return &pb.ReadTagsResponse{
		Tags: TagsToPbTags(tags),
	}, nil
}

func (s *TagService) UpdateTag(ctx context.Context, request *pb.UpdateTagRequest) (*pb.UpdateTagResponse, error) {
	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to create a tag")
	}

	organisationId := contextData.OrganisationID
	tag := PbTagToTag(request.GetTag())

	if err := utils.ValidateRequiredFields(k.NewConsts().TagsRequiredFields, tag); err != nil {
		return nil, err
	}

	err := s.tagStore.Update(tag, tag.ID.String(), organisationId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update tag: %v", err))
	}

	return &pb.UpdateTagResponse{
		Tag: TagToPbTag(tag),
	}, nil
}

func (s *TagService) DeleteTag(ctx context.Context, request *pb.DeleteTagRequest) (*pb.DeleteTagResponse, error) {
	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to delete a tag")
	}

	organisationId := contextData.OrganisationID
	tagId := request.GetId()

	err := s.tagStore.Delete(tagId, organisationId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("tag not found with id %s", tagId))
	}

	return &pb.DeleteTagResponse{
		Success: true,
	}, nil
}
