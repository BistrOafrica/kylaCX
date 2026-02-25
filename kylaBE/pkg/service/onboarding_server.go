package service

import (
	"context"
	"kyla-be/pkg/pb"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type OnboardingServiceServer struct {
	pb.UnimplementedOnboardingServiceServer
	store *OnboardingStore
}

func NewOnboardingServiceServer(db *gorm.DB) *OnboardingServiceServer {
	return &OnboardingServiceServer{
		store: NewOnboardingStore(db),
	}
}

func (s *OnboardingServiceServer) CreateOnboarding(ctx context.Context, req *pb.CreateOnboardingRequest) (*pb.CreateOnboardingResponse, error) {
	onboarding := &Onboarding{
		ID:            uuid.New(),
		Timestamp:     req.Timestamp,
		Status:        req.Status,
		Packages:      req.Packages,
		Products:      req.Products,
		NumberOfUsers: int(req.NumberOfUsers),
		Remarks:       req.Remarks,
		ContactEmail:  req.ContactEmail,
		ContactPhone:  req.ContactPhone,
		Name:          req.Name,
		Metadata:      req.Metadata,
	}

	createdOnboarding, err := s.store.CreateOnboarding(onboarding)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create onboarding: %v", err)
	}

	return &pb.CreateOnboardingResponse{
		Onboarding: toPbOnboarding(createdOnboarding),
	}, nil
}

func (s *OnboardingServiceServer) ReadOnboarding(ctx context.Context, req *pb.ReadOnboardingRequest) (*pb.ReadOnboardingResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	onboarding, err := s.store.GetOnboardingByID(req.Id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "onboarding not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to get onboarding: %v", err)
	}

	return &pb.ReadOnboardingResponse{
		Onboarding: toPbOnboarding(onboarding),
	}, nil
}

func (s *OnboardingServiceServer) UpdateOnboarding(ctx context.Context, req *pb.UpdateOnboardingRequest) (*pb.UpdateOnboardingResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	uuidID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid uuid: %v", err)
	}

	onboarding := &Onboarding{
		ID:            uuidID,
		Timestamp:     req.Timestamp,
		Status:        req.Status,
		Packages:      req.Packages,
		Products:      req.Products,
		NumberOfUsers: int(req.NumberOfUsers),
		Remarks:       req.Remarks,
		ContactEmail:  req.ContactEmail,
		ContactPhone:  req.ContactPhone,
		Name:          req.Name,
		Metadata:      req.Metadata,
	}

	updatedOnboarding, err := s.store.UpdateOnboarding(onboarding)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update onboarding: %v", err)
	}

	return &pb.UpdateOnboardingResponse{
		Onboarding: toPbOnboarding(updatedOnboarding),
	}, nil
}

func (s *OnboardingServiceServer) DeleteOnboarding(ctx context.Context, req *pb.DeleteOnboardingRequest) (*pb.DeleteOnboardingResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	err := s.store.DeleteOnboarding(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete onboarding: %v", err)
	}

	return &pb.DeleteOnboardingResponse{
		Success: true,
		Message: "onboarding deleted successfully",
	}, nil
}

func (s *OnboardingServiceServer) ListOnboardings(ctx context.Context, req *pb.ListOnboardingsRequest) (*pb.ListOnboardingsResponse, error) {
	params := ListOnboardingsParams{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
		Status:   req.Status,
		SortBy:   req.SortBy,
		SortDesc: req.SortDesc,
	}

	onboardings, total, err := s.store.ListOnboardings(params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list onboardings: %v", err)
	}

	pbOnboardings := make([]*pb.Onboarding, 0, len(onboardings))
	for _, onboarding := range onboardings {
		pbOnboardings = append(pbOnboardings, toPbOnboarding(&onboarding))
	}

	return &pb.ListOnboardingsResponse{
		Onboardings: pbOnboardings,
		Total:       total,
	}, nil
}

// Helper function to convert domain model to protobuf model
func toPbOnboarding(onboarding *Onboarding) *pb.Onboarding {
	return &pb.Onboarding{
		Id:            onboarding.ID.String(),
		Timestamp:     onboarding.Timestamp,
		Status:        onboarding.Status,
		Packages:      onboarding.Packages,
		Products:      onboarding.Products,
		NumberOfUsers: int32(onboarding.NumberOfUsers),
		Remarks:       onboarding.Remarks,
		ContactEmail:  onboarding.ContactEmail,
		ContactPhone:  onboarding.ContactPhone,
		Name:          onboarding.Name,
		Metadata:      onboarding.Metadata,
		CreatedAt:     onboarding.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     onboarding.UpdatedAt.Format(time.RFC3339),
	}
}
