package service

import (
	"context"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

type SharingServer struct {
	pb.UnimplementedResourceSharingServer
	SharingStore *SharingStore
	AuthStore    *AuthStore
}

func NewSharingServer(sharingStore *SharingStore, authStore *AuthStore) *SharingServer {
	return &SharingServer{SharingStore: sharingStore, AuthStore: authStore}
}

func (s *SharingServer) ShareResource(ctx context.Context, req *pb.ShareResourceRequest) (*pb.ShareResourceResponse, error) {
	resourceID, err := uuid.Parse(req.ResourceId)
	if err != nil {
		return nil, status.Errorf(500, "invalid request")
	}
	fromOwnerId, err := uuid.Parse(req.FromOwnerId)
	if err != nil {
		return nil, status.Errorf(500, "invalid request")
	}
	toEntityId, err := uuid.Parse(req.ToEntityId)
	if err != nil {
		return nil, status.Errorf(500, "invalid request")
	}

	er := s.SharingStore.ShareResource(resourceID, fromOwnerId, toEntityId, req.Roles, req.Permissions, req.Conditions)
	if er != nil {
		return nil, err
	}
	return &pb.ShareResourceResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Resource Shared successfully",
		},
	}, nil
}

func (s *SharingServer) RequestAccess(ctx context.Context, req *pb.RequestAccessRequest) (*pb.RequestAccessResponse, error) {
	resourceId, err := uuid.Parse(req.ResourceId)
	if err != nil {
		return nil, status.Errorf(500, "invalid resource")
	}
	requesterId, err := uuid.Parse(req.RequesterId)
	if err != nil {
		return nil, status.Errorf(500, "invalid request")
	}
	resourceOwner, err := uuid.Parse(req.ResourceOwner)
	if err != nil {
		return nil, status.Errorf(500, "invalid request")
	}
	if ok := s.SharingStore.HasOwnership(resourceOwner.String(), resourceId.String()); !ok {
		return nil, status.Errorf(401, "access denied by owner")
	}
	er := s.SharingStore.RequestAccess(resourceOwner, resourceId, requesterId, req.RequestedRoles)
	if er != nil {
		return nil, err
	}
	return &pb.RequestAccessResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Resource Shared successfully",
		},
	}, nil
}

func (s *SharingServer) GetRequests(ctx context.Context, req *pb.AccessRequestsRequest) (*pb.AccessRequestResponse, error) {
	log.Println("Get Requests")
	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	go s.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(403, "Forbidden, You do not have access to get resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(500, "error while fetching user %v", err)
	}

	requests, err := s.SharingStore.GetRequests(contextData.UserID)
	if err != nil {
		return nil, err
	}
	return &pb.AccessRequestResponse{
		Requests: AccessRequestsToPbList(*requests),
	}, nil
}

func (s *SharingServer) GrantAccess(ctx context.Context, req *pb.GrantAccessRequest) (*pb.GrantAccessResponse, error) {
	requestID, err := uuid.Parse(req.RequestId)
	if err != nil {
		return nil, status.Errorf(500, "invalid request")
	}
	granter, err := uuid.Parse(req.GranterId)
	if err != nil {
		return nil, status.Errorf(500, "invalid request")
	}
	er := s.SharingStore.GrantAccess(requestID, granter, req.Permissions)
	if er != nil {
		return &pb.GrantAccessResponse{
			Status: &pb.Status{
				Code:    200,
				Message: er.Error(),
			},
		}, err
	}
	return &pb.GrantAccessResponse{
		Status: &pb.Status{
			Code:    200,
			Message: "Resource access granted successfully",
		},
	}, nil
}

func (s *SharingServer) GetResources(ctx context.Context, req *pb.GetResourcesRequest) (*pb.GetResourcesResponse, error) {
	resources, err := s.SharingStore.GetResources(req.OwnerId)
	if err != nil {
		return nil, err
	}
	return &pb.GetResourcesResponse{Resources: EntityLinksToPbEntityLinks(resources)}, nil
}

func (s *SharingServer) GetOwners(ctx context.Context, req *pb.GetOwnersRequest) (*pb.GetOwnersResponse, error) {
	owners, err := s.SharingStore.GetOwners(req.ResourceId)
	if err != nil {
		return nil, err
	}
	return &pb.GetOwnersResponse{Owners: EntityLinksToPbEntityLinks(owners)}, nil
}
