-- Phase 6 — Campaigns
-- Channel-agnostic campaigns + per-recipient state + local WhatsApp template mirror.

-- ── campaigns ────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS campaigns (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id        UUID NOT NULL,
  workspace_id  UUID NOT NULL,
  name          TEXT NOT NULL,
  description   TEXT NOT NULL DEFAULT '',
  channel       TEXT NOT NULL,                          -- "whatsapp" | "sms" | "email"
  status        TEXT NOT NULL DEFAULT 'draft',          -- draft|scheduled|running|paused|completed|cancelled
  audience      JSONB NOT NULL DEFAULT '{}'::jsonb,
  schedule      JSONB NOT NULL DEFAULT '{}'::jsonb,
  payload       JSONB NOT NULL DEFAULT '{}'::jsonb,     -- channel-specific (template ref, body, etc.)

  -- Aggregate counters (denormalised for fast list/dashboard reads).
  audience_size INTEGER NOT NULL DEFAULT 0,
  queued_count  INTEGER NOT NULL DEFAULT 0,
  sent_count    INTEGER NOT NULL DEFAULT 0,
  delivered_count INTEGER NOT NULL DEFAULT 0,
  read_count    INTEGER NOT NULL DEFAULT 0,
  failed_count  INTEGER NOT NULL DEFAULT 0,

  -- Temporal coupling. workflow_id is deterministic per launch so re-launch
  -- doesn't duplicate runs. schedule_id is set only when schedule.mode='recurring'.
  temporal_workflow_id TEXT,
  temporal_schedule_id TEXT,

  created_by    UUID,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_campaigns_org_workspace ON campaigns (org_id, workspace_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_status        ON campaigns (status);
CREATE INDEX IF NOT EXISTS idx_campaigns_channel       ON campaigns (channel);

-- ── campaign_recipients ─────────────────────────────────────────────────────
-- One row per (campaign, audience member) post-resolution. Tracks per-row
-- delivery state so analytics aggregates are reconstructible and so retries
-- target only failed rows.
CREATE TABLE IF NOT EXISTS campaign_recipients (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  campaign_id  UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
  org_id       UUID NOT NULL,
  object_id    UUID NOT NULL,                          -- Object Core record (e.g. contact)
  contact_ref  TEXT NOT NULL,                          -- snapshot of phone/email at resolution time
  status       TEXT NOT NULL DEFAULT 'queued',         -- queued|sent|delivered|read|failed
  external_id  TEXT,                                   -- provider message ID (Meta/Twilio)
  error        TEXT,
  sent_at      TIMESTAMPTZ,
  delivered_at TIMESTAMPTZ,
  read_at      TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

  -- A given object should appear once per campaign; redeliveries update the row.
  UNIQUE (campaign_id, object_id)
);

CREATE INDEX IF NOT EXISTS idx_campaign_recipients_campaign_status
  ON campaign_recipients (campaign_id, status);

-- ── whatsapp_templates (local mirror) ───────────────────────────────────────
CREATE TABLE IF NOT EXISTS whatsapp_templates (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id           UUID NOT NULL,
  name             TEXT NOT NULL,                       -- Meta template name
  language         TEXT NOT NULL,                       -- e.g. "en_US"
  category         TEXT NOT NULL,                       -- MARKETING|UTILITY|AUTHENTICATION
  status           TEXT NOT NULL,                       -- APPROVED|PENDING|REJECTED
  header           TEXT NOT NULL DEFAULT '',
  body             TEXT NOT NULL DEFAULT '',
  footer           TEXT NOT NULL DEFAULT '',
  phone_number_id  TEXT NOT NULL DEFAULT '',
  waba_id          TEXT NOT NULL DEFAULT '',
  meta_template_id TEXT NOT NULL DEFAULT '',
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

  -- Per-org uniqueness on (name, language) — Meta uses this composite key.
  UNIQUE (org_id, name, language)
);

CREATE INDEX IF NOT EXISTS idx_whatsapp_templates_status ON whatsapp_templates (status);
