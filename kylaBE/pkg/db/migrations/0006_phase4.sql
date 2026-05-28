-- ============================================================================
-- Migration: Phase 4 — CRM, Ticketing, Knowledge Base, Forms
-- Description: Adds pipeline, ticket room, knowledge base, and form tables
-- ============================================================================

-- ── CRM Pipelines ────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS crm_pipelines (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID NOT NULL,
    workspace_id UUID,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    type         TEXT NOT NULL DEFAULT 'sales',
    color        TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_crm_pipelines_org ON crm_pipelines(org_id);
CREATE INDEX idx_crm_pipelines_org_ws ON crm_pipelines(org_id, workspace_id);

COMMENT ON TABLE crm_pipelines IS 'CRM deal pipeline definitions';
COMMENT ON COLUMN crm_pipelines.type IS 'sales | support | custom';

-- ── CRM Pipeline Stages ───────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS crm_pipeline_stages (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL,
    org_id      UUID NOT NULL,
    name        TEXT NOT NULL,
    color       TEXT NOT NULL DEFAULT '',
    index       INTEGER NOT NULL DEFAULT 0,
    probability INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_crm_stages_pipeline ON crm_pipeline_stages(pipeline_id);
CREATE INDEX idx_crm_stages_org ON crm_pipeline_stages(org_id);

COMMENT ON TABLE crm_pipeline_stages IS 'Ordered stages within a CRM pipeline';
COMMENT ON COLUMN crm_pipeline_stages.index IS 'Zero-based display order within the pipeline';
COMMENT ON COLUMN crm_pipeline_stages.probability IS 'Win probability percentage for this stage (0-100)';

-- ── Ticket Rooms ──────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS ticket_rooms (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id     UUID NOT NULL,
    org_id        UUID NOT NULL,
    name          TEXT NOT NULL,
    type          TEXT NOT NULL DEFAULT 'internal',
    message_count INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ticket_rooms_ticket ON ticket_rooms(ticket_id);
CREATE INDEX idx_ticket_rooms_org ON ticket_rooms(org_id);

COMMENT ON TABLE ticket_rooms IS 'Threaded discussion rooms attached to tickets';
COMMENT ON COLUMN ticket_rooms.type IS 'internal | customer_reply';

-- ── Ticket Room Messages ─────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS ticket_room_messages (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id    UUID NOT NULL,
    org_id     UUID NOT NULL,
    author_id  UUID NOT NULL,
    content    TEXT NOT NULL,
    is_private BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_trm_room ON ticket_room_messages(room_id);
CREATE INDEX idx_trm_org ON ticket_room_messages(org_id);
CREATE INDEX idx_trm_room_created ON ticket_room_messages(room_id, created_at DESC);

COMMENT ON TABLE ticket_room_messages IS 'Individual messages in a ticket room';

-- ── Ticket Macros ─────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS ticket_macros (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID NOT NULL,
    workspace_id UUID,
    name         TEXT NOT NULL,
    content      TEXT NOT NULL DEFAULT '',
    actions      JSONB NOT NULL DEFAULT '[]',
    visibility   TEXT NOT NULL DEFAULT 'private',
    created_by   UUID,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ticket_macros_org ON ticket_macros(org_id);
CREATE INDEX idx_ticket_macros_org_ws ON ticket_macros(org_id, workspace_id);

COMMENT ON TABLE ticket_macros IS 'Canned responses and field-patch action sets for tickets';
COMMENT ON COLUMN ticket_macros.visibility IS 'private | team | public';
COMMENT ON COLUMN ticket_macros.actions IS 'JSON array of {field, value} patch instructions';

-- ── KB Categories ─────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS kb_categories (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL,
    workspace_id  UUID,
    name          TEXT NOT NULL,
    slug          TEXT NOT NULL,
    icon          TEXT NOT NULL DEFAULT '',
    parent_id     UUID,
    position      INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_kb_categories_org ON kb_categories(org_id);
CREATE INDEX idx_kb_categories_org_ws ON kb_categories(org_id, workspace_id);
CREATE INDEX idx_kb_categories_parent ON kb_categories(parent_id) WHERE parent_id IS NOT NULL;

COMMENT ON TABLE kb_categories IS 'Workspace-scoped knowledge base categories';

-- ── KB Articles ───────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS kb_articles (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID NOT NULL,
    workspace_id UUID,
    category_id  UUID,
    title        TEXT NOT NULL,
    slug         TEXT NOT NULL,
    content      TEXT NOT NULL DEFAULT '',
    excerpt      TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'draft',
    author_id    UUID,
    published_at TIMESTAMPTZ,
    view_count   INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_kb_articles_org ON kb_articles(org_id);
CREATE INDEX idx_kb_articles_org_ws ON kb_articles(org_id, workspace_id);
CREATE INDEX idx_kb_articles_category ON kb_articles(category_id);
CREATE INDEX idx_kb_articles_status ON kb_articles(org_id, status);
CREATE INDEX idx_kb_articles_title ON kb_articles USING gin(to_tsvector('english', title));
CREATE INDEX idx_kb_articles_fts ON kb_articles USING gin(to_tsvector('english', title || ' ' || content));

COMMENT ON TABLE kb_articles IS 'Workspace-scoped knowledge base articles';
COMMENT ON COLUMN kb_articles.status IS 'draft | published | archived';

-- ── Forms ─────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS forms (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id           UUID NOT NULL,
    workspace_id     UUID,
    name             TEXT NOT NULL,
    description      TEXT NOT NULL DEFAULT '',
    fields           JSONB NOT NULL DEFAULT '[]',
    status           TEXT NOT NULL DEFAULT 'draft',
    submit_redirect  TEXT NOT NULL DEFAULT '',
    submission_count INTEGER NOT NULL DEFAULT 0,
    created_by       UUID,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_forms_org ON forms(org_id);
CREATE INDEX idx_forms_org_ws ON forms(org_id, workspace_id);
CREATE INDEX idx_forms_status ON forms(org_id, status);

COMMENT ON TABLE forms IS 'Data-collection form definitions';
COMMENT ON COLUMN forms.status IS 'draft | active | closed';
COMMENT ON COLUMN forms.fields IS 'JSON array of field definition objects';

-- ── Form Submissions ──────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS form_submissions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id    UUID NOT NULL,
    org_id     UUID NOT NULL,
    data       JSONB NOT NULL DEFAULT '{}',
    object_id  UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_form_submissions_form ON form_submissions(form_id);
CREATE INDEX idx_form_submissions_org ON form_submissions(org_id);
CREATE INDEX idx_form_submissions_form_org ON form_submissions(form_id, org_id);

COMMENT ON TABLE form_submissions IS 'Individual form submission responses';
COMMENT ON COLUMN form_submissions.data IS 'JSON object mapping field keys to submitted values';
COMMENT ON COLUMN form_submissions.object_id IS 'Linked Object Core record (nullable, best-effort)';

-- ============================================================================
-- End Migration
-- ============================================================================
