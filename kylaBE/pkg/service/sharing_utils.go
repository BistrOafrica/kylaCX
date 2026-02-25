package service

import (
	"kyla-be/pkg/pb"
	"time"
)

func EntityToPbEntity(e *Entity) *pb.Entity {
	if e == nil {
		return nil
	}

	return &pb.Entity{
		Id:              e.ID.String(),
		Type:            e.Type,
		OwnershipEntity: e.OwnershipEntity,
		CreatedAt:       e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       e.UpdatedAt.Format(time.RFC3339),
		Links:           EntityLinksToPbEntityLinks(e.Resources),
	}
}

func EntityLinkToPbEntityLink(el *EntityLink) *pb.EntityLink {
	if el == nil {
		return nil
	}

	return &pb.EntityLink{
		Id:          el.ID.String(),
		FromType:    el.FromType,
		FromId:      el.FromID.String(),
		ToType:      el.ToType,
		ToId:        el.ToID.String(),
		Type:        string(el.Type),
		Roles:       el.Roles,
		Permissions: el.Permissions,
		SharedBy:    el.SharedBy,
		CreatedAt:   el.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   el.UpdatedAt.Format(time.RFC3339),
		// From:        EntityToPbEntity(&el.FromEntity),
		// To:          EntityToPbEntity(&el.ToEntity),
	}
}

// Convert slices
func EntitiesToPbEntities(entities []Entity) []*pb.Entity {
	pbEntities := make([]*pb.Entity, len(entities))
	for i, e := range entities {
		pbEntities[i] = EntityToPbEntity(&e)
	}
	return pbEntities
}

func EntityLinksToPbEntityLinks(links []EntityLink) []*pb.EntityLink {
	pbLinks := make([]*pb.EntityLink, len(links))
	for i, link := range links {
		pbLinks[i] = EntityLinkToPbEntityLink(&link)
	}
	return pbLinks
}

func AccessRequestToPb(ar *AccessRequest) *pb.AccessRequest {
	if ar == nil {
		return nil
	}

	return &pb.AccessRequest{
		Id:             ar.ID.String(),
		ResourceId:     ar.ResourceID.String(),
		RequesterId:    ar.RequesterID.String(),
		RequestedRoles: ar.RequestedRoles,
		Status:         ar.Status,
		Timestamp:      ar.Timestamp.Format(time.RFC3339),
		CreatedAt:      ar.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      ar.UpdatedAt.Format(time.RFC3339),
	}
}

func AccessRequestsToPbList(accessRequests []AccessRequest) []*pb.AccessRequest {
	pbRequests := make([]*pb.AccessRequest, len(accessRequests))
	for i, ar := range accessRequests {
		pbRequests[i] = AccessRequestToPb(&ar)
	}
	return pbRequests
}
