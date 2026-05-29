package telephony

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"kyla-be/shared/events"
)

// IVRHook is the minimal interface the EventBridge uses to drive the IVR
// executor. Defined here so internal/telephony/ivr can satisfy it without
// the bridge importing the ivr package (avoiding a circular import).
//
// LookupFlowForDID returns the flow + scope to run for an inbound DID, or
// ("", "", "", err) when no mapping exists. Errors are treated as "no IVR".
//
// StartForCall and Advance mirror the IVR executor's signatures.
type IVRHook interface {
	LookupFlowForDID(did string) (flowID, orgID, workspaceID string, err error)
	StartForCall(ctx context.Context, callUUID, flowID, orgID, workspaceID string) (string, error)
	Advance(ctx context.Context, callUUID string, eventType PBXEventType, input string)
}

// QueueHook is the minimal interface for queue runtime callbacks. Same
// circular-import-avoidance pattern as IVRHook — defined here so
// internal/telephony/queues can satisfy it.
//
// OnAgentAnswered fires when an originated agent leg is answered (Router
// bridges to the caller leg).
// OnCallEnded fires for both caller and agent leg hangups (Router cleans
// up the entry and pulls the next waiting caller).
type QueueHook interface {
	OnAgentAnswered(ctx context.Context, agentCallUUID string)
	OnCallEnded(ctx context.Context, callUUID string)
}

// EventBridge consumes PBXEvents from the controller's event stream, updates
// the DB projection (Call rows and CallEvents), and emits domain events on
// NATS so downstream systems (communication.VoiceCallBridge, automation
// consumer, analytics) react.
//
// One bridge per binary — it's the single subscriber to the stream channel.
type EventBridge struct {
	store     *Store
	publisher events.Publisher
	stream    *CallEventStream
	ivr       IVRHook   // optional — nil disables IVR routing
	queues    QueueHook // optional — nil disables queue runtime callbacks
}

func NewEventBridge(store *Store, publisher events.Publisher, stream *CallEventStream) *EventBridge {
	return &EventBridge{store: store, publisher: publisher, stream: stream}
}

// AttachIVR wires an IVR hook into the bridge. Must be called before Start
// for the routing to apply to inbound calls. Optional — the bridge functions
// normally without it; IVR routing simply doesn't happen.
func (b *EventBridge) AttachIVR(hook IVRHook) {
	b.ivr = hook
}

// AttachQueues wires the queues runtime into the bridge. Optional.
func (b *EventBridge) AttachQueues(hook QueueHook) {
	b.queues = hook
}

// Start drains the event stream until ctx is cancelled. Returns when the
// stream is closed or ctx fires — does not run in a goroutine; the caller
// decides whether to spawn one.
func (b *EventBridge) Start(ctx context.Context) {
	if b.stream == nil {
		log.Println("telephony event bridge: nil stream; bridge disabled")
		return
	}
	log.Println("telephony event bridge: started")
	for {
		select {
		case <-ctx.Done():
			log.Println("telephony event bridge: stopping (ctx cancelled)")
			return
		case evt, ok := <-b.stream.Events:
			if !ok {
				log.Println("telephony event bridge: stream closed")
				return
			}
			b.handle(evt)
		}
	}
}

func (b *EventBridge) handle(evt PBXEvent) {
	switch evt.Type {
	case EventChannelCreate:
		b.onCreate(evt)
	case EventChannelAnswer:
		b.onAnswer(evt)
		// Agent legs originated by the queue runtime announce themselves via
		// CHANNEL_ANSWER. The router decides whether the UUID belongs to one
		// of its in-flight dispatches.
		if b.queues != nil && evt.CallUUID != "" {
			b.queues.OnAgentAnswered(context.Background(), evt.CallUUID)
		}
	case EventChannelHangup:
		b.onHangup(evt)
		// Notify the IVR executor so any active run is finalised.
		if b.ivr != nil && evt.CallUUID != "" {
			b.ivr.Advance(context.Background(), evt.CallUUID, EventChannelHangup, "")
		}
		// Same for queues — caller or agent leg ending triggers entry
		// finalisation + pull-next.
		if b.queues != nil && evt.CallUUID != "" {
			b.queues.OnCallEnded(context.Background(), evt.CallUUID)
		}
	case EventSofiaRegister:
		b.onRegister(evt)
	case EventRecordingComplete:
		b.onRecordingComplete(evt)
	case EventPlaybackStop:
		// A play_audio / say node completed. Advance the IVR.
		if b.ivr != nil && evt.CallUUID != "" {
			b.ivr.Advance(context.Background(), evt.CallUUID, EventPlaybackStop, "")
		}
	case EventDTMFCaptured:
		// A menu node returned the captured digit(s).
		if b.ivr != nil && evt.CallUUID != "" {
			digits, _ := evt.Data["captured_digits"].(string)
			b.ivr.Advance(context.Background(), evt.CallUUID, EventDTMFCaptured, digits)
		}
	default:
		// Drop. The PBX emits many event types we don't care about.
	}
}

// ── Handlers ─────────────────────────────────────────────────────────────────

// onCreate handles CHANNEL_CREATE for INBOUND calls — outbound calls already
// have their Call row from OriginateCall. We detect direction from kyla
// variables: outbound calls have kyla_org_id set by Originate, inbound calls
// arrive without it and we treat them as inbound from the start.
func (b *EventBridge) onCreate(evt PBXEvent) {
	if evt.CallUUID == "" {
		return
	}
	if _, err := b.store.GetCallByIDOnly(evt.CallUUID); err == nil {
		// Outbound — row exists, log a "started" event.
		b.appendEvent(evt.CallUUID, "started", evt)
		return
	}
	// Inbound — synthesize a Call row. The PBX provides the from/to numbers
	// in Caller-Caller-ID-Number / Caller-Destination-Number; org/workspace
	// resolution can come from one of two paths:
	//   1. The dialplan injected kyla_org_id / kyla_workspace_id via
	//      mod_xml_curl (set channel variables before reaching ESL).
	//   2. The IVR DID mapping table — we look up by Caller-Destination-Number
	//      and use the org/workspace from there.
	orgID := mapString(evt.Data, "variable_kyla_org_id")
	workspaceID := mapString(evt.Data, "variable_kyla_workspace_id")
	toNumber := mapString(evt.Data, "Caller-Destination-Number")

	// Path 2: try the IVR DID mapping when the dialplan didn't supply context.
	var ivrFlowID string
	if b.ivr != nil && toNumber != "" {
		if flowID, flowOrg, flowWs, err := b.ivr.LookupFlowForDID(toNumber); err == nil {
			ivrFlowID = flowID
			if orgID == "" {
				orgID = flowOrg
			}
			if workspaceID == "" {
				workspaceID = flowWs
			}
		}
	}

	if orgID == "" {
		log.Printf("[telephony bridge] CHANNEL_CREATE without org context (call=%s); skipping persistence", evt.CallUUID)
		return
	}
	call := &Call{
		ID:          evt.CallUUID,
		OrgID:       orgID,
		WorkspaceID: workspaceID,
		Direction:   string(DirectionInbound),
		Status:      string(StatusRinging),
		FromNumber:  mapString(evt.Data, "Caller-Caller-ID-Number"),
		ToNumber:    toNumber,
		IvrFlowID:   ivrFlowID,
		StartedAt:   evt.OccurredAt,
	}
	if _, err := b.store.CreateCall(call); err != nil {
		log.Printf("[telephony bridge] create inbound call %s: %v", evt.CallUUID, err)
		return
	}
	b.appendEvent(evt.CallUUID, "started", evt)
	b.publish(call, "call.started")

	// IVR hand-off: if the DID mapped to an active flow, start the run.
	// The executor handles its own failures gracefully (logs and ends the
	// run); we don't want bridge to block on PBX command latency.
	if b.ivr != nil && ivrFlowID != "" {
		go func() {
			if _, err := b.ivr.StartForCall(context.Background(), evt.CallUUID, ivrFlowID, orgID, workspaceID); err != nil {
				log.Printf("[telephony bridge] IVR start failed for call=%s flow=%s: %v", evt.CallUUID, ivrFlowID, err)
			}
		}()
	}
}

func (b *EventBridge) onAnswer(evt PBXEvent) {
	if evt.CallUUID == "" {
		return
	}
	if err := b.store.MarkAnswered(evt.CallUUID, evt.OccurredAt); err != nil {
		log.Printf("[telephony bridge] mark answered %s: %v", evt.CallUUID, err)
		return
	}
	b.appendEvent(evt.CallUUID, "answered", evt)
	if c, err := b.store.GetCallByIDOnly(evt.CallUUID); err == nil {
		b.publish(c, "call.answered")
	}
}

func (b *EventBridge) onHangup(evt PBXEvent) {
	if evt.CallUUID == "" {
		return
	}
	cause := mapString(evt.Data, "Hangup-Cause")
	dispo := dispositionFromCause(cause)

	c, err := b.store.GetCallByIDOnly(evt.CallUUID)
	if err != nil {
		log.Printf("[telephony bridge] hangup for unknown call %s: %v", evt.CallUUID, err)
		return
	}

	// Prefer PBX-reported durations when present; otherwise compute from our
	// timestamps. FreeSWITCH exposes them as integer seconds in the event.
	ringSecs := mapInt(evt.Data, "variable_progress_uepoch") // first ring -> answer
	talkSecs := mapInt(evt.Data, "variable_billsec")
	if ringSecs == 0 {
		ringSecs = computeRingSeconds(c.StartedAt, c.AnsweredAt, evt.OccurredAt)
	}
	if talkSecs == 0 {
		talkSecs = computeTalkSeconds(c.AnsweredAt, evt.OccurredAt)
	}

	if err := b.store.MarkEnded(evt.CallUUID, cause, dispo, evt.OccurredAt, ringSecs, talkSecs); err != nil {
		log.Printf("[telephony bridge] mark ended %s: %v", evt.CallUUID, err)
		return
	}
	b.appendEvent(evt.CallUUID, "ended", evt)

	if refreshed, err := b.store.GetCallByIDOnly(evt.CallUUID); err == nil {
		// VoiceCallBridge listens for "call.ended" and creates the
		// conversation row + post-call linkage, so this is the trigger
		// that bridges telephony into the inbox.
		b.publish(refreshed, "call.ended")
	}
}

func (b *EventBridge) onRegister(evt PBXEvent) {
	if evt.Extension == "" {
		return
	}
	// Org lookup is a join; for now we rely on extension uniqueness within
	// the SIP realm. The registration timestamp + status flip is the value.
	if err := b.store.db.Model(&SipExtension{}).
		Where("extension = ?", evt.Extension).
		Updates(map[string]interface{}{
			"status":            "registered",
			"last_registration": evt.OccurredAt,
			"updated_at":        time.Now().UTC(),
		}).Error; err != nil {
		log.Printf("[telephony bridge] mark registered %s: %v", evt.Extension, err)
	}
}

func (b *EventBridge) onRecordingComplete(evt PBXEvent) {
	if evt.CallUUID == "" {
		return
	}
	url := mapString(evt.Data, "Record-File-Path")
	if url == "" {
		return
	}
	if err := b.store.SetRecordingURL(evt.CallUUID, url); err != nil {
		log.Printf("[telephony bridge] set recording url %s: %v", evt.CallUUID, err)
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func (b *EventBridge) appendEvent(callID, eventType string, evt PBXEvent) {
	detail, _ := json.Marshal(evt.Data)
	c, err := b.store.GetCallByIDOnly(callID)
	if err != nil {
		return
	}
	_, _ = b.store.AppendEvent(&CallEvent{
		CallID:    callID,
		OrgID:     c.OrgID,
		EventType: eventType,
		Detail:    detail,
		At:        evt.OccurredAt,
	})
}

func (b *EventBridge) publish(c *Call, subject string) {
	if b.publisher == nil || c == nil {
		return
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"call_id":         c.ID,
		"direction":       c.Direction,
		"status":          c.Status,
		"from_number":     c.FromNumber,
		"to_number":       c.ToNumber,
		"agent_id":        c.AgentID,
		"contact_id":      c.ContactID,
		"hangup_cause":    c.HangupCause,
		"disposition":     c.Disposition,
		"ring_seconds":    c.RingSeconds,
		"talk_seconds":    c.TalkSeconds,
	})
	action := subject
	if idx := lastDot(subject); idx >= 0 {
		action = subject[idx+1:]
	}
	evt, err := events.NewEvent(c.OrgID, c.WorkspaceID, "call", action, c.ID, c.AgentID, payload)
	if err != nil {
		return
	}
	evt.Subject = subject
	_ = b.publisher.Publish(evt)
}

func mapString(m map[string]interface{}, k string) string {
	if v, ok := m[k]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func mapInt(m map[string]interface{}, k string) int {
	s := mapString(m, k)
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func lastDot(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return i
		}
	}
	return -1
}

// dispositionFromCause maps a FreeSWITCH hangup cause to our terminal
// disposition. The full Q.850 cause list is huge; we collapse it down to the
// dispositions the UI displays.
func dispositionFromCause(cause string) string {
	switch cause {
	case "NORMAL_CLEARING":
		return string(DispositionCompleted)
	case "NO_ANSWER", "NO_USER_RESPONSE":
		return string(DispositionNoAnswer)
	case "USER_BUSY":
		return string(DispositionBusy)
	case "", "ORIGINATOR_CANCEL":
		return string(DispositionFailed)
	default:
		return string(DispositionFailed)
	}
}
