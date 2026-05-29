package campaigns

import (
	"encoding/json"
	"time"

	"kyla-be/pkg/pb"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ── Campaign ─────────────────────────────────────────────────────────────────

// CampaignToProto serialises a Campaign DB row to its proto form.
// Audience / Schedule / Payload are pulled out of their JSONB raw bytes via
// structpb so the wire format matches what the frontend constructs.
func CampaignToProto(c *Campaign) *pb.Campaign {
	if c == nil {
		return nil
	}
	out := &pb.Campaign{
		Id:          c.ID,
		OrgId:       c.OrgID,
		WorkspaceId: c.WorkspaceID,
		Name:        c.Name,
		Description: c.Description,
		Channel:     c.Channel,
		Status:      c.Status,
		Audience:    audienceProto(c.Audience),
		Schedule:    scheduleProto(c.Schedule),
		Payload:     rawToStruct(c.Payload),
		Stats: &pb.CampaignStats{
			AudienceSize: int64(c.AudienceSize),
			Queued:       int64(c.QueuedCount),
			Sent:         int64(c.SentCount),
			Delivered:    int64(c.DeliveredCount),
			Read:         int64(c.ReadCount),
			Failed:       int64(c.FailedCount),
		},
		CreatedBy: c.CreatedBy,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
	return out
}

// CampaignFromProto applies writable fields from a proto onto a Campaign model.
// IDs, counters, and Temporal IDs are deliberately not copied — those live
// under server / workflow control, not client input.
func CampaignFromProto(p *pb.Campaign) *Campaign {
	if p == nil {
		return nil
	}
	c := &Campaign{
		ID:          p.GetId(),
		OrgID:       p.GetOrgId(),
		WorkspaceID: p.GetWorkspaceId(),
		Name:        p.GetName(),
		Description: p.GetDescription(),
		Channel:     p.GetChannel(),
		Status:      p.GetStatus(),
		Audience:    audienceToRaw(p.GetAudience()),
		Payload:     structToRaw(p.GetPayload()),
	}
	if p.GetSchedule() != nil {
		c.Schedule = scheduleToRaw(p.GetSchedule())
	}
	if c.Status == "" {
		c.Status = string(StatusDraft)
	}
	return c
}

// ── Audience / Schedule helpers ──────────────────────────────────────────────

// DecodeAudience parses Campaign.Audience JSONB into the typed struct.
// Workflow code uses this to drive resolution.
func DecodeAudience(raw json.RawMessage) (CampaignAudience, error) {
	var a CampaignAudience
	if len(raw) == 0 {
		return a, nil
	}
	err := json.Unmarshal(raw, &a)
	return a, err
}

// DecodeSchedule parses Campaign.Schedule JSONB into the typed struct.
func DecodeSchedule(raw json.RawMessage) (CampaignSchedule, error) {
	var s CampaignSchedule
	if len(raw) == 0 {
		return s, nil
	}
	err := json.Unmarshal(raw, &s)
	return s, err
}

func scheduleProto(raw json.RawMessage) *pb.CampaignSchedule {
	s, err := DecodeSchedule(raw)
	if err != nil {
		return &pb.CampaignSchedule{}
	}
	out := &pb.CampaignSchedule{
		Mode:     string(s.Mode),
		Cron:     s.Cron,
		Timezone: s.Timezone,
	}
	if s.StartAt != nil {
		out.StartAt = timestamppb.New(*s.StartAt)
	}
	return out
}

func scheduleToRaw(p *pb.CampaignSchedule) json.RawMessage {
	if p == nil {
		return json.RawMessage(`{}`)
	}
	s := CampaignSchedule{
		Mode:     ScheduleMode(p.GetMode()),
		Cron:     p.GetCron(),
		Timezone: p.GetTimezone(),
	}
	if p.GetStartAt() != nil {
		t := p.GetStartAt().AsTime()
		s.StartAt = &t
	}
	b, _ := json.Marshal(s)
	return b
}

// audienceProto serialises CampaignAudience JSONB into its proto shape.
// Returns an empty (non-nil) message for empty input so the wire format is
// always populated.
func audienceProto(raw json.RawMessage) *pb.CampaignAudience {
	a, err := DecodeAudience(raw)
	if err != nil {
		return &pb.CampaignAudience{}
	}
	out := &pb.CampaignAudience{
		Kind:      string(a.Kind),
		TypeSlug:  a.TypeSlug,
		ObjectIds: a.ObjectIDs,
	}
	if len(a.Filter) > 0 {
		if s, err := structpb.NewStruct(a.Filter); err == nil {
			out.Filter = s
		}
	}
	return out
}

// audienceToRaw inverts audienceProto for the write path.
func audienceToRaw(p *pb.CampaignAudience) json.RawMessage {
	if p == nil {
		return json.RawMessage(`{}`)
	}
	a := CampaignAudience{
		Kind:      AudienceKind(p.GetKind()),
		TypeSlug:  p.GetTypeSlug(),
		ObjectIDs: p.GetObjectIds(),
	}
	if p.GetFilter() != nil {
		a.Filter = p.GetFilter().AsMap()
	}
	b, _ := json.Marshal(a)
	return b
}

// ── structpb <-> json.RawMessage ────────────────────────────────────────────

// rawToStruct converts a JSONB raw message to structpb for wire transport.
// Empty or invalid input returns an empty struct rather than nil so callers
// don't have to nil-check downstream.
func rawToStruct(raw json.RawMessage) *structpb.Struct {
	if len(raw) == 0 {
		return &structpb.Struct{}
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return &structpb.Struct{}
	}
	s, err := structpb.NewStruct(m)
	if err != nil {
		return &structpb.Struct{}
	}
	return s
}

func structToRaw(s *structpb.Struct) json.RawMessage {
	if s == nil {
		return json.RawMessage(`{}`)
	}
	b, err := json.Marshal(s.AsMap())
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}

// ── Recipient ────────────────────────────────────────────────────────────────

func RecipientToProto(r *CampaignRecipient) *pb.CampaignRecipient {
	if r == nil {
		return nil
	}
	out := &pb.CampaignRecipient{
		Id:         r.ID,
		CampaignId: r.CampaignID,
		OrgId:      r.OrgID,
		ObjectId:   r.ObjectID,
		ContactRef: r.ContactRef,
		Status:     r.Status,
		ExternalId: r.ExternalID,
		Error:      r.Error,
	}
	if r.SentAt != nil {
		out.SentAt = timestamppb.New(*r.SentAt)
	}
	if r.DeliveredAt != nil {
		out.DeliveredAt = timestamppb.New(*r.DeliveredAt)
	}
	if r.ReadAt != nil {
		out.ReadAt = timestamppb.New(*r.ReadAt)
	}
	return out
}

// ── WhatsApp templates ──────────────────────────────────────────────────────

func TemplateToProto(t *WhatsAppTemplate) *pb.WhatsAppTemplate {
	if t == nil {
		return nil
	}
	return &pb.WhatsAppTemplate{
		Id:             t.ID,
		OrgId:          t.OrgID,
		Name:           t.Name,
		Language:       t.Language,
		Category:       t.Category,
		Status:         t.Status,
		Header:         t.Header,
		Body:           t.Body,
		Footer:         t.Footer,
		PhoneNumberId:  t.PhoneNumberID,
		WabaId:         t.WabaID,
		MetaTemplateId: t.MetaTemplateID,
		CreatedAt:      timestamppb.New(t.CreatedAt),
		UpdatedAt:      timestamppb.New(t.UpdatedAt),
	}
}

func TemplateFromProto(p *pb.WhatsAppTemplate) *WhatsAppTemplate {
	if p == nil {
		return nil
	}
	return &WhatsAppTemplate{
		ID:             p.GetId(),
		OrgID:          p.GetOrgId(),
		Name:           p.GetName(),
		Language:       p.GetLanguage(),
		Category:       p.GetCategory(),
		Status:         p.GetStatus(),
		Header:         p.GetHeader(),
		Body:           p.GetBody(),
		Footer:         p.GetFooter(),
		PhoneNumberID:  p.GetPhoneNumberId(),
		WabaID:         p.GetWabaId(),
		MetaTemplateID: p.GetMetaTemplateId(),
	}
}

// ScheduleStartTime returns the effective send start for a schedule.
// "immediate" → now; "scheduled_once" → start_at; "recurring" → next cron tick
// (the caller computes that against Temporal Schedules — see workflow code).
func ScheduleStartTime(s CampaignSchedule) time.Time {
	now := time.Now().UTC()
	switch s.Mode {
	case ScheduleOnce:
		if s.StartAt != nil {
			return *s.StartAt
		}
		return now
	default:
		return now
	}
}
