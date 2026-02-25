package service

import (
	"context"
	"fmt"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Groups struct {
	pb.UnimplementedGroupServiceServer
	groupStore   *ContactGroupStore
	AuthStore    *AuthStore
	ContactStore *ContactStore
}

func NewContactGroupServer(groupStore *ContactGroupStore, authStore *AuthStore, ContactStore *ContactStore) *Groups {
	return &Groups{groupStore: groupStore, AuthStore: authStore, ContactStore: ContactStore}
}

func (g *Groups) CreateGroup(ctx context.Context, request *pb.CreateGroupRequest) (*pb.CreateGroupResponse, error) {
	log.Println("Create contact group")
	var contextData *RequestMetadata
	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(&pb.Scope{
		OwnerType: request.GetGroup().GetOwnerType(),
		OwnerId:   request.GetGroup().GetOwnerId(),
	})

	go g.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		if !CheckOpScope(contextChanData, scope) {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		contextData = contextChanData
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)

	}

	scopeID, err := uuid.Parse(scope.ID)
	if err != nil {
		return nil, fmt.Errorf("error while parsing scope id: %v", err)
	}
	group := PbContactGroupToContactGroup(request.GetGroup())
	group.ID = uuid.New()
	group.OwnerType = scope.Owner
	group.OwnerID = scopeID
	group.CreatedBy = contextData.UserID.String()

	if err := g.groupStore.Save(group); err != nil {
		return nil, fmt.Errorf("error while creating group: %v", err)
	}

	return &pb.CreateGroupResponse{
		Group: ContactGroupToPbContactGroup(group),
	}, nil
}

func (g *Groups) ReadGroup(ctx context.Context, request *pb.ReadGroupRequest) (*pb.ReadGroupResponse, error) {
	log.Println("Read contact group")

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go g.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)

	var contextChanData *RequestMetadata

	select {
	case contextChanData = <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	id := request.GetId()
	group, err := g.groupStore.FindByID(id, contextChanData)
	if err != nil {
		return nil, fmt.Errorf("error while reading group: %v", err)
	}

	return &pb.ReadGroupResponse{
		Group: ContactGroupToPbContactGroup(group),
	}, nil
}

func (g *Groups) ReadGroups(ctx context.Context, request *pb.ReadGroupsRequest) (*pb.ReadGroupsResponse, error) {
	log.Println("Read contact groups")

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go g.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to access this resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)

	}

	contextData, authErr := g.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "Forbidden: You don't have permission to read contact groups")
	}

	contextDataForGroups, err := g.AuthStore.AuthInternalRequests(contextData.Authorization, "READ_GROUPS")
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}

	idsAllowingAccess := GetScopeIDs(contextDataForGroups.Scopes)

	page := request.GetPage()
	if page == 0 {
		page = 1
	}

	perPage := request.GetPerPage()
	if perPage == 0 {
		perPage = 10
	}

	groups, count, err := g.groupStore.FindAll(idsAllowingAccess, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("error while reading groups: %v", err)
	}

	pbGroups := make([]*pb.Group, len(groups))
	for i, group := range groups {
		pbGroups[i] = ContactGroupToPbContactGroup(group)
	}

	return &pb.ReadGroupsResponse{
		Groups: pbGroups,
		Count:  count,
	}, nil

}

func (g *Groups) UpdateGroup(ctx context.Context, request *pb.UpdateGroupRequest) (*pb.UpdateGroupResponse, error) {
	log.Println("Update contact group")

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	scope := PbScopeToOpScope(&pb.Scope{
		OwnerType: request.GetGroup().GetOwnerType(),
		OwnerId:   request.GetGroup().GetOwnerId(),
	})

	go g.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		if !CheckOpScope(contextChanData, scope) {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)

	}

	group := PbContactGroupToContactGroup(request.GetGroup())

	updatedGroup, err := g.groupStore.Update(group, scope)

	if err != nil {
		return nil, fmt.Errorf("error while updating group: %v", err)
	}

	return &pb.UpdateGroupResponse{
		Group: ContactGroupToPbContactGroup(updatedGroup),
	}, nil
}

func (g *Groups) DeleteGroup(ctx context.Context, request *pb.DeleteGroupRequest) (*pb.DeleteGroupResponse, error) {
	log.Println("Delete contact group")

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)
	if request.GetScope() == nil {
		return nil, status.Error(codes.InvalidArgument, "Scope is required")
	}

	scope := PbScopeToOpScope(request.GetScope())

	go g.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
		if !CheckOpScope(contextChanData, scope) {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to get user resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)
	}

	id := request.GetId()

	if err := g.groupStore.Delete(id, scope); err != nil {
		return nil, fmt.Errorf("error while deleting group: %v", err)
	}

	return &pb.DeleteGroupResponse{
		Success: true,
	}, nil
}

func (g *Groups) ReadGroupContacts(ctx context.Context, request *pb.ReadGroupContactsRequest) (*pb.ReadGroupContactsResponse, error) {
	log.Println("Read group contacts")

	contextDataChan := make(chan *RequestMetadata)
	errChan := make(chan error)

	go g.AuthStore.GetUserRequestMetadata(ctx, contextDataChan, errChan)
	select {
	case contextChanData := <-contextDataChan:
		if contextChanData.RequestAuth == k.NewConsts().FALSE {
			return nil, status.Error(codes.PermissionDenied, "Forbidden, You do not have access to access this resource")
		}
	case err := <-errChan:
		return nil, status.Errorf(codes.Unknown, "error while fetching user %v", err)

	}

	contextData, authErr := g.AuthStore.GetServiceAuthMetadata(ctx)
	if authErr != nil || contextData.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(codes.PermissionDenied, "Forbidden: You don't have permission to read group's contacts")
	}

	contextDataForGroups, err := g.AuthStore.AuthInternalRequests(contextData.Authorization, "READ_GROUP_CONTACTS")
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}

	groupId := request.GetGroupId()

	group, err := g.groupStore.FindByID(groupId, contextDataForGroups)
	if err != nil {
		return nil, fmt.Errorf("error while getting group: %v", err)
	}

	if !CheckIfIDInScope(contextDataForGroups.Scopes, group.OwnerID.String()) {
		return nil, status.Error(codes.PermissionDenied, "Forbidden: You don't have permission to read group's contacts")
	}

	page := request.GetPage()
	if page == 0 {
		page = 1
	}

	perPage := request.GetPerPage()
	if perPage == 0 {
		perPage = 10
	}

	contacts, count, err := g.groupStore.ReadGroupContacts(groupId, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("error while getting group contacts: %v", err)
	}

	pbContacts := make([]*pb.Contact, 0, len(contacts))
	for _, contact := range contacts {
		fields, definitions, err := g.getContactCustomFields(contact.ID.String())
		if err != nil {
			contact.CustomFieldValues = []CustomFieldValue{}
		}
		pbContacts = append(pbContacts, ContactToPbContact(contact, definitions, fields))
	}

	return &pb.ReadGroupContactsResponse{
		Contacts: pbContacts,
		Count:    count,
	}, nil
}

func (g *Groups) getContactCustomFields(contactId string) (fields []*CustomFieldValue, definitions []*CustomFieldDefinition, err error) {
	customFields, err := g.ContactStore.FindCustomFieldByContactID(contactId)
	if err != nil {
		return nil, nil, err
	}
	var defList []*CustomFieldDefinition
	for _, field := range customFields {
		definition, err := g.ContactStore.FindCustomFieldDefinitionByID(field.CustomFieldDefinitionID)
		if err != nil {
			return nil, nil, err
		}
		defList = append(defList, definition)
	}
	return customFields, defList, nil
}
