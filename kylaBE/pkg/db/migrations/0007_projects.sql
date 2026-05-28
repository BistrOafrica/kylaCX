-- ============================================================================
-- Migration: Phase 4 completion — Projects
-- Description: Adds projects table backing project.proto service
-- ============================================================================

CREATE TABLE IF NOT EXISTS projects (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID NOT NULL,
    title        TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'active',
    description  TEXT NOT NULL DEFAULT '',
    visibility   TEXT NOT NULL DEFAULT 'private',
    archived_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_projects_org ON projects(org_id);
CREATE INDEX IF NOT EXISTS idx_projects_org_status ON projects(org_id, status);

COMMENT ON TABLE projects IS 'Workspace-agnostic project entities scoped by organisation';
COMMENT ON COLUMN projects.status IS 'active | archived | custom values';

-- ============================================================================
-- End Migration
-- ============================================================================
