-- Phase 5d — Call queues
-- Queue config + agent membership + in-flight queue state for callers.
-- The runtime that picks the next agent lives in internal/telephony/queues/.

-- ── call_queues ──────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS call_queues (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id          UUID NOT NULL,
  workspace_id    UUID NOT NULL,
  name            TEXT NOT NULL,
  description     TEXT NOT NULL DEFAULT '',

  -- Routing strategy. v1 supports two:
  --   "round_robin" — rotate through eligible agents.
  --   "longest_idle" — agent who's been idle the longest answers.
  strategy        TEXT NOT NULL DEFAULT 'longest_idle',

  -- Optional MoH (music on hold) audio file path (FreeSWITCH local_stream path).
  moh_path        TEXT NOT NULL DEFAULT 'local_stream://moh',

  -- Wait limits.
  max_wait_seconds INTEGER NOT NULL DEFAULT 600,
  overflow_action  TEXT NOT NULL DEFAULT 'voicemail',  -- voicemail|hangup|transfer
  overflow_target  TEXT NOT NULL DEFAULT '',

  is_active       BOOLEAN NOT NULL DEFAULT true,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (org_id, name)
);

CREATE INDEX IF NOT EXISTS idx_call_queues_workspace ON call_queues (workspace_id, is_active);

-- ── call_queue_members ──────────────────────────────────────────────────────
-- Maps agents to the queues they can answer for. The agent's SIP extension
-- is resolved at routing time via sip_extensions.user_id.
CREATE TABLE IF NOT EXISTS call_queue_members (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  queue_id    UUID NOT NULL REFERENCES call_queues(id) ON DELETE CASCADE,
  org_id      UUID NOT NULL,
  user_id     UUID NOT NULL,

  -- Skill priority (higher = more preferred). Used by both routing strategies
  -- to weight tie-breaks.
  priority    INTEGER NOT NULL DEFAULT 0,

  -- agentops.AgentStatus is the source of truth for availability; this
  -- column is a cached "is this agent currently active in the queue"
  -- toggle the agent themselves manage (pause/resume).
  is_active   BOOLEAN NOT NULL DEFAULT true,

  -- Tracks longest-idle strategy. Updated when the agent ends a call.
  last_call_ended_at TIMESTAMPTZ,

  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (queue_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_call_queue_members_active
  ON call_queue_members (queue_id, is_active, last_call_ended_at);

-- ── call_queue_entries ──────────────────────────────────────────────────────
-- One row per caller currently waiting in a queue. Created when a call is
-- pushed to a queue and deleted when the call is answered, abandoned, or
-- timed out. Doubles as the wallboard data source.
CREATE TABLE IF NOT EXISTS call_queue_entries (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  queue_id        UUID NOT NULL REFERENCES call_queues(id) ON DELETE CASCADE,
  call_id         UUID NOT NULL,                       -- references calls.id; no FK because call_events fire on rolling time
  org_id          UUID NOT NULL,
  workspace_id    UUID NOT NULL,

  -- Position priorities (higher = served first). VIP callers get a boost.
  priority        INTEGER NOT NULL DEFAULT 0,

  status          TEXT NOT NULL DEFAULT 'waiting',     -- waiting|ringing|connected|abandoned|overflow|timeout
  assigned_agent_id UUID,                              -- set when ringing an agent
  assigned_at     TIMESTAMPTZ,

  entered_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  ended_at        TIMESTAMPTZ,
  ended_reason    TEXT,

  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_call_queue_entries_queue_status
  ON call_queue_entries (queue_id, status, priority DESC, entered_at);
CREATE INDEX IF NOT EXISTS idx_call_queue_entries_call ON call_queue_entries (call_id);
