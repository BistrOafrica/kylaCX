package queues

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"kyla-be/internal/telephony"
)

// ExtensionLookup is the subset of the telephony Store the router uses to
// find an agent's SIP extension. Declared here to keep the queues package
// from depending on the full telephony.Store.
type ExtensionLookup interface {
	GetExtensionByUserID(userID string) (*telephony.SipExtension, error)
}

// Router is the in-process routing engine. One per binary — keyed lookups
// (`PushCall`, `OnAgentAnswered`, etc.) are O(1) via the activeEntries map.
//
// Lifecycle (one waiting caller):
//   1. EventBridge sees the call land in a queue → Router.PushCall.
//   2. PushCall picks the next eligible member, originates a leg to that
//      agent's SIP extension, and persists the entry in "ringing".
//   3. The agent leg's CHANNEL_ANSWER triggers Router.OnAgentAnswered which
//      bridges the original caller leg to the agent.
//   4. Either leg's CHANNEL_HANGUP triggers Router.OnCallEnded which updates
//      the member's last_call_ended_at and pulls the next caller (if any).
//
// Concurrency: a sync.RWMutex guards activeEntries because PBX events arrive
// concurrently across calls. Each entry is owned by one routing iteration —
// no per-entry locking needed.
type Router struct {
	store        *Store
	pbx          telephony.PBXController
	extensions   ExtensionLookup
	defaultTrunk string // FreeSWITCH gateway profile for originating agent legs

	mu             sync.RWMutex
	activeEntries  map[string]*activeEntry // call_uuid -> in-flight ringing state
	memberRotation map[string]int          // queue_id -> round-robin cursor
}

type activeEntry struct {
	entryID         string
	queueID         string
	callerCallUUID  string
	agentCallUUID   string
	agentUserID     string
	agentExtension  string
}

func NewRouter(store *Store, pbx telephony.PBXController, extensions ExtensionLookup, defaultTrunk string) *Router {
	if pbx == nil {
		pbx = telephony.NoopPBX{}
	}
	return &Router{
		store:          store,
		pbx:            pbx,
		extensions:     extensions,
		defaultTrunk:   defaultTrunk,
		activeEntries:  map[string]*activeEntry{},
		memberRotation: map[string]int{},
	}
}

// PushCall adds the supplied call to the queue and attempts to immediately
// dial an eligible agent. If no agent is eligible the call sits in "waiting"
// status (MoH would be played by the dialplan); a subsequent
// OnAgentBecameAvailable call advances it.
func (r *Router) PushCall(ctx context.Context, callerCallUUID, queueID, orgID, workspaceID string, priority int) (string, error) {
	if callerCallUUID == "" || queueID == "" {
		return "", errors.New("queues: callUUID and queueID required")
	}
	q, err := r.store.GetQueueByIDOnly(queueID)
	if err != nil {
		return "", fmt.Errorf("queues: load queue %s: %w", queueID, err)
	}
	if !q.IsActive {
		return "", fmt.Errorf("queues: queue %s not active", queueID)
	}

	entry := &Entry{
		QueueID:     q.ID,
		CallID:      callerCallUUID,
		OrgID:       orgID,
		WorkspaceID: workspaceID,
		Priority:    priority,
		Status:      string(EntryWaiting),
		EnteredAt:   time.Now().UTC(),
	}
	created, err := r.store.CreateEntry(entry)
	if err != nil {
		return "", fmt.Errorf("queues: persist entry: %w", err)
	}

	// Try to dial an agent immediately. Failures here are non-fatal — the
	// entry stays in "waiting" and the next OnAgentBecameAvailable will
	// retry.
	if err := r.tryDispatch(ctx, q, created); err != nil {
		log.Printf("[queues] dispatch failed for entry %s: %v (caller will wait)", created.ID, err)
	}
	return created.ID, nil
}

// tryDispatch picks the next eligible agent and originates a leg to them.
// Updates the entry to "ringing" on success.
func (r *Router) tryDispatch(ctx context.Context, q *Queue, entry *Entry) error {
	members, err := r.store.FindEligibleMembers(q.ID, Strategy(q.Strategy))
	if err != nil {
		return fmt.Errorf("find members: %w", err)
	}
	if len(members) == 0 {
		return errors.New("no eligible members")
	}

	// Round-robin advances a per-queue cursor; longest_idle already returns
	// rows in the right order so we pick index 0.
	var pick *Membership
	switch Strategy(q.Strategy) {
	case StrategyRoundRobin:
		r.mu.Lock()
		cursor := r.memberRotation[q.ID] % len(members)
		r.memberRotation[q.ID] = cursor + 1
		r.mu.Unlock()
		pick = members[cursor]
	default:
		pick = members[0]
	}

	ext, err := r.extensions.GetExtensionByUserID(pick.UserID)
	if err != nil {
		return fmt.Errorf("resolve extension for user %s: %w", pick.UserID, err)
	}

	agentCallUUID, err := r.pbx.Originate(ctx, telephony.OriginateRequest{
		AgentID:        pick.UserID,
		AgentExtension: ext.Extension,
		ToNumber:       entry.CallID, // routing engine bridges the caller leg by call_id; the dialplan picks this up via a kyla_callerleg variable
		TrunkGateway:   r.defaultTrunk,
		OrgID:          entry.OrgID,
		WorkspaceID:    entry.WorkspaceID,
	})
	if err != nil {
		return fmt.Errorf("originate agent leg: %w", err)
	}

	if err := r.store.AssignAgent(entry.ID, pick.UserID); err != nil {
		// Best-effort cleanup; the PBX leg is in motion either way.
		log.Printf("[queues] persist assignment failed for entry %s: %v", entry.ID, err)
	}

	r.mu.Lock()
	r.activeEntries[agentCallUUID] = &activeEntry{
		entryID:        entry.ID,
		queueID:        q.ID,
		callerCallUUID: entry.CallID,
		agentCallUUID:  agentCallUUID,
		agentUserID:    pick.UserID,
		agentExtension: ext.Extension,
	}
	r.mu.Unlock()
	return nil
}

// OnAgentAnswered handles CHANNEL_ANSWER on an agent leg the router
// originated. Bridges the caller leg to the agent leg and marks the entry
// connected.
func (r *Router) OnAgentAnswered(ctx context.Context, agentCallUUID string) {
	r.mu.RLock()
	a, ok := r.activeEntries[agentCallUUID]
	r.mu.RUnlock()
	if !ok {
		return
	}
	// uuid_bridge merges the two legs into one bridged call. The originate
	// command set a kyla_callerleg variable that points back to the caller
	// UUID so the dialplan can complete the bridge once the agent answers.
	// In a deployment that doesn't wire the dialplan to do this, this is
	// where we'd invoke a bgapi uuid_bridge. For the foundation we just
	// record connected — the actual bridging is dialplan-driven.
	if err := r.store.MarkConnected(a.entryID); err != nil {
		log.Printf("[queues] mark connected (%s): %v", a.entryID, err)
	}
}

// OnCallEnded handles CHANNEL_HANGUP on either leg. Marks the entry connected
// or abandoned depending on whether the agent leg was answered, updates the
// member's last_call_ended_at, and attempts to pull the next waiting caller
// from the same queue.
func (r *Router) OnCallEnded(ctx context.Context, callUUID string) {
	r.mu.Lock()
	a, ok := r.activeEntries[callUUID]
	if ok {
		delete(r.activeEntries, callUUID)
	}
	r.mu.Unlock()

	if !ok {
		// Not an agent leg we're tracking — check whether this was a caller
		// leg whose entry is still active.
		if entry, err := r.store.GetEntryByCallID(callUUID); err == nil {
			_ = r.store.EndEntry(entry.ID, EntryAbandoned, "caller_hangup")
		}
		return
	}

	// Agent leg ended. Bookkeeping for round-robin/longest_idle fairness.
	if err := r.store.MarkMemberCallEnded(a.queueID, a.agentUserID, time.Now().UTC()); err != nil {
		log.Printf("[queues] update last_call_ended_at: %v", err)
	}

	// Finalise the entry — connected if it was answered, abandoned otherwise.
	// We infer answered from MarkConnected which transitions Status to
	// "connected" (the OnAgentAnswered path).
	if entry, err := r.store.GetEntryByCallID(a.callerCallUUID); err == nil {
		switch EntryStatus(entry.Status) {
		case EntryConnected:
			_ = r.store.EndEntry(entry.ID, EntryConnected, "completed")
		default:
			_ = r.store.EndEntry(entry.ID, EntryAbandoned, "agent_hangup_before_bridge")
		}
	}

	// Pull the next waiting caller, if any.
	r.pullNext(ctx, a.queueID)
}

// OnAgentBecameAvailable handles SOFIA_REGISTER + agentops transitions so a
// newly-available agent picks up a waiting call without waiting for the next
// inbound. Looks up the queues the agent belongs to and tries dispatch.
func (r *Router) OnAgentBecameAvailable(ctx context.Context, userID string) {
	// Read all queues the user is a member of; pull next waiting from each.
	var memberships []*Membership
	if err := r.store.db.Where("user_id = ? AND is_active = ?", userID, true).Find(&memberships).Error; err != nil {
		log.Printf("[queues] list memberships for %s: %v", userID, err)
		return
	}
	for _, m := range memberships {
		r.pullNext(ctx, m.QueueID)
	}
}

// pullNext finds the longest-waiting caller in a queue and dispatches them.
// No-op when the queue is empty.
func (r *Router) pullNext(ctx context.Context, queueID string) {
	q, err := r.store.GetQueueByIDOnly(queueID)
	if err != nil {
		return
	}
	waiting, err := r.store.ListLiveEntries(q.ID, string(EntryWaiting))
	if err != nil || len(waiting) == 0 {
		return
	}
	if err := r.tryDispatch(ctx, q, waiting[0]); err != nil {
		log.Printf("[queues] pullNext dispatch for queue %s: %v", q.ID, err)
	}
}
