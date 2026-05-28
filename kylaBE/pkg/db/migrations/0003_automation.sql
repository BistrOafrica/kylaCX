-- Migration: 0003_automation.sql
-- Phase 2/6 — Workflow / Automation Engine tables

CREATE TABLE IF NOT EXISTS workflows (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    workspace_id  UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    description   TEXT,
    status        TEXT NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('active','inactive','draft')),
    -- trigger: { "type": "event", "event_subject_pattern": "kyla.*.ticket.created" }
    trigger       JSONB NOT NULL DEFAULT '{}',
    -- conditions: [ { "operator": "AND", "conditions": [...] } ]
    conditions    JSONB NOT NULL DEFAULT '[]',
    -- actions: [ { "id": "...", "type": "update_object", "config": {...} } ]
    actions       JSONB NOT NULL DEFAULT '[]',
    run_count     BIGINT NOT NULL DEFAULT 0,
    created_by    UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_workflows_org    ON workflows (org_id);
CREATE INDEX IF NOT EXISTS idx_workflows_ws     ON workflows (workspace_id);
CREATE INDEX IF NOT EXISTS idx_workflows_status ON workflows (status) WHERE status = 'active';

CREATE TABLE IF NOT EXISTS workflow_runs (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id       UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    org_id            UUID NOT NULL,
    trigger_event_id  TEXT,
    status            TEXT NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending','running','success','failed','skipped')),
    context           JSONB,
    error             TEXT,
    started_at        TIMESTAMPTZ,
    finished_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_workflow_runs_wf     ON workflow_runs (workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_status ON workflow_runs (status);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_ts     ON workflow_runs (created_at DESC);
