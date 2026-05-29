-- Phase 5c — IVR (Interactive Voice Response)
-- Node-based flow engine that executes against in-progress inbound calls.
-- The node model mirrors Phase 6 workflows so the visual builder can reuse
-- the same React Flow canvas.

-- ── ivr_flows ────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS ivr_flows (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id        UUID NOT NULL,
  workspace_id  UUID NOT NULL,
  name          TEXT NOT NULL,
  description   TEXT NOT NULL DEFAULT '',

  -- Node graph: { "start_node_id": "...", "nodes": [{id, type, config, ...}] }
  -- Stored as JSONB so the visual builder can persist arbitrary node shapes
  -- without round-tripping through dedicated columns.
  definition    JSONB NOT NULL DEFAULT '{}'::jsonb,

  is_active     BOOLEAN NOT NULL DEFAULT false,
  version       INTEGER NOT NULL DEFAULT 1,

  created_by    UUID,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (org_id, name)
);

CREATE INDEX IF NOT EXISTS idx_ivr_flows_workspace ON ivr_flows (workspace_id, is_active);

-- ── ivr_did_mappings ────────────────────────────────────────────────────────
-- Routes inbound DIDs (called numbers) to IVR flows. One DID can map to
-- exactly one active flow per org. Unmapped DIDs fall through to the default
-- workspace inbound handling (which is typically "route to queue").
CREATE TABLE IF NOT EXISTS ivr_did_mappings (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id        UUID NOT NULL,
  workspace_id  UUID NOT NULL,
  did           TEXT NOT NULL,                       -- E.164
  flow_id       UUID NOT NULL REFERENCES ivr_flows(id) ON DELETE CASCADE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (org_id, did)
);

CREATE INDEX IF NOT EXISTS idx_ivr_did_mappings_did ON ivr_did_mappings (did);

-- ── ivr_runs ────────────────────────────────────────────────────────────────
-- One row per IVR execution. Tracks the path the caller took through the
-- flow so analytics + debugging can replay the experience. Updated as the
-- executor advances through nodes.
CREATE TABLE IF NOT EXISTS ivr_runs (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  flow_id         UUID NOT NULL REFERENCES ivr_flows(id) ON DELETE CASCADE,
  call_id         UUID NOT NULL,                     -- references calls.id; no FK because call may not exist yet at insert time
  org_id          UUID NOT NULL,
  workspace_id    UUID NOT NULL,

  status          TEXT NOT NULL DEFAULT 'running',   -- running|completed|failed|abandoned
  current_node_id TEXT NOT NULL DEFAULT '',          -- mirrors the position in the flow graph
  visited_nodes   JSONB NOT NULL DEFAULT '[]'::jsonb, -- ordered list of {node_id, entered_at, input?}

  started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  ended_at        TIMESTAMPTZ,
  end_reason      TEXT,                              -- e.g. "completed", "caller_hangup", "transfer", "error"

  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ivr_runs_call    ON ivr_runs (call_id);
CREATE INDEX IF NOT EXISTS idx_ivr_runs_flow    ON ivr_runs (flow_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_ivr_runs_status  ON ivr_runs (status);
