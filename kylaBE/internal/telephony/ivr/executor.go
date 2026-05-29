package ivr

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"kyla-be/internal/telephony"
)

// Executor drives node execution for active IVR runs.
//
// Lifecycle per call:
//   1. EventBridge sees an inbound CHANNEL_CREATE, looks up the DID, and
//      calls StartForCall(callUUID, flowID, orgID, workspaceID).
//   2. Executor creates an ivr_runs row, loads the flow definition, and
//      runs the start node — issuing the appropriate PBX command (PlayAudio,
//      PlayAndGetDigits, etc.).
//   3. PBX events (PLAYBACK_STOP, DTMF_CAPTURED, CHANNEL_HANGUP) flow back
//      via Advance(...) which the EventBridge calls. Executor transitions
//      to the next node based on the event and the active node's branches.
//   4. Terminal nodes (hangup, transfer) end the run.
//
// Concurrency: one active run per call_uuid. The executor stores the run-in-
// memory cache keyed by call_uuid so per-call state lookups are O(1).
type Executor struct {
	store *Store
	pbx   telephony.PBXController

	mu     sync.RWMutex
	active map[string]*activeRun // call_uuid -> in-memory run state
}

// activeRun is the in-memory mirror of an ivr_runs row. Holds the parsed
// definition so the executor doesn't re-parse JSONB on every event.
type activeRun struct {
	runID      string
	flow       *Flow
	definition Definition
	callUUID   string
	orgID      string
}

func NewExecutor(store *Store, pbx telephony.PBXController) *Executor {
	if pbx == nil {
		pbx = telephony.NoopPBX{}
	}
	return &Executor{
		store:  store,
		pbx:    pbx,
		active: map[string]*activeRun{},
	}
}

// StartForCall kicks off an IVR run for the supplied call. Returns the run ID
// on success. Failure to load the flow or run the first node is logged but
// does not propagate — the call still rings; the caller (EventBridge) treats
// IVR as best-effort augmentation, not a hard precondition.
func (e *Executor) StartForCall(ctx context.Context, callUUID, flowID, orgID, workspaceID string) (string, error) {
	if callUUID == "" || flowID == "" {
		return "", errors.New("ivr: callUUID and flowID required")
	}
	flow, err := e.store.GetFlowByIDOnly(flowID)
	if err != nil {
		return "", fmt.Errorf("ivr: load flow %s: %w", flowID, err)
	}
	if !flow.IsActive {
		return "", fmt.Errorf("ivr: flow %s is not active", flowID)
	}
	def, err := DecodeDefinition(flow.Definition)
	if err != nil {
		return "", fmt.Errorf("ivr: decode definition: %w", err)
	}
	startNode, ok := def.FindNode(def.StartNodeID)
	if !ok {
		return "", fmt.Errorf("ivr: start_node_id %q not present in nodes", def.StartNodeID)
	}

	run := &Run{
		FlowID:        flow.ID,
		CallID:        callUUID,
		OrgID:         orgID,
		WorkspaceID:   workspaceID,
		Status:        string(RunStatusRunning),
		CurrentNodeID: startNode.ID,
		StartedAt:     time.Now().UTC(),
	}
	created, err := e.store.CreateRun(run)
	if err != nil {
		return "", fmt.Errorf("ivr: persist run: %w", err)
	}

	a := &activeRun{
		runID:      created.ID,
		flow:       flow,
		definition: def,
		callUUID:   callUUID,
		orgID:      orgID,
	}
	e.mu.Lock()
	e.active[callUUID] = a
	e.mu.Unlock()

	if err := e.executeNode(ctx, a, startNode); err != nil {
		log.Printf("[ivr] start_node execution failed (run=%s): %v", created.ID, err)
		e.endRun(a, RunStatusFailed, fmt.Sprintf("start_node_execution: %v", err))
		return created.ID, err
	}
	log.Printf("[ivr] started run=%s flow=%s call=%s", created.ID, flow.ID, callUUID)
	return created.ID, nil
}

// Advance applies a PBX event to an active run, transitioning to the next
// node. EventBridge calls this for events that occur during an IVR run.
//
// input carries node-specific data: for menu nodes it's the captured DTMF
// digit; for play_audio it's empty (the branch is the default next_node_id).
func (e *Executor) Advance(ctx context.Context, callUUID string, eventType telephony.PBXEventType, input string) {
	e.mu.RLock()
	a, ok := e.active[callUUID]
	e.mu.RUnlock()
	if !ok {
		// Not an IVR call — silently drop.
		return
	}
	current, ok := a.definition.FindNode(currentNodeID(e.store, a.runID))
	if !ok {
		log.Printf("[ivr] advance: current node missing (run=%s)", a.runID)
		e.endRun(a, RunStatusFailed, "current_node_missing")
		return
	}

	// Hangup ends the run regardless of where we are.
	if eventType == telephony.EventChannelHangup {
		e.endRun(a, RunStatusAbandoned, "caller_hangup")
		return
	}

	// Determine which event the active node was waiting for.
	waitsFor := waitedEventFor(current.Type)
	if waitsFor != "" && eventType != waitsFor {
		// Event isn't what this node expected — drop. (E.g. DTMF arriving for
		// a play_audio node that never asked for input.)
		return
	}

	// Compute the next node ID based on branches + default.
	nextID := chooseNextNode(current, input)
	if nextID == "" {
		// No next node — flow ends successfully.
		e.endRun(a, RunStatusCompleted, "completed")
		return
	}

	nextNode, ok := a.definition.FindNode(nextID)
	if !ok {
		log.Printf("[ivr] advance: branch points to unknown node %q (run=%s)", nextID, a.runID)
		e.endRun(a, RunStatusFailed, "unknown_branch_target")
		return
	}

	if err := e.store.AdvanceRun(a.runID, nextID, input); err != nil {
		log.Printf("[ivr] persist advance failed (run=%s): %v", a.runID, err)
	}
	if err := e.executeNode(ctx, a, nextNode); err != nil {
		log.Printf("[ivr] execute_node failed (run=%s node=%s): %v", a.runID, nextNode.ID, err)
		e.endRun(a, RunStatusFailed, fmt.Sprintf("node_execution: %v", err))
	}
}

// executeNode issues the PBX command corresponding to the node's type.
// Terminal node types (hangup) end the run inline.
func (e *Executor) executeNode(ctx context.Context, a *activeRun, node *Node) error {
	switch node.Type {
	case NodePlayAudio:
		path, _ := node.Config["audio_path"].(string)
		if path == "" {
			return errors.New("play_audio: audio_path required")
		}
		return e.pbx.PlayAudio(ctx, a.callUUID, path)

	case NodeSay:
		text, _ := node.Config["text"].(string)
		if text == "" {
			return errors.New("say: text required")
		}
		voice, _ := node.Config["voice"].(string)
		return e.pbx.SayText(ctx, a.callUUID, voice, text)

	case NodeMenu:
		opts := telephony.PlayAndGetDigitsOpts{
			PromptFile:    stringOf(node.Config, "prompt_file"),
			MinDigits:     intOf(node.Config, "min_digits", 1),
			MaxDigits:     intOf(node.Config, "max_digits", 1),
			Tries:         intOf(node.Config, "tries", 1),
			Timeout:       time.Duration(intOf(node.Config, "timeout_ms", 5000)) * time.Millisecond,
			TerminatorKey: stringOf(node.Config, "terminator"),
			InvalidFile:   stringOf(node.Config, "invalid_file"),
			Regex:         stringOf(node.Config, "regex"),
		}
		return e.pbx.PlayAndGetDigits(ctx, a.callUUID, opts)

	case NodeTransfer:
		target, _ := node.Config["target"].(string)
		if target == "" {
			return errors.New("transfer: target required")
		}
		blind, _ := node.Config["blind"].(bool)
		if !blind {
			blind = true // attended transfer isn't wired yet; force blind
		}
		// Transfer ends the IVR — the call continues but the executor's
		// involvement stops. IVR flows only support blind transfer (a menu
		// can't drive a human consultation step).
		if _, err := e.pbx.Transfer(ctx, a.callUUID, target, true); err != nil {
			return err
		}
		_ = blind // reserved for future "ask before transfer" pattern
		e.endRun(a, RunStatusCompleted, "transferred")
		return nil

	case NodeRecord:
		path, _ := node.Config["recording_path"].(string)
		if path == "" {
			return errors.New("record: recording_path required")
		}
		maxSecs := intOf(node.Config, "max_seconds", 0)
		return e.pbx.StartRecording(ctx, a.callUUID, path, maxSecs)

	case NodeHangup:
		_ = e.pbx.Hangup(ctx, a.callUUID, "NORMAL_CLEARING")
		e.endRun(a, RunStatusCompleted, "hangup_node")
		return nil

	case NodeGoto:
		// Pure routing primitive — immediately advance.
		nextID, _ := node.Config["target_node_id"].(string)
		if nextID == "" {
			return errors.New("goto: target_node_id required")
		}
		nextNode, ok := a.definition.FindNode(nextID)
		if !ok {
			return fmt.Errorf("goto: target node %q missing", nextID)
		}
		_ = e.store.AdvanceRun(a.runID, nextID, "")
		return e.executeNode(ctx, a, nextNode)

	default:
		return fmt.Errorf("unknown node type %q", node.Type)
	}
}

// endRun finalises the in-memory state and persists the terminal status.
func (e *Executor) endRun(a *activeRun, status RunStatus, reason string) {
	e.mu.Lock()
	delete(e.active, a.callUUID)
	e.mu.Unlock()
	if err := e.store.EndRun(a.runID, status, reason); err != nil {
		log.Printf("[ivr] end_run persist failed (run=%s): %v", a.runID, err)
	}
}

// chooseNextNode resolves the next node ID for a transition. Order:
//   1. branches[input] (when input present, e.g. menu DTMF "1" -> nodeA)
//   2. node.next_node_id (default flow)
func chooseNextNode(node *Node, input string) string {
	if input != "" {
		if next, ok := node.Branches[input]; ok {
			return next
		}
		// Allow a "default" branch as fallback.
		if next, ok := node.Branches["default"]; ok {
			return next
		}
	}
	return node.NextNodeID
}

// waitedEventFor returns the PBXEventType the executor expects after issuing
// a command for the given node type. Empty means "doesn't wait — advance on
// the next event of any kind" (used for synchronous-looking nodes like Goto
// that already advanced inline before this gets consulted).
func waitedEventFor(t NodeType) telephony.PBXEventType {
	switch t {
	case NodePlayAudio, NodeSay:
		return telephony.EventPlaybackStop
	case NodeMenu:
		return telephony.EventDTMFCaptured
	case NodeRecord:
		// Record is fire-and-forget from the IVR's perspective; the flow
		// proceeds while recording continues in the background.
		return ""
	}
	return ""
}

// currentNodeID is a helper that reads the most up-to-date current_node_id
// from the DB. Used by Advance to avoid stale in-memory state when the
// previous Advance call hasn't returned yet.
func currentNodeID(store *Store, runID string) string {
	var r Run
	if err := store.db.Select("current_node_id").Where("id = ?", runID).First(&r).Error; err != nil {
		return ""
	}
	return r.CurrentNodeID
}

// ── tiny config helpers ─────────────────────────────────────────────────────

func stringOf(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func intOf(m map[string]interface{}, key string, def int) int {
	if v, ok := m[key]; ok {
		switch x := v.(type) {
		case float64:
			return int(x)
		case int:
			return x
		case int64:
			return int(x)
		}
	}
	return def
}
