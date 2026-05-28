-- Migration: 0001_workspaces.sql
-- Phase 1 — Workspace primitive

CREATE TABLE IF NOT EXISTS workspaces (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id           UUID NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    name             TEXT NOT NULL,
    slug             TEXT NOT NULL,
    description      TEXT,
    icon             TEXT,
    color            TEXT,
    domain_template  TEXT NOT NULL DEFAULT 'custom'
                        CHECK (domain_template IN ('support','sales','marketing','operations','custom')),
    status           TEXT NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active','archived','suspended')),
    config           JSONB NOT NULL DEFAULT '{}',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (org_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_workspaces_org_id ON workspaces (org_id);

CREATE TABLE IF NOT EXISTS workspace_members (
    workspace_id  UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role          TEXT NOT NULL DEFAULT 'member'
                     CHECK (role IN ('owner', 'admin', 'member', 'guest')),
    joined_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (workspace_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_workspace_members_user ON workspace_members (user_id);
