package service

import (
	"context"
	"fmt"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/utils"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LabelService struct {
	pb.UnimplementedLabelServiceServer
	labelStore *LabelStore
	AuthStore  *AuthStore
}

func NewLabelService(labelStore *LabelStore, authStore *AuthStore) *LabelService {
	return &LabelService{
		labelStore: labelStore,
		AuthStore:  authStore,
	}
}

func (s *LabelService) CreateLabel(ctx context.Context, request *pb.CreateLabelRequest) (*pb.CreateLabelResponse, error) {
	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to create a label")
	}

	scope := PbScopeToOpScope(request.GetScope())
	label := PbLabelToLabel(request.GetLabel())

	// Validate required fields
	if err := utils.ValidateRequiredFields(k.NewConsts().LabelsRequiredFields, label); err != nil {
		return nil, err
	}

	// Generate UUID for the label
	labelID := uuid.New()
	label.ID = labelID
	label.CreatedBy = contextData.UserID.String()
	label.OwnerID = uuid.MustParse(scope.ID)
	label.OwnerType = scope.Owner

	// Save the label to the store
	err := s.labelStore.Save(label)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create label: %v", err))
	}

	return &pb.CreateLabelResponse{
		Label: LabelToPbLabel(label),
	}, nil
}

func (s *LabelService) ReadLabel(ctx context.Context, request *pb.ReadLabelRequest) (*pb.ReadLabelResponse, error) {
	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to create a label")
	}

	labelID := request.GetId()

	organisationId := contextData.OrganisationID
	// Find the label by ID
	label, err := s.labelStore.FindById(labelID, organisationId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("label not found with ID %s", labelID))
	}

	return &pb.ReadLabelResponse{
		Label: LabelToPbLabel(label),
	}, nil
}

func (s *LabelService) ReadLabels(ctx context.Context, request *pb.ReadLabelsRequest) (*pb.ReadLabelsResponse, error) {
	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to create a label")
	}

	organisationId := contextData.OrganisationID

	// Use the correct method name here
	labels, err := s.labelStore.FindByOrganisationID(organisationId)
	if err != nil {
		return nil, fmt.Errorf("failed to read labels: %v", err)
	}

	response := &pb.ReadLabelsResponse{
		Labels: LabelsToPbLabels(labels),
	}

	return response, nil
}

func (s *LabelService) UpdateLabel(ctx context.Context, request *pb.UpdateLabelRequest) (*pb.UpdateLabelResponse, error) {
	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to create a label")
	}

	organisationId := contextData.OrganisationID

	label := PbLabelToLabel(request.GetLabel())

	// Validate required fields
	if err := utils.ValidateRequiredFields(k.NewConsts().LabelsRequiredFields, label); err != nil {
		return nil, err
	}

	// Save the updated label to the store
	err := s.labelStore.Update(label, label.ID.String(), organisationId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update label: %v", err))
	}

	return &pb.UpdateLabelResponse{
		Label: LabelToPbLabel(label),
	}, nil
}

func (s *LabelService) DeleteLabel(ctx context.Context, request *pb.DeleteLabelRequest) (*pb.DeleteLabelResponse, error) {
	contextData, authErr := s.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden: You don't have permission to delete a label")
	}

	organisationId := contextData.OrganisationID

	labelId := request.GetId()

	// Delete the label by id
	err := s.labelStore.Delete(labelId, organisationId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("label not found with id %s", labelId))
	}

	return &pb.DeleteLabelResponse{
		Success: true,
	}, nil
}
