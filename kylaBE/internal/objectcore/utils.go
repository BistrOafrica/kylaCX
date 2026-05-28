package objectcore

import (
	"encoding/json"

	"kyla-be/pkg/pb"
)

// ── ObjectType ─────────────────────────────────────────────────────────────────

func ObjectTypeToPb(ot *ObjectType) *pb.ObjectType {
	return &pb.ObjectType{
		Id:          ot.ID,
		OrgId:       ot.OrgID,
		WorkspaceId: ot.WorkspaceID,
		Slug:        ot.Slug,
		Name:        ot.Name,
		PluralName:  ot.PluralName,
		Icon:        ot.Icon,
		Color:       ot.Color,
		IsSystem:    ot.IsSystem,
		Schema:      SchemaToPb(ot.Schema),
		CreatedAt:   ot.CreatedAt.String(),
		UpdatedAt:   ot.UpdatedAt.String(),
	}
}

func SchemaToPb(s ObjectSchema) *pb.ObjectSchema {
	fields := make([]*pb.FieldDefinition, len(s.Fields))
	for i, f := range s.Fields {
		fields[i] = FieldDefToPb(f)
	}
	return &pb.ObjectSchema{Fields: fields}
}

func FieldDefToPb(f FieldDefinition) *pb.FieldDefinition {
	opts := make([]*pb.SelectOption, len(f.Options))
	for i, o := range f.Options {
		opts[i] = &pb.SelectOption{Value: o.Value, Label: o.Label, Color: o.Color}
	}
	return &pb.FieldDefinition{
		Key:        f.Key,
		Label:      f.Label,
		Type:       pb.FieldType(f.Type.toPbFieldType()),
		Required:   f.Required,
		Unique:     f.Unique,
		Searchable: f.Searchable,
		Options:    opts,
		RelatesTo:  f.RelatesTo,
	}
}

func PbToSchema(ps *pb.ObjectSchema) ObjectSchema {
	if ps == nil {
		return ObjectSchema{}
	}
	fields := make([]FieldDefinition, len(ps.Fields))
	for i, f := range ps.Fields {
		fields[i] = PbToFieldDef(f)
	}
	return ObjectSchema{Fields: fields}
}

func PbToFieldDef(f *pb.FieldDefinition) FieldDefinition {
	opts := make([]SelectOption, len(f.Options))
	for i, o := range f.Options {
		opts[i] = SelectOption{Value: o.Value, Label: o.Label, Color: o.Color}
	}
	return FieldDefinition{
		Key:        f.Key,
		Label:      f.Label,
		Type:       fieldTypeFromPb(f.Type),
		Required:   f.Required,
		Unique:     f.Unique,
		Searchable: f.Searchable,
		Options:    opts,
		RelatesTo:  f.RelatesTo,
	}
}

// ── Object record ──────────────────────────────────────────────────────────────

func ObjectToPb(o *Object) *pb.Object {
	return &pb.Object{
		Id:          o.ID,
		OrgId:       o.OrgID,
		WorkspaceId: o.WorkspaceID,
		TypeSlug:    o.TypeSlug,
		Data:        string(o.Data),
		CreatedBy:   o.CreatedBy,
		CreatedAt:   o.CreatedAt.String(),
		UpdatedAt:   o.UpdatedAt.String(),
	}
}

func ObjectsToPb(objs []*Object) []*pb.Object {
	out := make([]*pb.Object, len(objs))
	for i, o := range objs {
		out[i] = ObjectToPb(o)
	}
	return out
}

// ── ObjectRelation ─────────────────────────────────────────────────────────────

func RelationToPb(r *ObjectRelation) *pb.ObjectRelation {
	return &pb.ObjectRelation{
		Id:        r.ID,
		OrgId:     r.OrgID,
		FromId:    r.FromID,
		ToId:      r.ToID,
		Relation:  r.Relation,
		CreatedAt: r.CreatedAt.String(),
	}
}

// ── ObjectEvent ────────────────────────────────────────────────────────────────

func EventToPb(e *ObjectEvent) *pb.ObjectEvent {
	return &pb.ObjectEvent{
		Id:        e.ID,
		OrgId:     e.OrgID,
		ObjectId:  e.ObjectID,
		ActorId:   e.ActorID,
		ActorType: e.ActorType,
		EventType: e.EventType,
		Payload:   string(e.Payload),
		CreatedAt: e.CreatedAt.String(),
	}
}

func EventsToPb(evts []*ObjectEvent) []*pb.ObjectEvent {
	out := make([]*pb.ObjectEvent, len(evts))
	for i, e := range evts {
		out[i] = EventToPb(e)
	}
	return out
}

// ── SavedView ──────────────────────────────────────────────────────────────────

func ViewToPb(v *SavedView) *pb.SavedView {
	return &pb.SavedView{
		Id:          v.ID,
		WorkspaceId: v.WorkspaceID,
		OrgId:       v.OrgID,
		Name:        v.Name,
		TypeSlug:    v.TypeSlug,
		Filters:     string(v.Filters),
		Sort:        string(v.Sort),
		Columns:     string(v.Columns),
		IsShared:    v.IsShared,
		CreatedBy:   v.CreatedBy,
		CreatedAt:   v.CreatedAt.String(),
		UpdatedAt:   v.UpdatedAt.String(),
	}
}

func PbToView(p *pb.SavedView) *SavedView {
	filters := json.RawMessage("[]")
	if p.Filters != "" {
		filters = json.RawMessage(p.Filters)
	}
	sort := json.RawMessage("{}")
	if p.Sort != "" {
		sort = json.RawMessage(p.Sort)
	}
	cols := json.RawMessage("[]")
	if p.Columns != "" {
		cols = json.RawMessage(p.Columns)
	}
	return &SavedView{
		ID:          p.Id,
		WorkspaceID: p.WorkspaceId,
		OrgID:       p.OrgId,
		Name:        p.Name,
		TypeSlug:    p.TypeSlug,
		Filters:     filters,
		Sort:        sort,
		Columns:     cols,
		IsShared:    p.IsShared,
		CreatedBy:   p.CreatedBy,
	}
}

func ViewsToPb(views []*SavedView) []*pb.SavedView {
	out := make([]*pb.SavedView, len(views))
	for i, v := range views {
		out[i] = ViewToPb(v)
	}
	return out
}

// ── FieldType mapping ──────────────────────────────────────────────────────────

func (ft FieldType) toPbFieldType() int32 {
	switch ft {
	case FieldTypeNumber:
		return int32(pb.FieldType_FIELD_TYPE_NUMBER)
	case FieldTypeDate:
		return int32(pb.FieldType_FIELD_TYPE_DATE)
	case FieldTypeDateTime:
		return int32(pb.FieldType_FIELD_TYPE_DATETIME)
	case FieldTypeSelect:
		return int32(pb.FieldType_FIELD_TYPE_SELECT)
	case FieldTypeMulti:
		return int32(pb.FieldType_FIELD_TYPE_MULTI)
	case FieldTypeBoolean:
		return int32(pb.FieldType_FIELD_TYPE_BOOLEAN)
	case FieldTypeUser:
		return int32(pb.FieldType_FIELD_TYPE_USER)
	case FieldTypeRelation:
		return int32(pb.FieldType_FIELD_TYPE_RELATION)
	case FieldTypeFile:
		return int32(pb.FieldType_FIELD_TYPE_FILE)
	case FieldTypeEmail:
		return int32(pb.FieldType_FIELD_TYPE_EMAIL)
	case FieldTypePhone:
		return int32(pb.FieldType_FIELD_TYPE_PHONE)
	case FieldTypeURL:
		return int32(pb.FieldType_FIELD_TYPE_URL)
	case FieldTypeCurrency:
		return int32(pb.FieldType_FIELD_TYPE_CURRENCY)
	default: // FieldTypeText
		return int32(pb.FieldType_FIELD_TYPE_TEXT)
	}
}

func fieldTypeFromPb(ft pb.FieldType) FieldType {
	switch ft {
	case pb.FieldType_FIELD_TYPE_NUMBER:
		return FieldTypeNumber
	case pb.FieldType_FIELD_TYPE_DATE:
		return FieldTypeDate
	case pb.FieldType_FIELD_TYPE_DATETIME:
		return FieldTypeDateTime
	case pb.FieldType_FIELD_TYPE_SELECT:
		return FieldTypeSelect
	case pb.FieldType_FIELD_TYPE_MULTI:
		return FieldTypeMulti
	case pb.FieldType_FIELD_TYPE_BOOLEAN:
		return FieldTypeBoolean
	case pb.FieldType_FIELD_TYPE_USER:
		return FieldTypeUser
	case pb.FieldType_FIELD_TYPE_RELATION:
		return FieldTypeRelation
	case pb.FieldType_FIELD_TYPE_FILE:
		return FieldTypeFile
	case pb.FieldType_FIELD_TYPE_EMAIL:
		return FieldTypeEmail
	case pb.FieldType_FIELD_TYPE_PHONE:
		return FieldTypePhone
	case pb.FieldType_FIELD_TYPE_URL:
		return FieldTypeURL
	case pb.FieldType_FIELD_TYPE_CURRENCY:
		return FieldTypeCurrency
	default:
		return FieldTypeText
	}
}
