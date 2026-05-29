-- Phase 5 — Telephony (self-hosted SIP via FreeSWITCH)
-- Call sessions, per-call event log, and SIP admin tables.

-- ── calls ────────────────────────────────────────────────────────────────────
-- One row per call leg. The id matches the FreeSWITCH call UUID so we can
-- correlate ESL events back to our row without a separate mapping table.
CREATE TABLE IF NOT EXISTS calls (
  id              UUID PRIMARY KEY,                       -- FreeSWITCH UUID (assigned by PBX)
  org_id          UUID NOT NULL,
  workspace_id    UUID NOT NULL,
  direction       TEXT NOT NULL,                          -- inbound | outbound
  status          TEXT NOT NULL DEFAULT 'ringing',        -- ringing | answered | ended | failed

  from_number     TEXT NOT NULL DEFAULT '',
  to_number       TEXT NOT NULL DEFAULT '',

  agent_id        UUID,
  contact_id      UUID,
  queue_id        UUID,
  trunk_id        UUID,
  ivr_flow_id     UUID,

  -- Cross-domain linkage. Populated by automation or post-call activities.
  conversation_id UUID,
  deal_id         UUID,
  ticket_id       UUID,

  recording_enabled BOOLEAN NOT NULL DEFAULT false,
  recording_url     TEXT,

  started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  answered_at     TIMESTAMPTZ,
  ended_at        TIMESTAMPTZ,
  ring_seconds    INTEGER NOT NULL DEFAULT 0,
  talk_seconds    INTEGER NOT NULL DEFAULT 0,

  hangup_cause    TEXT,
  disposition     TEXT,

  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_calls_workspace_started ON calls (workspace_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_calls_agent             ON calls (agent_id);
CREATE INDEX IF NOT EXISTS idx_calls_contact           ON calls (contact_id);
CREATE INDEX IF NOT EXISTS idx_calls_status            ON calls (status);

-- ── call_events ──────────────────────────────────────────────────────────────
-- Per-call event log: ring, answer, transfer, hold, hangup. Mirrors the
-- FreeSWITCH event stream so we can reconstruct call history without keeping
-- raw ESL traces forever.
CREATE TABLE IF NOT EXISTS call_events (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  call_id     UUID NOT NULL REFERENCES calls(id) ON DELETE CASCADE,
  org_id      UUID NOT NULL,
  event_type  TEXT NOT NULL,                              -- started|answered|transferred|held|resumed|ended|note
  detail      JSONB NOT NULL DEFAULT '{}'::jsonb,
  at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_call_events_call_at ON call_events (call_id, at);

-- ── sip_domains ──────────────────────────────────────────────────────────────
-- A SIP realm (FQDN-like). One default per org; extensions belong to a domain.
CREATE TABLE IF NOT EXISTS sip_domains (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id     UUID NOT NULL,
  domain     TEXT NOT NULL,
  is_default BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (org_id, domain)
);

-- Only one default domain per org.
CREATE UNIQUE INDEX IF NOT EXISTS uq_sip_domains_default
  ON sip_domains (org_id) WHERE is_default;

-- ── sip_extensions ───────────────────────────────────────────────────────────
-- One per agent in v1. Maps a SIP extension number (e.g. "1001") to a kyla
-- user so inbound calls can be routed to the right agent and outbound calls
-- can be originated from the right caller-ID.
CREATE TABLE IF NOT EXISTS sip_extensions (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id            UUID NOT NULL,
  workspace_id      UUID NOT NULL,
  user_id           UUID NOT NULL,
  domain_id         UUID REFERENCES sip_domains(id),
  extension         TEXT NOT NULL,
  display_name      TEXT NOT NULL DEFAULT '',

  -- Stored SIP credentials. Password is stored as a hash so a leaked DB
  -- can't be replayed against the PBX. The PBX is provisioned with the
  -- plain credential separately during extension creation.
  password_hash     TEXT NOT NULL DEFAULT '',

  status            TEXT NOT NULL DEFAULT 'unregistered', -- registered|unregistered
  last_registration TIMESTAMPTZ,

  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (org_id, extension),
  UNIQUE (user_id)  -- v1: one extension per user
);

CREATE INDEX IF NOT EXISTS idx_sip_extensions_workspace ON sip_extensions (workspace_id);

-- ── sip_trunks ───────────────────────────────────────────────────────────────
-- Outbound SIP trunks for PSTN connectivity. The gateway_name maps to a
-- FreeSWITCH gateway profile provisioned out-of-band; this row is the
-- control-plane source of truth for routing decisions.
CREATE TABLE IF NOT EXISTS sip_trunks (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id       UUID NOT NULL,
  name         TEXT NOT NULL,
  gateway_name TEXT NOT NULL,                              -- FS gateway profile name
  provider     TEXT NOT NULL DEFAULT 'custom',
  sip_server   TEXT NOT NULL DEFAULT '',
  username     TEXT NOT NULL DEFAULT '',
  password     TEXT NOT NULL DEFAULT '',                   -- write-only on read path
  from_uri     TEXT NOT NULL DEFAULT '',
  is_active    BOOLEAN NOT NULL DEFAULT true,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (org_id, name)
);

CREATE INDEX IF NOT EXISTS idx_sip_trunks_active ON sip_trunks (org_id, is_active);
