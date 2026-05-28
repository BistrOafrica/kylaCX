-- ============================================================================
-- Migration: Routing Rules and SLA Policies
-- Description: Adds tables for routing automation and SLA tracking
-- ============================================================================

-- ── Routing Rules ───────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS routing_rules (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID NOT NULL,
    workspace_id UUID NOT NULL,
    name         TEXT NOT NULL,
    priority     INTEGER NOT NULL DEFAULT 0,
    conditions   JSONB NOT NULL DEFAULT '[]'::jsonb,
    actions      JSONB NOT NULL DEFAULT '[]'::jsonb,
    strategy     TEXT NOT NULL DEFAULT 'round_robin',
    is_active    BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_routing_rules_org_workspace ON routing_rules(org_id, workspace_id);
CREATE INDEX idx_routing_rules_priority ON routing_rules(priority DESC) WHERE is_active = true;
CREATE INDEX idx_routing_rules_active ON routing_rules(is_active);

COMMENT ON TABLE routing_rules IS 'Automatic conversation routing rules';
COMMENT ON COLUMN routing_rules.priority IS 'Higher values are evaluated first';
COMMENT ON COLUMN routing_rules.conditions IS 'Array of {field,op,value} conditions';
COMMENT ON COLUMN routing_rules.actions IS 'Array of {type,target_id} actions';
COMMENT ON COLUMN routing_rules.strategy IS 'round_robin | skill_based | direct';

-- ── SLA Policies ────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS sla_policies (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID NOT NULL,
    workspace_id UUID NOT NULL,
    name         TEXT NOT NULL,
    conditions   JSONB NOT NULL DEFAULT '[]'::jsonb,
    metrics      JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_default   BOOLEAN NOT NULL DEFAULT false,
    is_active    BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sla_policies_org_workspace ON sla_policies(org_id, workspace_id);
CREATE INDEX idx_sla_policies_active ON sla_policies(is_active);
CREATE INDEX idx_sla_policies_default ON sla_policies(is_default) WHERE is_default = true;

COMMENT ON TABLE sla_policies IS 'Service-level agreement policies';
COMMENT ON COLUMN sla_policies.conditions IS 'Array of {field,op,value} matching conditions';
COMMENT ON COLUMN sla_policies.metrics IS 'JSON object: {first_response_hours, resolution_hours}';
COMMENT ON COLUMN sla_policies.is_default IS 'If true, applies when no other policy matches';

-- ── SLA Records ─────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS sla_records (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id          UUID NOT NULL UNIQUE,
    policy_id                UUID NOT NULL,
    org_id                   UUID NOT NULL,
    started_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    first_response_deadline  TIMESTAMPTZ,
    first_responded_at       TIMESTAMPTZ,
    resolution_deadline      TIMESTAMPTZ,
    resolved_at              TIMESTAMPTZ,
    first_response_breached  BOOLEAN NOT NULL DEFAULT false,
    resolution_breached      BOOLEAN NOT NULL DEFAULT false,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_sla_records_conversation ON sla_records(conversation_id);
CREATE INDEX idx_sla_records_org ON sla_records(org_id);
CREATE INDEX idx_sla_records_policy ON sla_records(policy_id);
CREATE INDEX idx_sla_records_breaching ON sla_records(first_response_deadline, first_responded_at)
    WHERE first_response_breached = false AND first_responded_at IS NULL;
CREATE INDEX idx_sla_records_resolution ON sla_records(resolution_deadline, resolved_at)
    WHERE resolution_breached = false AND resolved_at IS NULL;

COMMENT ON TABLE sla_records IS 'SLA tracking for individual conversations';
COMMENT ON COLUMN sla_records.conversation_id IS 'One record per conversation';
COMMENT ON COLUMN sla_records.first_response_breached IS 'Set to true when deadline passes without response';
COMMENT ON COLUMN sla_records.resolution_breached IS 'Set to true when deadline passes without resolution';

-- ── Foreign Keys ────────────────────────────────────────────────────────────

-- Note: Uncomment these when conversation table exists in this migration sequence
-- ALTER TABLE sla_records ADD CONSTRAINT fk_sla_records_conversation
--     FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE;

-- ALTER TABLE sla_records ADD CONSTRAINT fk_sla_records_policy
--     FOREIGN KEY (policy_id) REFERENCES sla_policies(id) ON DELETE RESTRICT;

-- ============================================================================
-- End Migration
-- ============================================================================
