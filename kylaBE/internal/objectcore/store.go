package objectcore

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ObjectCoreStore is the database layer for ObjectType, Object, ObjectRelation,
// and ObjectEvent tables.  All writes append an ObjectEvent for timeline tracking.
type ObjectCoreStore struct {
	db *gorm.DB
}

// NewObjectCoreStore constructs an ObjectCoreStore backed by the given DB.
func NewObjectCoreStore(db *gorm.DB) *ObjectCoreStore {
	return &ObjectCoreStore{db: db}
}

// ── ObjectType ────────────────────────────────────────────────────────────────

// CreateObjectType persists a new ObjectType.
func (s *ObjectCoreStore) CreateObjectType(ot *ObjectType) (*ObjectType, error) {
	if ot.ID == "" {
		ot.ID = uuid.New().String()
	}
	now := time.Now()
	ot.CreatedAt = now
	ot.UpdatedAt = now
	if err := s.db.Create(ot).Error; err != nil {
		return nil, err
	}
	return ot, nil
}

// FindObjectTypeBySlug returns an ObjectType by org + slug.
func (s *ObjectCoreStore) FindObjectTypeBySlug(orgID, slug string) (*ObjectType, error) {
	var ot ObjectType
	if err := s.db.Where("org_id = ? AND slug = ?", orgID, slug).First(&ot).Error; err != nil {
		return nil, err
	}
	return &ot, nil
}

// FindObjectByDataField looks up the first Object whose JSONB data column contains
// data->>'key' = value. Returns nil, nil when not found.
func (s *ObjectCoreStore) FindObjectByDataField(orgID, workspaceID, typeSlug, key, value string) (*Object, error) {
	var obj Object
	q := s.db.Where("org_id = ? AND type_slug = ?", orgID, typeSlug).
		Where("data->>? = ?", key, value)
	if workspaceID != "" {
		q = q.Where("workspace_id = ?", workspaceID)
	}
	err := q.First(&obj).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

// ListObjectTypes returns all object types for an org, optionally filtered by workspace.
func (s *ObjectCoreStore) ListObjectTypes(orgID, workspaceID string) ([]*ObjectType, error) {
	var types []*ObjectType
	q := s.db.Where("org_id = ?", orgID)
	if workspaceID != "" {
		q = q.Where("workspace_id = ? OR workspace_id IS NULL OR workspace_id = ''", workspaceID)
	}
	if err := q.Order("name ASC").Find(&types).Error; err != nil {
		return nil, err
	}
	return types, nil
}

// UpdateObjectSchema replaces the schema on an existing ObjectType.
func (s *ObjectCoreStore) UpdateObjectSchema(orgID, slug string, schema ObjectSchema) (*ObjectType, error) {
	ot, err := s.FindObjectTypeBySlug(orgID, slug)
	if err != nil {
		return nil, err
	}
	ot.Schema = schema
	ot.UpdatedAt = time.Now()
	if err := s.db.Model(ot).Where("org_id = ? AND slug = ?", orgID, slug).
		Updates(map[string]interface{}{"schema": ot.Schema, "updated_at": ot.UpdatedAt}).Error; err != nil {
		return nil, err
	}
	return ot, nil
}

// DeleteObjectType removes a non-system ObjectType.
func (s *ObjectCoreStore) DeleteObjectType(orgID, slug string) error {
	return s.db.Where("org_id = ? AND slug = ? AND is_system = false", orgID, slug).
		Delete(&ObjectType{}).Error
}

// ── Object CRUD ───────────────────────────────────────────────────────────────

// CreateObject persists a new Object record and appends a "created" timeline event.
func (s *ObjectCoreStore) CreateObject(obj *Object, actorID string) (*Object, error) {
	if obj.ID == "" {
		obj.ID = uuid.New().String()
	}
	now := time.Now()
	obj.CreatedAt = now
	obj.UpdatedAt = now

	return obj, s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(obj).Error; err != nil {
			return err
		}
		return tx.Create(&ObjectEvent{
			ID:        uuid.New().String(),
			OrgID:     obj.OrgID,
			ObjectID:  obj.ID,
			ActorID:   actorID,
			ActorType: "user",
			EventType: "created",
			Payload:   obj.Data,
			CreatedAt: now,
		}).Error
	})
}

// FindObjectByID returns an Object by primary key within an org.
func (s *ObjectCoreStore) FindObjectByID(id, orgID string) (*Object, error) {
	var obj Object
	if err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&obj).Error; err != nil {
		return nil, err
	}
	return &obj, nil
}

// ListObjectsParams bundles the options for ListObjects.
type ListObjectsParams struct {
	OrgID       string
	WorkspaceID string
	TypeSlug    string
	PageSize    int
	PageToken   string // last seen ID (keyset pagination)
	Filter      string // JSON filter expression
	SortBy      string
	SortDesc    bool
}

// ListObjects returns a paginated list of Object records.
// Filter is a flat JSON map of field=value pairs applied to the JSONB data column.
func (s *ObjectCoreStore) ListObjects(p ListObjectsParams) ([]*Object, string, int64, error) {
	if p.PageSize <= 0 || p.PageSize > 200 {
		p.PageSize = 50
	}

	q := s.db.Model(&Object{}).Where("org_id = ?", p.OrgID)
	if p.WorkspaceID != "" {
		q = q.Where("workspace_id = ?", p.WorkspaceID)
	}
	if p.TypeSlug != "" {
		q = q.Where("type_slug = ?", p.TypeSlug)
	}

	// Apply flat JSON filter: {"status": "open"} → data->>'status' = 'open'
	if p.Filter != "" {
		var filterMap map[string]interface{}
		if err := json.Unmarshal([]byte(p.Filter), &filterMap); err == nil {
			for k, v := range filterMap {
				q = q.Where("data->>? = ?", k, fmt.Sprintf("%v", v))
			}
		}
	}

	// Count total before pagination
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, "", 0, err
	}

	// Keyset pagination: WHERE id > page_token
	if p.PageToken != "" {
		q = q.Where("id > ?", p.PageToken)
	}

	// Sort
	sortCol := "created_at"
	if p.SortBy != "" {
		// Sort on JSONB field if present
		sortCol = fmt.Sprintf("data->>%q", p.SortBy)
	}
	dir := "ASC"
	if p.SortDesc {
		dir = "DESC"
	}
	q = q.Order(fmt.Sprintf("%s %s", sortCol, dir)).Limit(p.PageSize + 1)

	var objs []*Object
	if err := q.Find(&objs).Error; err != nil {
		return nil, "", total, err
	}

	nextToken := ""
	if len(objs) > p.PageSize {
		nextToken = objs[p.PageSize-1].ID
		objs = objs[:p.PageSize]
	}
	return objs, nextToken, total, nil
}

// SearchObjects performs a full-text search across JSONB data using PostgreSQL's
// to_tsvector on the JSON text representation. Results are limited to page_size.
func (s *ObjectCoreStore) SearchObjects(orgID, workspaceID, typeSlug, query string, pageSize int) ([]*Object, error) {
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 50
	}
	q := s.db.Where("org_id = ?", orgID)
	if workspaceID != "" {
		q = q.Where("workspace_id = ?", workspaceID)
	}
	if typeSlug != "" {
		q = q.Where("type_slug = ?", typeSlug)
	}
	// Full-text search over JSON text
	q = q.Where("to_tsvector('english', data::text) @@ websearch_to_tsquery('english', ?)", query).
		Limit(pageSize)
	var objs []*Object
	if err := q.Find(&objs).Error; err != nil {
		return nil, err
	}
	return objs, nil
}

// UpdateObject merges the provided data patch into an existing object and appends
// an "updated" timeline event.
func (s *ObjectCoreStore) UpdateObject(id, orgID, actorID string, patch json.RawMessage) (*Object, error) {
	obj, err := s.FindObjectByID(id, orgID)
	if err != nil {
		return nil, err
	}

	// Merge existing data with the patch
	existing := make(map[string]interface{})
	incoming := make(map[string]interface{})
	_ = json.Unmarshal(obj.Data, &existing)
	if err := json.Unmarshal(patch, &incoming); err != nil {
		return nil, fmt.Errorf("invalid data patch: %w", err)
	}
	for k, v := range incoming {
		existing[k] = v
	}
	merged, _ := json.Marshal(existing)
	obj.Data = merged
	obj.UpdatedAt = time.Now()

	return obj, s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Object{}).Where("id = ?", id).
			Updates(map[string]interface{}{"data": string(merged), "updated_at": obj.UpdatedAt}).Error; err != nil {
			return err
		}
		return tx.Create(&ObjectEvent{
			ID:        uuid.New().String(),
			OrgID:     orgID,
			ObjectID:  id,
			ActorID:   actorID,
			ActorType: "user",
			EventType: "updated",
			Payload:   patch,
			CreatedAt: obj.UpdatedAt,
		}).Error
	})
}

// DeleteObject hard-deletes an Object (cascades to relations and events via FK).
func (s *ObjectCoreStore) DeleteObject(id, orgID, actorID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Append a deleted event before the cascade removes the record.
		tx.Create(&ObjectEvent{
			ID:        uuid.New().String(),
			OrgID:     orgID,
			ObjectID:  id,
			ActorID:   actorID,
			ActorType: "user",
			EventType: "deleted",
			Payload:   json.RawMessage(`{}`),
			CreatedAt: time.Now(),
		})
		return tx.Where("id = ? AND org_id = ?", id, orgID).Delete(&Object{}).Error
	})
}

// ── Relations ─────────────────────────────────────────────────────────────────

// LinkObjects creates a named relation between two objects. Idempotent via ON CONFLICT DO NOTHING.
func (s *ObjectCoreStore) LinkObjects(orgID, fromID, toID, relation string) (*ObjectRelation, error) {
	rel := &ObjectRelation{
		ID:        uuid.New().String(),
		OrgID:     orgID,
		FromID:    fromID,
		ToID:      toID,
		Relation:  relation,
		CreatedAt: time.Now(),
	}
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(rel).Error; err != nil {
		return nil, err
	}
	return rel, nil
}

// UnlinkObjects removes a named relation between two objects.
func (s *ObjectCoreStore) UnlinkObjects(fromID, toID, relation string) error {
	return s.db.Where("from_id = ? AND to_id = ? AND relation = ?", fromID, toID, relation).
		Delete(&ObjectRelation{}).Error
}

// GetObjectRelations returns all relations for an object, optionally filtered by relation name.
func (s *ObjectCoreStore) GetObjectRelations(orgID, objectID, relation string) ([]*ObjectRelation, error) {
	q := s.db.Where("org_id = ? AND (from_id = ? OR to_id = ?)", orgID, objectID, objectID)
	if relation != "" {
		q = q.Where("relation = ?", relation)
	}
	var rels []*ObjectRelation
	if err := q.Order("created_at DESC").Find(&rels).Error; err != nil {
		return nil, err
	}
	return rels, nil
}

// ── Timeline ──────────────────────────────────────────────────────────────────

// AppendEvent manually appends an ObjectEvent (e.g., a comment or status change).
func (s *ObjectCoreStore) AppendEvent(evt *ObjectEvent) error {
	if evt.ID == "" {
		evt.ID = uuid.New().String()
	}
	evt.CreatedAt = time.Now()
	return s.db.Create(evt).Error
}

// GetObjectTimeline returns the most recent timeline events for an object.
func (s *ObjectCoreStore) GetObjectTimeline(orgID, objectID string, limit int) ([]*ObjectEvent, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	var evts []*ObjectEvent
	if err := s.db.Where("org_id = ? AND object_id = ?", orgID, objectID).
		Order("created_at DESC").Limit(limit).Find(&evts).Error; err != nil {
		return nil, err
	}
	return evts, nil
}

// ── System type seeding ───────────────────────────────────────────────────────

// SeedSystemObjectTypes creates all system ObjectTypes for the given workspace template.
// Implements workspace.SystemTypeSeedable.
func (s *ObjectCoreStore) SeedSystemObjectTypes(orgID, workspaceID, template string) error {
	types := systemObjectTypes(orgID, workspaceID, template)
	for _, ot := range types {
		ot := ot // capture
		existing, _ := s.FindObjectTypeBySlug(orgID, ot.Slug)
		if existing != nil {
			continue // already seeded (idempotent)
		}
		if _, err := s.CreateObjectType(&ot); err != nil {
			return fmt.Errorf("seed system type %q: %w", ot.Slug, err)
		}
	}
	return nil
}
