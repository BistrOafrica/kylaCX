package telephony

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Store wraps DB access for calls, call events, and SIP infrastructure.
type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

// ── Calls ────────────────────────────────────────────────────────────────────

func (s *Store) CreateCall(c *Call) (*Call, error) {
	if c.ID == "" {
		return nil, errors.New("call.id is required (must be PBX-assigned UUID)")
	}
	if err := s.db.Create(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Store) GetCall(id, orgID string) (*Call, error) {
	var c Call
	err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&c).Error
	return &c, err
}

// GetCallByIDOnly fetches without org scoping — used by ESL event handlers
// where the caller is trusted (PBX events have no auth context).
func (s *Store) GetCallByIDOnly(id string) (*Call, error) {
	var c Call
	err := s.db.Where("id = ?", id).First(&c).Error
	return &c, err
}

type ListCallsParams struct {
	OrgID       string
	WorkspaceID string
	Direction   string
	Status      string
	AgentID     string
	PageSize    int
	PageToken   string
}

func (s *Store) ListCalls(p ListCallsParams) ([]*Call, int64, error) {
	q := s.db.Model(&Call{}).Where("org_id = ?", p.OrgID)
	if p.WorkspaceID != "" {
		q = q.Where("workspace_id = ?", p.WorkspaceID)
	}
	if p.Direction != "" {
		q = q.Where("direction = ?", p.Direction)
	}
	if p.Status != "" {
		q = q.Where("status = ?", p.Status)
	}
	if p.AgentID != "" {
		q = q.Where("agent_id = ?", p.AgentID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if p.PageSize <= 0 || p.PageSize > 500 {
		p.PageSize = 100
	}
	var out []*Call
	if err := q.Order("started_at DESC, id DESC").Limit(p.PageSize).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// MarkAnswered transitions a call to "answered" and records the answer time.
// Called from the ESL CHANNEL_ANSWER handler.
func (s *Store) MarkAnswered(id string, answeredAt time.Time) error {
	return s.db.Model(&Call{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      string(StatusAnswered),
		"answered_at": &answeredAt,
		"updated_at":  time.Now().UTC(),
	}).Error
}

// MarkEnded finalises a call. ringSeconds is derived from started_at→answered_at
// at the call site (the PBX provides better numbers than we can compute from
// timestamps alone, particularly for queued calls).
func (s *Store) MarkEnded(id, hangupCause, disposition string, endedAt time.Time, ringSeconds, talkSeconds int) error {
	updates := map[string]interface{}{
		"status":       string(StatusEnded),
		"ended_at":     &endedAt,
		"hangup_cause": hangupCause,
		"disposition":  disposition,
		"ring_seconds": ringSeconds,
		"talk_seconds": talkSeconds,
		"updated_at":   time.Now().UTC(),
	}
	return s.db.Model(&Call{}).Where("id = ?", id).Updates(updates).Error
}

// SetRecordingURL attaches the recording URL post-hangup.
func (s *Store) SetRecordingURL(id, url string) error {
	return s.db.Model(&Call{}).Where("id = ?", id).Updates(map[string]interface{}{
		"recording_url": url,
		"updated_at":    time.Now().UTC(),
	}).Error
}

// SetTranscriptStatus updates the call-level transcript status flag. Used by
// the pipeline to surface progress without waiting for completion.
func (s *Store) SetTranscriptStatus(id, status, errMsg string) error {
	return s.db.Model(&Call{}).Where("id = ?", id).Updates(map[string]interface{}{
		"transcript_status": status,
		"transcript_error":  errMsg,
		"updated_at":        time.Now().UTC(),
	}).Error
}

// SetTranscript stores the final transcript (concatenated across all
// recordings) on the call row.
func (s *Store) SetTranscript(id, transcript, provider string) error {
	return s.db.Model(&Call{}).Where("id = ?", id).Updates(map[string]interface{}{
		"transcript":          transcript,
		"transcript_status":   string(TranscribeDone),
		"transcript_provider": provider,
		"transcript_error":    "",
		"updated_at":          time.Now().UTC(),
	}).Error
}

// ── Recordings ──────────────────────────────────────────────────────────────

func (s *Store) CreateRecording(r *CallRecording) (*CallRecording, error) {
	if err := s.db.Create(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Store) GetRecording(id string) (*CallRecording, error) {
	var r CallRecording
	err := s.db.Where("id = ?", id).First(&r).Error
	return &r, err
}

// ListRecordingsForCall returns all recordings belonging to a call, ordered
// by recorded_at so concatenated transcripts stay in temporal order.
func (s *Store) ListRecordingsForCall(callID string) ([]*CallRecording, error) {
	var out []*CallRecording
	err := s.db.Where("call_id = ?", callID).Order("recorded_at ASC").Find(&out).Error
	return out, err
}

// SetRecordingUploaded transitions the recording to uploaded with the
// destination URI.
func (s *Store) SetRecordingUploaded(id, bucket, key, url string, size int64) error {
	now := time.Now().UTC()
	return s.db.Model(&CallRecording{}).Where("id = ?", id).Updates(map[string]interface{}{
		"upload_status": string(UploadUploaded),
		"upload_error":  "",
		"s3_bucket":     bucket,
		"s3_key":        key,
		"s3_url":        url,
		"size_bytes":    size,
		"uploaded_at":   &now,
		"updated_at":    now,
	}).Error
}

// SetRecordingUploadFailed records a non-retryable upload error.
func (s *Store) SetRecordingUploadFailed(id, errMsg string) error {
	return s.db.Model(&CallRecording{}).Where("id = ?", id).Updates(map[string]interface{}{
		"upload_status": string(UploadFailed),
		"upload_error":  errMsg,
		"updated_at":    time.Now().UTC(),
	}).Error
}

// SetRecordingTranscribed stores the per-recording transcript and provider.
func (s *Store) SetRecordingTranscribed(id, transcript, provider string) error {
	now := time.Now().UTC()
	return s.db.Model(&CallRecording{}).Where("id = ?", id).Updates(map[string]interface{}{
		"transcribe_status": string(TranscribeDone),
		"transcribe_error":  "",
		"transcript":        transcript,
		"transcribed_by":    provider,
		"transcribed_at":    &now,
		"updated_at":        now,
	}).Error
}

func (s *Store) SetRecordingTranscribeFailed(id, errMsg string) error {
	return s.db.Model(&CallRecording{}).Where("id = ?", id).Updates(map[string]interface{}{
		"transcribe_status": string(TranscribeFailed),
		"transcribe_error":  errMsg,
		"updated_at":        time.Now().UTC(),
	}).Error
}

// LinkConversation stores the conversation_id created by the VoiceCallBridge
// post-call. Idempotent — repeated calls overwrite.
func (s *Store) LinkConversation(callID, conversationID string) error {
	return s.db.Model(&Call{}).Where("id = ?", callID).Updates(map[string]interface{}{
		"conversation_id": conversationID,
		"updated_at":      time.Now().UTC(),
	}).Error
}

// ── Call events ──────────────────────────────────────────────────────────────

func (s *Store) AppendEvent(e *CallEvent) (*CallEvent, error) {
	if err := s.db.Create(e).Error; err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Store) ListEvents(callID string) ([]*CallEvent, error) {
	var out []*CallEvent
	err := s.db.Where("call_id = ?", callID).Order("at ASC").Find(&out).Error
	return out, err
}

// ── SIP domains ─────────────────────────────────────────────────────────────

func (s *Store) CreateDomain(d *SipDomain) (*SipDomain, error) {
	if err := s.db.Create(d).Error; err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Store) GetDefaultDomain(orgID string) (*SipDomain, error) {
	var d SipDomain
	err := s.db.Where("org_id = ? AND is_default = ?", orgID, true).First(&d).Error
	return &d, err
}

func (s *Store) ListDomains(orgID string) ([]*SipDomain, error) {
	var out []*SipDomain
	err := s.db.Where("org_id = ?", orgID).Order("domain ASC").Find(&out).Error
	return out, err
}

func (s *Store) DeleteDomain(id, orgID string) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&SipDomain{}).Error
}

// ── SIP extensions ──────────────────────────────────────────────────────────

func (s *Store) CreateExtension(e *SipExtension) (*SipExtension, error) {
	if err := s.db.Create(e).Error; err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Store) GetExtension(id, orgID string) (*SipExtension, error) {
	var e SipExtension
	err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&e).Error
	return &e, err
}

// GetExtensionByUserID resolves the extension owned by a user — used at
// originate time to find the right SIP identity to bridge.
func (s *Store) GetExtensionByUserID(userID string) (*SipExtension, error) {
	var e SipExtension
	err := s.db.Where("user_id = ?", userID).First(&e).Error
	return &e, err
}

// GetExtensionByNumber resolves a registration request by extension number.
// Used by mod_xml_curl's directory handler — FreeSWITCH supplies the dialled
// user as `key_value`; we serve back the A1 hash + variables.
func (s *Store) GetExtensionByNumber(extension string) (*SipExtension, error) {
	var e SipExtension
	err := s.db.Where("extension = ?", extension).First(&e).Error
	return &e, err
}

func (s *Store) ListExtensions(workspaceID string) ([]*SipExtension, error) {
	var out []*SipExtension
	err := s.db.Where("workspace_id = ?", workspaceID).Order("extension ASC").Find(&out).Error
	return out, err
}

// MarkRegistered records a successful registration event from FreeSWITCH.
// Called from the ESL handler when a SOFIA_REGISTER event arrives.
func (s *Store) MarkRegistered(extension, orgID string, at time.Time) error {
	return s.db.Model(&SipExtension{}).
		Where("extension = ? AND org_id = ?", extension, orgID).
		Updates(map[string]interface{}{
			"status":            "registered",
			"last_registration": &at,
			"updated_at":        time.Now().UTC(),
		}).Error
}

func (s *Store) DeleteExtension(id, orgID string) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&SipExtension{}).Error
}

// ── SIP trunks ──────────────────────────────────────────────────────────────

func (s *Store) CreateTrunk(t *SipTrunk) (*SipTrunk, error) {
	if err := s.db.Create(t).Error; err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Store) GetTrunk(id, orgID string) (*SipTrunk, error) {
	var t SipTrunk
	err := s.db.Where("id = ? AND org_id = ?", id, orgID).First(&t).Error
	return &t, err
}

func (s *Store) GetTrunkByName(orgID, name string) (*SipTrunk, error) {
	var t SipTrunk
	err := s.db.Where("org_id = ? AND name = ?", orgID, name).First(&t).Error
	return &t, err
}

func (s *Store) ListTrunks(orgID string) ([]*SipTrunk, error) {
	var out []*SipTrunk
	err := s.db.Where("org_id = ?", orgID).Order("name ASC").Find(&out).Error
	return out, err
}

// ListAllActiveTrunks returns active trunks across ALL orgs. Used by
// mod_xml_curl's configuration handler to serve a sofia.conf populated with
// every tenant's outbound gateway. Per-tenant FreeSWITCH instances are a
// follow-up — today one PBX handles all orgs.
func (s *Store) ListAllActiveTrunks() ([]*SipTrunk, error) {
	var out []*SipTrunk
	err := s.db.Where("is_active = ?", true).Order("org_id ASC, name ASC").Find(&out).Error
	return out, err
}

// UpdateTrunk saves writable fields. Password is overwritten only if a
// non-empty value is supplied — empty means "keep existing".
func (s *Store) UpdateTrunk(t *SipTrunk) (*SipTrunk, error) {
	updates := map[string]interface{}{
		"name":         t.Name,
		"gateway_name": t.GatewayName,
		"provider":     t.Provider,
		"sip_server":   t.SipServer,
		"username":     t.Username,
		"from_uri":     t.FromURI,
		"is_active":    t.IsActive,
		"updated_at":   time.Now().UTC(),
	}
	if t.Password != "" {
		updates["password"] = t.Password
	}
	if err := s.db.Model(&SipTrunk{}).Where("id = ? AND org_id = ?", t.ID, t.OrgID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.GetTrunk(t.ID, t.OrgID)
}

func (s *Store) DeleteTrunk(id, orgID string) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&SipTrunk{}).Error
}
