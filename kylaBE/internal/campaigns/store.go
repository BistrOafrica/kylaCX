package campaigns

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Store wraps DB access for campaigns, recipients, and the WhatsApp template
// mirror. It deliberately does NOT call into Temporal — the gRPC server and
// workflow code own that.
type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

// ── Campaigns ────────────────────────────────────────────────────────────────

func (s *Store) CreateCampaign(c *Campaign) (*Campaign, error) {
	if err := s.db.Create(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Store) GetCampaign(id, orgID string) (*Campaign, error) {
	var c Campaign
	err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&c).Error
	return &c, err
}

type ListCampaignsParams struct {
	OrgID       string
	WorkspaceID string
	Status      string
	Channel     string
	PageSize    int
	PageToken   string
}

// ListCampaigns returns campaigns matching the filter, page-token-paginated by
// CreatedAt DESC then ID DESC (keyset). page_token is the last seen
// "createdAtUnix:id" pair, decoded by the server layer.
func (s *Store) ListCampaigns(p ListCampaignsParams) ([]*Campaign, int64, error) {
	q := s.db.Model(&Campaign{}).Where("org_id = ?", p.OrgID)
	if p.WorkspaceID != "" {
		q = q.Where("workspace_id = ?", p.WorkspaceID)
	}
	if p.Status != "" {
		q = q.Where("status = ?", p.Status)
	}
	if p.Channel != "" {
		q = q.Where("channel = ?", p.Channel)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if p.PageSize <= 0 || p.PageSize > 200 {
		p.PageSize = 50
	}
	var out []*Campaign
	if err := q.Order("created_at DESC, id DESC").Limit(p.PageSize).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// UpdateCampaign saves all writable fields. Counter columns and Temporal IDs
// are owned by the workflow path — callers should use the dedicated helpers
// below to mutate those, not this generic Save.
func (s *Store) UpdateCampaign(c *Campaign) (*Campaign, error) {
	c.UpdatedAt = time.Now().UTC()
	if err := s.db.Model(&Campaign{}).Where("id = ? AND org_id = ?", c.ID, c.OrgID).Updates(map[string]interface{}{
		"name":        c.Name,
		"description": c.Description,
		"channel":     c.Channel,
		"audience":    c.Audience,
		"schedule":    c.Schedule,
		"payload":     c.Payload,
		"updated_at":  c.UpdatedAt,
	}).Error; err != nil {
		return nil, err
	}
	return s.GetCampaign(c.ID, c.OrgID)
}

func (s *Store) DeleteCampaign(id, orgID string) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&Campaign{}).Error
}

// SetStatus transitions a campaign to a new status. Returns the refreshed row.
func (s *Store) SetStatus(id, orgID string, status CampaignStatus) (*Campaign, error) {
	if err := s.db.Model(&Campaign{}).Where("id = ? AND org_id = ?", id, orgID).Updates(map[string]interface{}{
		"status":     string(status),
		"updated_at": time.Now().UTC(),
	}).Error; err != nil {
		return nil, err
	}
	return s.GetCampaign(id, orgID)
}

// SetTemporalIDs stores the Temporal workflow and (optionally) schedule IDs
// the gRPC server obtained at launch time. Looking them up later powers
// pause / cancel and links the UI into Temporal Web.
func (s *Store) SetTemporalIDs(id, orgID, workflowID, scheduleID string) error {
	return s.db.Model(&Campaign{}).Where("id = ? AND org_id = ?", id, orgID).Updates(map[string]interface{}{
		"temporal_workflow_id": workflowID,
		"temporal_schedule_id": scheduleID,
		"updated_at":           time.Now().UTC(),
	}).Error
}

// ── Recipients ───────────────────────────────────────────────────────────────

// BulkInsertRecipients writes a resolved audience in chunks. Used by the
// CampaignExecutionWorkflow's resolve step. Idempotent via ON CONFLICT
// (campaign_id, object_id) so a workflow retry doesn't duplicate rows.
func (s *Store) BulkInsertRecipients(rows []*CampaignRecipient) error {
	if len(rows) == 0 {
		return nil
	}
	const batch = 500
	for i := 0; i < len(rows); i += batch {
		end := i + batch
		if end > len(rows) {
			end = len(rows)
		}
		if err := s.db.Exec(`
			INSERT INTO campaign_recipients (campaign_id, org_id, object_id, contact_ref, status)
			SELECT * FROM unnest(?::uuid[], ?::uuid[], ?::uuid[], ?::text[], ?::text[])
			ON CONFLICT (campaign_id, object_id) DO NOTHING`,
			extractStrings(rows[i:end], func(r *CampaignRecipient) string { return r.CampaignID }),
			extractStrings(rows[i:end], func(r *CampaignRecipient) string { return r.OrgID }),
			extractStrings(rows[i:end], func(r *CampaignRecipient) string { return r.ObjectID }),
			extractStrings(rows[i:end], func(r *CampaignRecipient) string { return r.ContactRef }),
			extractStrings(rows[i:end], func(r *CampaignRecipient) string {
				if r.Status == "" {
					return string(RecipientQueued)
				}
				return r.Status
			}),
		).Error; err != nil {
			return fmt.Errorf("bulk insert recipients chunk [%d:%d]: %w", i, end, err)
		}
	}
	return nil
}

// ListQueuedRecipients returns recipients still pending send for a campaign.
// Used by the fan-out workflow step.
func (s *Store) ListQueuedRecipients(campaignID string, limit int) ([]*CampaignRecipient, error) {
	if limit <= 0 || limit > 5000 {
		limit = 1000
	}
	var rows []*CampaignRecipient
	err := s.db.Where("campaign_id = ? AND status = ?", campaignID, string(RecipientQueued)).
		Order("created_at ASC").Limit(limit).Find(&rows).Error
	return rows, err
}

type ListRecipientsParams struct {
	CampaignID string
	Status     string
	PageSize   int
	PageToken  string
}

func (s *Store) ListRecipients(p ListRecipientsParams) ([]*CampaignRecipient, int64, error) {
	q := s.db.Model(&CampaignRecipient{}).Where("campaign_id = ?", p.CampaignID)
	if p.Status != "" {
		q = q.Where("status = ?", p.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if p.PageSize <= 0 || p.PageSize > 500 {
		p.PageSize = 100
	}
	var out []*CampaignRecipient
	if err := q.Order("created_at ASC, id ASC").Limit(p.PageSize).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// MarkRecipientSent updates a recipient to sent with the provider's external
// message ID so subsequent delivery/read webhooks can find this row.
func (s *Store) MarkRecipientSent(recipientID, externalID string) error {
	now := time.Now().UTC()
	return s.db.Model(&CampaignRecipient{}).Where("id = ?", recipientID).Updates(map[string]interface{}{
		"status":      string(RecipientSent),
		"external_id": externalID,
		"sent_at":     &now,
		"updated_at":  now,
	}).Error
}

// MarkRecipientFailed records a non-retryable failure for a recipient.
func (s *Store) MarkRecipientFailed(recipientID, errMsg string) error {
	now := time.Now().UTC()
	return s.db.Model(&CampaignRecipient{}).Where("id = ?", recipientID).Updates(map[string]interface{}{
		"status":     string(RecipientFailed),
		"error":      errMsg,
		"updated_at": now,
	}).Error
}

// MarkRecipientStatusByExternalID promotes status when a provider delivery /
// read webhook arrives. Called from the channel webhook handlers, not the
// workflow.
func (s *Store) MarkRecipientStatusByExternalID(externalID string, status RecipientStatus) error {
	now := time.Now().UTC()
	updates := map[string]interface{}{
		"status":     string(status),
		"updated_at": now,
	}
	switch status {
	case RecipientDelivered:
		updates["delivered_at"] = &now
	case RecipientRead:
		updates["read_at"] = &now
	}
	return s.db.Model(&CampaignRecipient{}).Where("external_id = ?", externalID).Updates(updates).Error
}

// RecomputeStats rebuilds the denormalised counters on the campaign row from
// the per-recipient table. Called at workflow milestones and on demand.
func (s *Store) RecomputeStats(campaignID string) error {
	var row struct {
		Total     int64
		Queued    int64
		Sent      int64
		Delivered int64
		Read      int64
		Failed    int64
	}
	if err := s.db.Raw(`
		SELECT
		  COUNT(*) AS total,
		  COUNT(*) FILTER (WHERE status='queued')    AS queued,
		  COUNT(*) FILTER (WHERE status='sent')      AS sent,
		  COUNT(*) FILTER (WHERE status='delivered') AS delivered,
		  COUNT(*) FILTER (WHERE status='read')      AS read,
		  COUNT(*) FILTER (WHERE status='failed')    AS failed
		FROM campaign_recipients WHERE campaign_id = ?`, campaignID).Scan(&row).Error; err != nil {
		return err
	}
	return s.db.Model(&Campaign{}).Where("id = ?", campaignID).Updates(map[string]interface{}{
		"audience_size":   row.Total,
		"queued_count":    row.Queued,
		"sent_count":      row.Sent,
		"delivered_count": row.Delivered,
		"read_count":      row.Read,
		"failed_count":    row.Failed,
		"updated_at":      time.Now().UTC(),
	}).Error
}

// ── WhatsApp templates ──────────────────────────────────────────────────────

// UpsertTemplate inserts or updates a template by (org_id, name, language).
func (s *Store) UpsertTemplate(t *WhatsAppTemplate) (*WhatsAppTemplate, error) {
	now := time.Now().UTC()
	t.UpdatedAt = now
	if err := s.db.Exec(`
		INSERT INTO whatsapp_templates
		  (id, org_id, name, language, category, status, header, body, footer,
		   phone_number_id, waba_id, meta_template_id, created_at, updated_at)
		VALUES (gen_random_uuid(), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (org_id, name, language) DO UPDATE SET
		  category         = EXCLUDED.category,
		  status           = EXCLUDED.status,
		  header           = EXCLUDED.header,
		  body             = EXCLUDED.body,
		  footer           = EXCLUDED.footer,
		  phone_number_id  = EXCLUDED.phone_number_id,
		  waba_id          = EXCLUDED.waba_id,
		  meta_template_id = EXCLUDED.meta_template_id,
		  updated_at       = EXCLUDED.updated_at`,
		t.OrgID, t.Name, t.Language, t.Category, t.Status,
		t.Header, t.Body, t.Footer,
		t.PhoneNumberID, t.WabaID, t.MetaTemplateID,
		now, now,
	).Error; err != nil {
		return nil, err
	}
	return s.GetTemplateByName(t.OrgID, t.Name, t.Language)
}

func (s *Store) GetTemplateByName(orgID, name, language string) (*WhatsAppTemplate, error) {
	var t WhatsAppTemplate
	err := s.db.Where("org_id = ? AND name = ? AND language = ?", orgID, name, language).First(&t).Error
	return &t, err
}

func (s *Store) GetTemplate(id, orgID string) (*WhatsAppTemplate, error) {
	var t WhatsAppTemplate
	err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&t).Error
	return &t, err
}

func (s *Store) ListTemplates(orgID, status, phoneNumberID string) ([]*WhatsAppTemplate, error) {
	q := s.db.Where("org_id = ?", orgID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if phoneNumberID != "" {
		q = q.Where("phone_number_id = ?", phoneNumberID)
	}
	var out []*WhatsAppTemplate
	err := q.Order("name ASC").Find(&out).Error
	return out, err
}

func (s *Store) DeleteTemplate(id, orgID string) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&WhatsAppTemplate{}).Error
}

// ── helpers ──────────────────────────────────────────────────────────────────

func extractStrings(rows []*CampaignRecipient, get func(*CampaignRecipient) string) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = get(r)
	}
	return out
}
