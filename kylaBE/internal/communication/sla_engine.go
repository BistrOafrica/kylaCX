package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"kyla-be/shared/events"
)

// SLAEngine manages SLA timers and breach detection.
type SLAEngine struct {
	store     *SLAStore
	convStore *ConversationStore
	eventBus  events.Publisher
	interval  time.Duration
}

// NewSLAEngine constructs an SLAEngine.
func NewSLAEngine(
	store *SLAStore,
	convStore *ConversationStore,
	eventBus events.Publisher,
	scanIntervalSecs int,
) *SLAEngine {
	interval := time.Duration(scanIntervalSecs) * time.Second
	if scanIntervalSecs == 0 {
		interval = 60 * time.Second
	}
	return &SLAEngine{
		store:     store,
		convStore: convStore,
		eventBus:  eventBus,
		interval:  interval,
	}
}

// StartTimer creates an SLA record for a new conversation.
func (e *SLAEngine) StartTimer(conv *Conversation) error {
	// Find matching policy
	policies, err := e.store.FindActivePolicies(conv.OrgID, conv.WorkspaceID)
	if err != nil {
		return fmt.Errorf("load SLA policies: %w", err)
	}

	var selectedPolicy *SLAPolicy
	for _, p := range policies {
		if e.matchesPolicy(p, conv) {
			selectedPolicy = p
			break
		}
	}

	if selectedPolicy == nil {
		// Try default policy
		for _, p := range policies {
			if p.IsDefault {
				selectedPolicy = p
				break
			}
		}
	}

	if selectedPolicy == nil {
		return nil // No policy applies
	}

	// Parse metrics
	var metrics struct {
		FirstResponseHours float64 `json:"first_response_hours"`
		ResolutionHours    float64 `json:"resolution_hours"`
	}
	if err := json.Unmarshal(selectedPolicy.Metrics, &metrics); err != nil {
		return fmt.Errorf("parse policy metrics: %w", err)
	}

	now := time.Now()
	var frDeadline, resDeadline *time.Time

	if metrics.FirstResponseHours > 0 {
		t := now.Add(time.Duration(metrics.FirstResponseHours * float64(time.Hour)))
		frDeadline = &t
	}
	if metrics.ResolutionHours > 0 {
		t := now.Add(time.Duration(metrics.ResolutionHours * float64(time.Hour)))
		resDeadline = &t
	}

	record := &SLARecord{
		ConversationID:        conv.ID,
		PolicyID:              selectedPolicy.ID,
		OrgID:                 conv.OrgID,
		StartedAt:             now,
		FirstResponseDeadline: frDeadline,
		ResolutionDeadline:    resDeadline,
	}

	if _, err := e.store.CreateRecord(record); err != nil {
		return fmt.Errorf("create SLA record: %w", err)
	}

	// Update conversation.sla_deadline for UI display
	if resDeadline != nil {
		conv.SLADeadline = resDeadline
		_, _ = e.convStore.Update(conv)
	}

	log.Printf("[sla] started timer conv=%s policy=%s", conv.ID, selectedPolicy.ID)
	return nil
}

// RecordFirstResponse marks first response time.
func (e *SLAEngine) RecordFirstResponse(conversationID string) error {
	record, err := e.store.FindRecordByConversationID(conversationID)
	if err != nil {
		return err
	}

	if record.FirstRespondedAt != nil {
		return nil // Already recorded
	}

	now := time.Now()
	record.FirstRespondedAt = &now

	if err := e.store.UpdateRecord(record); err != nil {
		return fmt.Errorf("update SLA record: %w", err)
	}

	log.Printf("[sla] recorded first response conv=%s", conversationID)
	return nil
}

// RecordResolution marks resolution time.
func (e *SLAEngine) RecordResolution(conversationID string) error {
	record, err := e.store.FindRecordByConversationID(conversationID)
	if err != nil {
		return err
	}

	if record.ResolvedAt != nil {
		return nil // Already resolved
	}

	now := time.Now()
	record.ResolvedAt = &now

	if err := e.store.UpdateRecord(record); err != nil {
		return fmt.Errorf("update SLA record: %w", err)
	}

	log.Printf("[sla] recorded resolution conv=%s", conversationID)
	return nil
}

// Start runs the breach scanner goroutine.
func (e *SLAEngine) Start(ctx context.Context) {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	log.Printf("[sla] breach scanner started (interval=%s)", e.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("[sla] breach scanner stopped")
			return
		case <-ticker.C:
			e.scanBreaches(ctx)
		}
	}
}

func (e *SLAEngine) scanBreaches(ctx context.Context) {
	records, err := e.store.FindBreachingRecords()
	if err != nil {
		log.Printf("[sla] find breaching records error: %v", err)
		return
	}

	for _, record := range records {
		e.processBreachedRecord(ctx, record)
	}

	if len(records) > 0 {
		log.Printf("[sla] processed %d breached records", len(records))
	}
}

func (e *SLAEngine) processBreachedRecord(ctx context.Context, record *SLARecord) {
	updated := false

	// Check first response breach
	if record.FirstResponseDeadline != nil &&
		record.FirstRespondedAt == nil &&
		!record.FirstResponseBreached &&
		time.Now().After(*record.FirstResponseDeadline) {

		record.FirstResponseBreached = true
		updated = true

		e.publishBreachEvent(record.OrgID, record.ConversationID, record.PolicyID, "first_response")
		log.Printf("[sla] first_response breach conv=%s", record.ConversationID)
	}

	// Check resolution breach
	if record.ResolutionDeadline != nil &&
		record.ResolvedAt == nil &&
		!record.ResolutionBreached &&
		time.Now().After(*record.ResolutionDeadline) {

		record.ResolutionBreached = true
		updated = true

		e.publishBreachEvent(record.OrgID, record.ConversationID, record.PolicyID, "resolution")
		log.Printf("[sla] resolution breach conv=%s", record.ConversationID)
	}

	if updated {
		_ = e.store.UpdateRecord(record)
	}
}

func (e *SLAEngine) publishBreachEvent(orgID, conversationID, policyID, breachType string) {
	payload := map[string]interface{}{
		"conversation_id": conversationID,
		"policy_id":       policyID,
		"breach_type":     breachType,
	}

	ev, err := events.NewEvent(orgID, "", "sla", "breaching", conversationID, "system", payload)
	if err != nil {
		log.Printf("[sla] build breach event error: %v", err)
		return
	}

	if err := e.eventBus.Publish(ev); err != nil {
		log.Printf("[sla] publish breach event error: %v", err)
	}
}

func (e *SLAEngine) matchesPolicy(policy *SLAPolicy, conv *Conversation) bool {
	var conditions []struct {
		Field    string      `json:"field"`
		Operator string      `json:"op"`
		Value    interface{} `json:"value"`
	}

	if err := json.Unmarshal(policy.Conditions, &conditions); err != nil {
		return false
	}

	for _, cond := range conditions {
		if !e.evaluateCondition(cond.Field, cond.Operator, cond.Value, conv) {
			return false
		}
	}

	return true
}

func (e *SLAEngine) evaluateCondition(field, op string, value interface{}, conv *Conversation) bool {
	var actual string

	switch field {
	case "channel":
		actual = conv.Channel
	case "priority":
		actual = conv.Priority
	default:
		return false
	}

	expected, ok := value.(string)
	if !ok {
		return false
	}

	switch op {
	case "eq", "equals":
		return actual == expected
	default:
		return false
	}
}
