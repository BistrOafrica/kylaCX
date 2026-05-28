package objectcore

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"kyla-be/internal/authctx"
	"kyla-be/pkg/k"
	"kyla-be/pkg/pb"
	"kyla-be/shared/events"

	"google.golang.org/grpc/status"
)

// AuthGateway is the subset of the auth stack that ObjectCoreServer needs.
type AuthGateway interface {
	GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
	ScopeCheck(ctx context.Context, scopeID string) (bool, *authctx.RequestMetadata, error)
}

// ObjectCoreServer implements pb.ObjectCoreServiceServer.
type ObjectCoreServer struct {
	store    *ObjectCoreStore
	auth     AuthGateway
	eventBus events.Publisher
	pb.UnimplementedObjectCoreServiceServer
}

// NewObjectCoreServer constructs an ObjectCoreServer.
func NewObjectCoreServer(store *ObjectCoreStore, auth AuthGateway, eventBus events.Publisher) *ObjectCoreServer {
	return &ObjectCoreServer{store: store, auth: auth, eventBus: eventBus}
}

// ── ObjectType RPCs ───────────────────────────────────────────────────────────

func (s *ObjectCoreServer) CreateObjectType(ctx context.Context, req *pb.CreateObjectTypeRequest) (*pb.ObjectType, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	ot := req.GetObjectType()
	if ot == nil {
		return nil, status.Error(400, "object_type is required")
	}

	model := &ObjectType{
		OrgID:       ot.GetOrgId(),
		WorkspaceID: ot.GetWorkspaceId(),
		Slug:        ot.GetSlug(),
		Name:        ot.GetName(),
		PluralName:  ot.GetPluralName(),
		Icon:        ot.GetIcon(),
		Color:       ot.GetColor(),
		IsSystem:    ot.GetIsSystem(),
		Schema:      PbToSchema(ot.GetSchema()),
	}

	created, err := s.store.CreateObjectType(model)
	if err != nil {
		return nil, status.Error(500, "failed to create object type")
	}
	return ObjectTypeToPb(created), nil
}

func (s *ObjectCoreServer) GetObjectType(ctx context.Context, req *pb.GetObjectTypeRequest) (*pb.ObjectType, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	ot, err := s.store.FindObjectTypeBySlug(req.GetOrgId(), req.GetSlug())
	if err != nil {
		return nil, status.Error(404, "object type not found")
	}
	return ObjectTypeToPb(ot), nil
}

func (s *ObjectCoreServer) ListObjectTypes(ctx context.Context, req *pb.ListObjectTypesRequest) (*pb.ListObjectTypesResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	types, err := s.store.ListObjectTypes(req.GetOrgId(), req.GetWorkspaceId())
	if err != nil {
		return nil, status.Error(500, "failed to list object types")
	}

	out := make([]*pb.ObjectType, len(types))
	for i, t := range types {
		out[i] = ObjectTypeToPb(t)
	}
	return &pb.ListObjectTypesResponse{ObjectTypes: out}, nil
}

func (s *ObjectCoreServer) UpdateObjectSchema(ctx context.Context, req *pb.UpdateObjectSchemaRequest) (*pb.ObjectType, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	schema := PbToSchema(req.GetSchema())
	updated, err := s.store.UpdateObjectSchema(req.GetOrgId(), req.GetSlug(), schema)
	if err != nil {
		return nil, status.Error(500, "failed to update schema")
	}
	return ObjectTypeToPb(updated), nil
}

func (s *ObjectCoreServer) DeleteObjectType(ctx context.Context, req *pb.DeleteObjectTypeRequest) (*pb.DeleteObjectTypeResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	if err := s.store.DeleteObjectType(req.GetOrgId(), req.GetSlug()); err != nil {
		return nil, status.Error(500, "failed to delete object type")
	}
	return &pb.DeleteObjectTypeResponse{Success: true}, nil
}

// ── Object CRUD RPCs ──────────────────────────────────────────────────────────

func (s *ObjectCoreServer) CreateObject(ctx context.Context, req *pb.CreateObjectRequest) (*pb.Object, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}
	if req.GetOrgId() == "" || req.GetTypeSlug() == "" {
		return nil, status.Error(400, "org_id and type_slug are required")
	}

	dataRaw := json.RawMessage(req.GetData())
	if len(dataRaw) == 0 {
		dataRaw = json.RawMessage("{}")
	}

	obj := &Object{
		OrgID:       req.GetOrgId(),
		WorkspaceID: req.GetWorkspaceId(),
		TypeSlug:    req.GetTypeSlug(),
		Data:        dataRaw,
		CreatedBy:   reqAuth.UserID.String(),
	}

	created, err := s.store.CreateObject(obj, reqAuth.UserID.String())
	if err != nil {
		return nil, status.Error(500, "failed to create object")
	}

	s.publishObjectEvent(created.OrgID, created.WorkspaceID, created.ID, "created", reqAuth.UserID.String(), created.Data)
	return ObjectToPb(created), nil
}

func (s *ObjectCoreServer) GetObject(ctx context.Context, req *pb.GetObjectRequest) (*pb.Object, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	obj, err := s.store.FindObjectByID(req.GetId(), req.GetOrgId())
	if err != nil {
		return nil, status.Error(404, "object not found")
	}
	return ObjectToPb(obj), nil
}

func (s *ObjectCoreServer) ListObjects(ctx context.Context, req *pb.ListObjectsRequest) (*pb.ListObjectsResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	params := ListObjectsParams{
		OrgID:       req.GetOrgId(),
		WorkspaceID: req.GetWorkspaceId(),
		TypeSlug:    req.GetTypeSlug(),
		PageSize:    int(req.GetPageSize()),
		PageToken:   req.GetPageToken(),
		Filter:      req.GetFilter(),
		SortBy:      req.GetSortBy(),
		SortDesc:    req.GetSortDesc(),
	}

	objs, nextToken, total, err := s.store.ListObjects(params)
	if err != nil {
		return nil, status.Error(500, "failed to list objects")
	}

	return &pb.ListObjectsResponse{
		Objects:       ObjectsToPb(objs),
		NextPageToken: nextToken,
		Total:         int32(total),
	}, nil
}

func (s *ObjectCoreServer) SearchObjects(ctx context.Context, req *pb.SearchObjectsRequest) (*pb.SearchObjectsResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	if req.GetQuery() == "" {
		return nil, status.Error(400, "query is required")
	}

	objs, err := s.store.SearchObjects(req.GetOrgId(), req.GetWorkspaceId(), req.GetTypeSlug(), req.GetQuery(), int(req.GetPageSize()))
	if err != nil {
		return nil, status.Error(500, "failed to search objects")
	}
	return &pb.SearchObjectsResponse{Objects: ObjectsToPb(objs)}, nil
}

func (s *ObjectCoreServer) UpdateObject(ctx context.Context, req *pb.UpdateObjectRequest) (*pb.Object, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	dataRaw := json.RawMessage(req.GetData())
	if len(dataRaw) == 0 {
		return nil, status.Error(400, "data patch is required")
	}

	updated, err := s.store.UpdateObject(req.GetId(), req.GetOrgId(), reqAuth.UserID.String(), dataRaw)
	if err != nil {
		return nil, status.Error(500, "failed to update object")
	}

	s.publishObjectEvent(updated.OrgID, updated.WorkspaceID, updated.ID, "updated", reqAuth.UserID.String(), dataRaw)
	return ObjectToPb(updated), nil
}

func (s *ObjectCoreServer) DeleteObject(ctx context.Context, req *pb.DeleteObjectRequest) (*pb.DeleteObjectResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	// Fetch before delete so we have workspaceID for the event.
	obj, fetchErr := s.store.FindObjectByID(req.GetId(), req.GetOrgId())

	if err := s.store.DeleteObject(req.GetId(), req.GetOrgId(), reqAuth.UserID.String()); err != nil {
		return nil, status.Error(500, "failed to delete object")
	}

	if fetchErr == nil {
		s.publishObjectEvent(obj.OrgID, obj.WorkspaceID, obj.ID, "deleted", reqAuth.UserID.String(), json.RawMessage(`{}`))
	}

	return &pb.DeleteObjectResponse{Success: true}, nil
}

// ── Relation RPCs ─────────────────────────────────────────────────────────────

func (s *ObjectCoreServer) LinkObjects(ctx context.Context, req *pb.LinkObjectsRequest) (*pb.ObjectRelation, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	rel, err := s.store.LinkObjects(req.GetOrgId(), req.GetFromId(), req.GetToId(), req.GetRelation())
	if err != nil {
		return nil, status.Error(500, "failed to link objects")
	}
	return RelationToPb(rel), nil
}

func (s *ObjectCoreServer) UnlinkObjects(ctx context.Context, req *pb.UnlinkObjectsRequest) (*pb.UnlinkObjectsResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	if err := s.store.UnlinkObjects(req.GetFromId(), req.GetToId(), req.GetRelation()); err != nil {
		return nil, status.Error(500, "failed to unlink objects")
	}
	return &pb.UnlinkObjectsResponse{Success: true}, nil
}

func (s *ObjectCoreServer) GetObjectRelations(ctx context.Context, req *pb.GetObjectRelationsRequest) (*pb.GetObjectRelationsResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	rels, err := s.store.GetObjectRelations(req.GetOrgId(), req.GetObjectId(), req.GetRelation())
	if err != nil {
		return nil, status.Error(500, "failed to get object relations")
	}

	out := make([]*pb.ObjectRelation, len(rels))
	for i, r := range rels {
		out[i] = RelationToPb(r)
	}
	return &pb.GetObjectRelationsResponse{Relations: out}, nil
}

// ── Timeline RPC ──────────────────────────────────────────────────────────────

func (s *ObjectCoreServer) GetObjectTimeline(ctx context.Context, req *pb.GetObjectTimelineRequest) (*pb.GetObjectTimelineResponse, error) {
	reqAuth, err := s.auth.GetServiceAuthMetadata(ctx)
	if err != nil || reqAuth.RequestAuth != k.NewConsts().TRUE {
		return nil, status.Error(403, "Forbidden")
	}

	evts, err := s.store.GetObjectTimeline(req.GetOrgId(), req.GetObjectId(), int(req.GetLimit()))
	if err != nil {
		return nil, status.Error(500, "failed to get object timeline")
	}
	return &pb.GetObjectTimelineResponse{Events: EventsToPb(evts)}, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (s *ObjectCoreServer) publishObjectEvent(orgID, workspaceID, objectID, action, actorID string, payload json.RawMessage) {
	subject := fmt.Sprintf(subjectForAction(action), orgID)
	ev, err := events.NewEvent(orgID, workspaceID, "object", action, objectID, actorID, payload)
	if err != nil {
		log.Printf("[objectcore] event build error (action=%s obj=%s): %v", action, objectID, err)
		return
	}
	ev.Subject = subject
	if err := s.eventBus.Publish(ev); err != nil {
		log.Printf("[objectcore] event publish error (action=%s obj=%s): %v", action, objectID, err)
	}
}

func subjectForAction(action string) string {
	switch action {
	case "updated":
		return events.SubjectObjectUpdated
	case "deleted":
		return events.SubjectObjectDeleted
	default:
		return events.SubjectObjectCreated
	}
}
