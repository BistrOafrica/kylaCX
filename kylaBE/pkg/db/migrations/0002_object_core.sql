-- Migration: 0002_object_core.sql
-- Phase 2 — Object Core Engine (dynamic entity model)

CREATE TABLE IF NOT EXISTS object_types (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    workspace_id  UUID REFERENCES workspaces(id) ON DELETE SET NULL,
    slug          TEXT NOT NULL,
    name          TEXT NOT NULL,
    plural_name   TEXT,
    icon          TEXT,
    color         TEXT,
    is_system     BOOLEAN NOT NULL DEFAULT false,
    -- schema: { "fields": [ { "key", "label", "type", ... } ] }
    schema        JSONB NOT NULL DEFAULT '{"fields":[]}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (org_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_object_types_org    ON object_types (org_id);
CREATE INDEX IF NOT EXISTS idx_object_types_ws     ON object_types (workspace_id);

CREATE TABLE IF NOT EXISTS objects (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL,
    workspace_id  UUID REFERENCES workspaces(id) ON DELETE SET NULL,
    type_slug     TEXT NOT NULL,
    -- data: { "name": "Acme Corp", "priority": "high", ... }
    data          JSONB NOT NULL DEFAULT '{}',
    created_by    UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_objects_org       ON objects (org_id);
CREATE INDEX IF NOT EXISTS idx_objects_ws        ON objects (workspace_id);
CREATE INDEX IF NOT EXISTS idx_objects_type      ON objects (org_id, type_slug);
CREATE INDEX IF NOT EXISTS idx_objects_data_gin  ON objects USING GIN (data);

CREATE TABLE IF NOT EXISTS object_relations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID NOT NULL,
    from_id    UUID NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    to_id      UUID NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    -- relation: "contact_of", "related_to", "child_of", "participates_in"
    relation   TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_obj_rel_from  ON object_relations (from_id);
CREATE INDEX IF NOT EXISTS idx_obj_rel_to    ON object_relations (to_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_obj_rel_unique ON object_relations (from_id, to_id, relation);

CREATE TABLE IF NOT EXISTS object_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL,
    object_id   UUID NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    actor_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_type  TEXT NOT NULL DEFAULT 'user',
    event_type  TEXT NOT NULL,
    payload     JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_obj_events_object ON object_events (object_id);
CREATE INDEX IF NOT EXISTS idx_obj_events_org    ON object_events (org_id);
CREATE INDEX IF NOT EXISTS idx_obj_events_ts     ON object_events (created_at DESC);
