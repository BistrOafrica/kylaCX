-- ============================================================================
-- Migration: Automation Workflows (Phase 6)
-- Description: Workflow definitions + Temporal run projections
-- ============================================================================

-- ── Workflow Definitions ────────────────────────────────────────────────────
-- A Workflow is a Temporal-backed automation: trigger → [conditions] → actions.
-- The definition lives here; Temporal owns the execution state.

CREATE TABLE IF NOT EXISTS workflows (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID NOT NULL,
    workspace_id UUID NOT NULL,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    trigger      JSONB NOT NULL DEFAULT '{}'::jsonb,
    conditions   JSONB NOT NULL DEFAULT '[]'::jsonb,
    actions      JSONB NOT NULL DEFAULT '[]'::jsonb,
    status       TEXT NOT NULL DEFAULT 'draft',
    run_count    INTEGER NOT NULL DEFAULT 0,
    created_by   UUID,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_workflows_org_workspace ON workflows(org_id, workspace_id);
CREATE INDEX IF NOT EXISTS idx_workflows_status ON workflows(status);
-- Fast trigger-matching lookups: the consumer queries workflows whose
-- trigger.type matches a given DomainEvent subject prefix.
CREATE INDEX IF NOT EXISTS idx_workflows_trigger_type
    ON workflows ((trigger->>'type'))
    WHERE status = 'active';

COMMENT ON TABLE workflows IS 'Automation workflow definitions; executed by Temporal';
COMMENT ON COLUMN workflows.trigger IS '{type, object_type, ...} JSON describing what fires this workflow';
COMMENT ON COLUMN workflows.conditions IS 'Array of {field, op, value} filters applied to the trigger event';
COMMENT ON COLUMN workflows.actions IS 'Ordered array of {type, params} action nodes';
COMMENT ON COLUMN workflows.status IS 'draft | active | inactive';

-- ── Workflow Run Projection ─────────────────────────────────────────────────
-- Temporal owns the canonical execution state, but we project a row here per
-- run so the Kyla UI can list runs without round-tripping to Temporal, and
-- deep-link into Temporal Web via temporal_run_id.

CREATE TABLE IF NOT EXISTS workflow_runs (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id      UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    temporal_run_id  TEXT NOT NULL,
    trigger_event_id TEXT,
    status           TEXT NOT NULL DEFAULT 'pending',
    context          JSONB NOT NULL DEFAULT '{}'::jsonb,
    error            TEXT,
    started_at       TIMESTAMPTZ,
    finished_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_workflow_runs_workflow ON workflow_runs(workflow_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_status ON workflow_runs(status);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_temporal_id ON workflow_runs(temporal_run_id);

COMMENT ON TABLE workflow_runs IS 'Per-execution projection of Temporal workflow runs';
COMMENT ON COLUMN workflow_runs.temporal_run_id IS 'Temporal RunID for deep-link into Temporal Web';
COMMENT ON COLUMN workflow_runs.status IS 'pending | running | success | failed | skipped';
COMMENT ON COLUMN workflow_runs.context IS 'Trigger event payload + execution metadata';

-- ============================================================================
-- End Migration
-- ============================================================================
