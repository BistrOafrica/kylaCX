-- Migration: 0004_conversations.sql
-- Phase 3 — Unified Inbox / Communication Layer

CREATE TABLE IF NOT EXISTS conversations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    -- channel: "whatsapp", "email", "sms", "voice", "webchat", "instagram", "messenger"
    channel         TEXT NOT NULL,
    -- channel_ref: external thread/thread ID specific to the channel provider
    channel_ref     TEXT,
    -- contact_id links to the Object Core contact record
    contact_id      UUID REFERENCES objects(id) ON DELETE SET NULL,
    assigned_to     UUID REFERENCES users(id) ON DELETE SET NULL,
    team_id         UUID,
    -- status: "open", "pending", "resolved", "snoozed"
    status          TEXT NOT NULL DEFAULT 'open'
                        CHECK (status IN ('open','pending','resolved','snoozed')),
    priority        TEXT NOT NULL DEFAULT 'normal'
                        CHECK (priority IN ('low','normal','high','urgent')),
    subject         TEXT,
    sla_deadline    TIMESTAMPTZ,
    snoozed_until   TIMESTAMPTZ,
    resolved_at     TIMESTAMPTZ,
    -- meta: provider-specific metadata (e.g. WA profile name, email headers)
    meta            JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_conv_org         ON conversations (org_id);
CREATE INDEX IF NOT EXISTS idx_conv_ws          ON conversations (workspace_id);
CREATE INDEX IF NOT EXISTS idx_conv_contact     ON conversations (contact_id);
CREATE INDEX IF NOT EXISTS idx_conv_assigned    ON conversations (assigned_to);
CREATE INDEX IF NOT EXISTS idx_conv_status      ON conversations (org_id, status);
CREATE INDEX IF NOT EXISTS idx_conv_channel     ON conversations (org_id, channel);
CREATE UNIQUE INDEX IF NOT EXISTS idx_conv_channel_ref ON conversations (org_id, channel, channel_ref)
    WHERE channel_ref IS NOT NULL;

CREATE TABLE IF NOT EXISTS messages (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id  UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    -- sender_id is user_id for agents/bots, contact object_id for customers
    sender_id        UUID,
    -- sender_type: "agent", "contact", "bot", "system"
    sender_type      TEXT NOT NULL DEFAULT 'agent'
                        CHECK (sender_type IN ('agent','contact','bot','system')),
    channel          TEXT NOT NULL,
    -- content_type: "text", "image", "audio", "video", "file", "template", "interactive"
    content_type     TEXT NOT NULL DEFAULT 'text',
    -- content: { "text": "Hello!", "url": "...", "template_name": "..." }
    content          JSONB NOT NULL,
    -- status: "pending", "sent", "delivered", "read", "failed"
    status           TEXT NOT NULL DEFAULT 'sent'
                        CHECK (status IN ('pending','sent','delivered','read','failed')),
    -- external_id: provider-assigned message ID for delivery tracking
    external_id      TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_msg_conversation ON messages (conversation_id);
CREATE INDEX IF NOT EXISTS idx_msg_sender       ON messages (sender_id);
CREATE INDEX IF NOT EXISTS idx_msg_ts           ON messages (conversation_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_msg_external_id  ON messages (external_id) WHERE external_id IS NOT NULL;
