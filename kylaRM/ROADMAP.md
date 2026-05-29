# Kyla CX — Platform Overhaul Roadmap
> Customer Operations OS Architecture · Go · gRPC · React · Flutter

---

## Current State Assessment

### What exists today (2026-05-26)

| Area | Status | Notes |
|------|--------|-------|
| Go module (`kyla-be`) | ✅ Working monolith | Single gRPC + Gin server; domain packages under `internal/` |
| Proto definitions (`pkg/pb/`) | ✅ Extensive | 150+ pb files: tickets, calls, CRM, billing, campaigns, KB, AI, etc. |
| Auth / RBAC / Org / User | ✅ Implemented | JWT, Firebase, WebAuthn, role-based access |
| **Casbin RBAC enforcement** | ✅ Added (not in original plan) | `internal/casbin/` — 603-line routes table, seeder, enforcer, policy model |
| **Audit interceptor** | ✅ Added (not in original plan) | `internal/audit/` — captures every gRPC call into `audit_logs` table |
| Contacts / Contact Groups / Tags / Labels | ✅ Implemented | With sharing, groups, tags |
| Branches / Departments / Teams | ✅ Implemented | Multi-org hierarchy |
| Shifts / Leaves / Breaks | ✅ Implemented | Workforce management basics |
| Apps / Sharing / Invitations | ✅ Implemented | App credentials, resource sharing |
| Onboarding | ✅ Implemented | REST + gRPC handler, workspace-template-driven |
| **NATS JetStream** | ✅ Live | In `docker-compose.yaml`; `internal/nats/` client + `EventPublisher`; `events.StreamBus` for realtime |
| **Workspace + members + seeder** | ✅ Implemented | `internal/workspace/` with domain template seeder |
| **Object Core (types, objects, relations, events, views)** | ✅ Implemented | `internal/objectcore/` — schema mgmt, CRUD, relations, timeline; views server + store |
| **Communication (Conversation/Message + 5 channels)** | ✅ Implemented | `internal/communication/` — WhatsApp Cloud, Email, SMS (Twilio + AT), Voice, WebChat adapters; routing engine; SLA engine; realtime streaming via NATS |
| **Webhook system (inbound + outbound)** | ✅ Implemented | WA verify/receive, Twilio + Africa's Talking SMS receivers; outbound webhook registration via `internal/apps/webhook*` |
| **CRM (pipelines, stages, deals)** | ✅ Implemented | `internal/crm/` — deals as Object Core records, JSONB stage moves |
| **Ticketing (rooms, messages, macros)** | ✅ Implemented | `internal/ticketing/` |
| **Knowledge Base (categories, articles)** | ✅ Implemented | `internal/knowledge/` |
| **Forms (forms + public submissions)** | ✅ Implemented | `internal/forms/` — public `SubmitForm` for unauthenticated visitors |
| **Projects / Tasks** | ✅ Implemented | `internal/projects/` |
| Telephony service | ⏳ Not started | Empty placeholder dirs (`internal/telephony/{events,model,server,store}`); pb stubs exist |
| Automation / Workflow engine | 🚧 Scaffolded only | `internal/automation/workflow.go` — 88-line type/model definitions; engine/server/store dirs empty; no Temporal integration yet |
| AI engine | ⏳ Not started | Empty placeholder dirs (`internal/ai/{server,store}`) |
| Campaigns | ⏳ Not started | Empty placeholder dirs (`internal/campaigns/{events,model,server,store}`) |
| Analytics | ⏳ Not started | Empty placeholder dirs |
| Billing | ⏳ Not started | Empty placeholder dirs; protos exist |
| Notification service | ⏳ Not started | Empty placeholder dirs |
| Temporal | ⏳ Not started | Not in `docker-compose.yaml`; no SDK import |
| OpenSearch / ClickHouse | ⏳ Not started | — |

### Critical architectural problems

1. **Flat service directory** — `pkg/service/` holds 60+ files across all domains with no bounded context separation. Adding more will make it unmaintainable.

2. **One God gRPC server** — Everything registered on a single server in `main.go`. Impossible to scale individual domains independently.

3. **No event backbone** — AWS SQS used for ad-hoc responses only. Inter-service communication has no structured event contract.

4. **No Workspace primitive** — The product models org → branch → department → team, but there is no "Workspace" concept that maps to the user-facing domain model (Sales, Support, etc).

5. **Missing Object Core** — There is no flexible, schema-driven object engine. Every entity is a hardcoded GORM model, making the "composable platform" vision impossible without significant refactoring.

6. **Proto sprawl** — All 150+ proto definitions are compiled into a single flat `pkg/pb/` directory. No grouping by domain, impossible to publish individual SDK packages.

---

## Target Architecture

### Multi-tenant hierarchy (refined)

```
Platform (Kyla)
  └── Organization  (Tenant — "Acme Ltd")
        ├── Workspaces
        │     ├── Sales Workspace
        │     │     ├── Spaces / Views
        │     │     ├── Objects (Leads, Deals, Contacts...)
        │     │     ├── Automations
        │     │     └── Members + Permissions
        │     ├── Support Workspace
        │     └── Marketing Workspace
        ├── Billing
        └── Platform Config (RBAC, Apps, Integrations)
```

> A **Workspace** replaces what is currently modeled as a Branch in the context of product usage.
> Branches remain as physical/regional entities. Workspaces are logical product domains.

---

### Service Map (Target)

```
kylaCX/
  kylaBE/                          ← Monorepo root (Go workspace or modules)
    services/
      identity/                    ← Auth, Users, Orgs, RBAC, Invitations, Apps
      workspace/                   ← Workspaces, Members, Permissions, Onboarding
      object-core/                 ← Dynamic schemas, Relations, CRUD, Versioning
      communication/               ← Inbox, Channels (WA/Email/SMS/Voice/Chat)
      crm/                         ← Contacts, Companies, Deals, Leads, Pipelines
      ticketing/                   ← Tickets, SLAs, Macros, Scripts, Queues
      telephony/                   ← VoIP, IVR, Call sessions, Recording
      campaigns/                   ← WhatsApp/SMS campaigns, Autodialer
      automation/                  ← Workflow engine, Business rules, Schedules
      knowledge/                   ← Knowledge base, FAQ, Articles
      projects/                    ← Tasks, Projects, Approvals
      analytics/                   ← Reporting, Dashboards, ClickHouse bridge
      ai/                          ← RAG, Summarization, Classification, Copilot
      billing/                     ← Plans, Wallets, Subscriptions, Payments
      notification/                ← Push, Email, In-app notifications
      realtime/                    ← NATS event bus, Presence, Streaming gateway
    gateway/                       ← Envoy + gRPC-Web HTTP/2 gateway
    shared/                        ← Shared proto definitions, interceptors, utils
    proto/                         ← All .proto source files, organized by domain
      identity/
      workspace/
      crm/
      ticketing/
      telephony/
      ...
    pkg/                           ← Generated pb Go code (domain-namespaced)
```

### Go Module Strategy

**Option A (Recommended short-term) — Go Workspace with sub-modules**

```
kylaBE/go.work                     ← Go workspace file
  services/identity/go.mod         ← module kyla-be/identity
  services/workspace/go.mod        ← module kyla-be/workspace
  services/crm/go.mod
  shared/go.mod                    ← module kyla-be/shared
  gateway/go.mod
```

Each service is an independent Go module but lives in the same repository.
They share `kyla-be/shared` for interceptors, proto types, event contracts, and utilities.

**Option B — Monorepo, single Go module, domain-structured packages**

Keep `module kyla-be`, but restructure into:

```go
kyla-be/internal/identity/...
kyla-be/internal/workspace/...
kyla-be/internal/crm/...
kyla-be/internal/ticketing/...
```

Each domain has its own:
```
internal/crm/
  model/      ← domain models (Go structs, not tied to pb)
  store/      ← data access (PostgreSQL via GORM or pgx)
  server/     ← gRPC server implementation
  handler/    ← REST/HTTP handlers if needed
  events/     ← event publishers / consumers
```

> **Recommendation: Start with Option B** (single module, internal domain packages).
> It lets you refactor incrementally without changing build tooling.
> Migrate to multi-module Go workspace when team grows or deployments need independent scaling.

---

## Object Core Design (The Platform Heart)

The most important architectural piece. Everything else depends on it.

### What it is

A dynamic entity engine that allows objects to be defined by schema, not by hardcoded Go structs.

### PostgreSQL Schema

```sql
-- Object type registry
CREATE TABLE object_types (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id      UUID NOT NULL REFERENCES organisations(id),
  workspace_id UUID REFERENCES workspaces(id),
  slug        TEXT NOT NULL,           -- "contact", "deal", "ticket", "lead"
  name        TEXT NOT NULL,
  icon        TEXT,
  is_system   BOOLEAN DEFAULT false,   -- system types cannot be deleted
  schema      JSONB NOT NULL DEFAULT '{}', -- field definitions
  created_at  TIMESTAMPTZ DEFAULT now(),
  updated_at  TIMESTAMPTZ DEFAULT now(),
  UNIQUE(org_id, slug)
);

-- Object records
CREATE TABLE objects (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id       UUID NOT NULL REFERENCES organisations(id),
  workspace_id UUID REFERENCES workspaces(id),
  type_slug    TEXT NOT NULL,
  data         JSONB NOT NULL DEFAULT '{}',  -- flexible field values
  created_by   UUID REFERENCES users(id),
  created_at   TIMESTAMPTZ DEFAULT now(),
  updated_at   TIMESTAMPTZ DEFAULT now()
);

-- Relations between objects
CREATE TABLE object_relations (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id     UUID NOT NULL,
  from_id    UUID NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
  to_id      UUID NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
  relation   TEXT NOT NULL,  -- "contact_of", "related_to", "child_of"
  created_at TIMESTAMPTZ DEFAULT now()
);

-- Object activity / timeline
CREATE TABLE object_events (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id     UUID NOT NULL,
  object_id  UUID NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
  actor_id   UUID REFERENCES users(id),
  event_type TEXT NOT NULL,   -- "created", "updated", "commented", "called"
  payload    JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ DEFAULT now()
);
```

### Object Core gRPC Service

```protobuf
service ObjectCoreService {
  // Schema management
  rpc CreateObjectType(CreateObjectTypeRequest) returns (ObjectType);
  rpc GetObjectType(GetObjectTypeRequest) returns (ObjectType);
  rpc ListObjectTypes(ListObjectTypesRequest) returns (ListObjectTypesResponse);
  rpc UpdateObjectSchema(UpdateObjectSchemaRequest) returns (ObjectType);

  // CRUD on records
  rpc CreateObject(CreateObjectRequest) returns (Object);
  rpc GetObject(GetObjectRequest) returns (Object);
  rpc ListObjects(ListObjectsRequest) returns (ListObjectsResponse);
  rpc UpdateObject(UpdateObjectRequest) returns (Object);
  rpc DeleteObject(DeleteObjectRequest) returns (DeleteObjectResponse);

  // Relations
  rpc LinkObjects(LinkObjectsRequest) returns (ObjectRelation);
  rpc UnlinkObjects(UnlinkObjectsRequest) returns (UnlinkResponse);
  rpc GetObjectRelations(GetObjectRelationsRequest) returns (ObjectRelationsResponse);

  // Timeline / Activity
  rpc GetObjectTimeline(GetObjectTimelineRequest) returns (stream ObjectEvent);
}
```

> System object types (Contact, Deal, Ticket, Lead...) will be seeded at org creation time.
> Custom object types can be defined per workspace by any admin.

---

## Event Bus Design

### Technology: NATS JetStream

NATS is lightweight, embeddable, and handles both pub/sub and persistent streams.
Preferred over Kafka for this scale — simpler ops, still durable.

### Event contract (shared Go struct + proto)

```go
// shared/events/event.go
type DomainEvent struct {
    ID          string          `json:"id"`
    OrgID       string          `json:"org_id"`
    WorkspaceID string          `json:"workspace_id"`
    Subject     string          `json:"subject"`   // "ticket.created", "deal.updated"
    ActorID     string          `json:"actor_id"`
    Payload     json.RawMessage `json:"payload"`
    OccurredAt  time.Time       `json:"occurred_at"`
}
```

### NATS subject namespacing

```
kyla.{org_id}.ticket.created
kyla.{org_id}.deal.stage_changed
kyla.{org_id}.conversation.message_received
kyla.{org_id}.contact.created
kyla.{org_id}.workflow.triggered
kyla.{org_id}.call.ended
```

Consumers:
- **Automation engine** — subscribes to all events, evaluates triggers
- **Notification service** — sends push/email/in-app alerts
- **Analytics** — writes to ClickHouse
- **Realtime gateway** — fans out to connected WebSocket/gRPC streams
- **AI engine** — triggers summarization, classification pipelines

---

## Workspace Service Design

### PostgreSQL Schema

```sql
CREATE TABLE workspaces (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id       UUID NOT NULL REFERENCES organisations(id),
  name         TEXT NOT NULL,
  slug         TEXT NOT NULL,
  icon         TEXT,
  domain_template TEXT,   -- "sales", "support", "marketing", "operations", "custom"
  config       JSONB NOT NULL DEFAULT '{}',
  created_at   TIMESTAMPTZ DEFAULT now(),
  UNIQUE(org_id, slug)
);

CREATE TABLE workspace_members (
  workspace_id UUID NOT NULL REFERENCES workspaces(id),
  user_id      UUID NOT NULL REFERENCES users(id),
  role         TEXT NOT NULL DEFAULT 'member',  -- "owner", "admin", "member", "guest"
  joined_at    TIMESTAMPTZ DEFAULT now(),
  PRIMARY KEY (workspace_id, user_id)
);
```

### Domain Templates (seeded on workspace creation)

When an org creates a "Support" workspace, the system seeds:
- Object types: `Ticket`, `Conversation`, `SLA`, `Knowledge Article`
- Views: `All Tickets`, `My Open Tickets`, `SLA Breaching`
- Default automations: SLA timer start, AI categorization, CSAT survey
- Inbox configuration: connect channels

This replaces the current hardcoded `onboarding_handlers.go` and `organisation_templates.go` with a data-driven template system.

---

## Automation / Workflow Engine Design

### Node model

```
Trigger → [Filter Conditions] → Actions[]
```

### PostgreSQL Schema

```sql
CREATE TABLE workflows (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id       UUID NOT NULL,
  workspace_id UUID NOT NULL,
  name         TEXT NOT NULL,
  trigger      JSONB NOT NULL,   -- {type: "object.created", object_type: "ticket"}
  conditions   JSONB NOT NULL DEFAULT '[]',
  actions      JSONB NOT NULL DEFAULT '[]',
  is_active    BOOLEAN DEFAULT true,
  created_at   TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE workflow_runs (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workflow_id  UUID NOT NULL REFERENCES workflows(id),
  trigger_event_id TEXT,
  status       TEXT NOT NULL,   -- "pending", "running", "success", "failed"
  context      JSONB,
  error        TEXT,
  started_at   TIMESTAMPTZ,
  finished_at  TIMESTAMPTZ
);
```

### Engine architecture in Go

```go
// internal/automation/engine/engine.go

type Engine struct {
    eventBus    events.Bus
    workflowRepo WorkflowRepository
    actionRunner ActionRunner
}

func (e *Engine) Start(ctx context.Context) {
    // Subscribe to all org events from NATS
    e.eventBus.Subscribe("kyla.*.>", e.handleEvent)
}

func (e *Engine) handleEvent(event DomainEvent) {
    workflows := e.workflowRepo.FindMatchingWorkflows(event)
    for _, wf := range workflows {
        go e.executeWorkflow(wf, event)
    }
}
```

### Built-in Action types

| Action ID | Description |
|-----------|-------------|
| `update_object` | Update a field on any object |
| `create_object` | Create a related object |
| `send_message` | Send WhatsApp / SMS / Email |
| `assign_user` | Assign to agent / team |
| `create_task` | Create a task object |
| `invoke_webhook` | HTTP POST to external URL |
| `run_ai_skill` | Call AI engine (classify, summarize, reply) |
| `start_workflow` | Chain into another workflow |
| `send_notification` | In-app / push notification |
| `delay` | Time delay node |
| `set_sla` | Set/reset SLA timer |

---

## Inbox / Conversation Model

All inbound and outbound communication becomes a **Conversation object**.

### Core model

```sql
CREATE TABLE conversations (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id         UUID NOT NULL,
  workspace_id   UUID NOT NULL,
  channel        TEXT NOT NULL,   -- "whatsapp", "email", "sms", "voice", "webchat"
  channel_ref    TEXT,            -- external ID (WA thread ID, email thread, etc.)
  contact_id     UUID REFERENCES objects(id),
  assigned_to    UUID REFERENCES users(id),
  team_id        UUID,
  status         TEXT DEFAULT 'open',  -- "open", "pending", "resolved", "snoozed"
  priority       TEXT DEFAULT 'normal',
  sla_deadline   TIMESTAMPTZ,
  meta           JSONB DEFAULT '{}',
  created_at     TIMESTAMPTZ DEFAULT now(),
  updated_at     TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE messages (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  conversation_id UUID NOT NULL REFERENCES conversations(id),
  sender_id       UUID,            -- user or contact ID
  sender_type     TEXT NOT NULL,   -- "agent", "contact", "bot", "system"
  channel         TEXT NOT NULL,
  content_type    TEXT NOT NULL,   -- "text", "image", "audio", "template"
  content         JSONB NOT NULL,
  status          TEXT DEFAULT 'sent',  -- "pending", "sent", "delivered", "read", "failed"
  created_at      TIMESTAMPTZ DEFAULT now()
);
```

---

## Phased Implementation Roadmap

---

### Phase 0 — Structural Refactoring (Weeks 1–3)  ✅ COMPLETE
*No new features. Clean the foundation.*

> **Status:** All domain packages live under `internal/`. NATS JetStream is in `docker-compose.yaml`. `internal/nats/` provides client + `EventPublisher`. `cmd/server/main.go` wires every implemented service.

**Goal:** Restructure `kylaBE` without breaking what works.

**Tasks:**

1. **Directory restructure** — Move existing service layer into domain packages:
   ```
   pkg/service/*.go  →  internal/identity/
                         internal/workspace/
                         internal/crm/
                         internal/ticketing/
                         internal/telephony/
   ```

2. **Proto reorganization** — Group `pkg/pb/` files by domain. Add domain namespace prefix to imports in future proto compilations:
   ```
   pb/identity/    ← auth, user, org, rbac, invitation, app
   pb/workspace/   ← workspace, onboarding, sharing, tag, label
   pb/crm/         ← contact, contact_groups, deal, lead, pipeline
   pb/ticketing/   ← ticket, sla, macro, thread, rooms, analytics
   pb/telephony/   ← call_*, sip_*, ivr_*, trunk_*
   pb/campaigns/   ← whatsapp_campaigns, sms, autodialer
   pb/knowledge/   ← knowledge_base, faq, category
   pb/ai/          ← rag_agents, virtual_agents, summarization, classification
   pb/billing/     ← billing_*
   pb/automation/  ← automation, business_rule, flows
   pb/comms/       ← messaging, chatdesk_messaging, whatsapp_messaging
   pb/projects/    ← task, project
   pb/notification/← notification, notification_template
   ```

3. **Add `internal/` convention** — All domain packages go under `internal/`. Shared utilities, interceptors, and event contracts go into `shared/`.

4. **NATS setup** — Add NATS JetStream to `docker-compose.yaml`. Add NATS client to shared package.

5. **Wire `cmd/server/main.go`** — Split into `cmd/server/main.go` (keeps working) so existing functionality is unaffected.

**Deliverables:**
- Same binary, same API surface, refactored package structure
- NATS running in Docker Compose
- CI passes (go build, go vet)

---

### Phase 1 — Workspace + Identity Hardening (Weeks 4–8)  ✅ COMPLETE
*The core platform layer that everything else sits on.*

> **Status:** Workspace CRUD, members, and domain-template seeder shipped (`internal/workspace/`). Onboarding rewired to be template-driven. Casbin enforcer + 603-line route policy table (`internal/casbin/`) gives per-route RBAC beyond the original plan. Audit interceptor (`internal/audit/`) logs every gRPC call. **Open:** scoped API tokens per workspace (webhook registration is done via `internal/apps/webhook*`, but token scoping is still org-level).

**Goal:** Implement the Workspace primitive and harden the RBAC / permission system.

**Tasks:**

1. **Workspace service** (`internal/workspace/`)
   - `workspaces` table + CRUD gRPC endpoints
   - `workspace_members` with roles
   - Domain template seeder (Support, Sales, Marketing, Operations templates)
   - Replace Branch-as-product-unit with proper Workspace model
   - Keep Branch as a physical/org-hierarchy concept

2. **Refine onboarding flow**
   - Org creation → default workspace creation → domain template applied → initial members invited
   - Replace current `organisation_templates.go` stub with data-driven template system

3. **RBAC enhancements**
   - Permissions scoped to workspace (not just org)
   - Permission matrix: Org → Workspace → Object type → Record
   - Audit log table for all permission checks

4. **API App / SDK token model**
   - Scoped API tokens per workspace
   - Webhook registration (for marketplace extensibility)

**Deliverables:**
- Workspace API live and tested
- Updated onboarding flow
- Workspace-scoped RBAC

---

### Phase 2 — Object Core Engine (Weeks 9–14)  ✅ COMPLETE
*The single most impactful architectural piece.*

> **Status:** Object Core ships in `internal/objectcore/` with `object_types`, `objects`, `object_relations`, `object_events`, and a views layer (`view.go`, `view_server.go`, `view_store.go`). System types are seeded per workspace template. Contacts/CRM/Ticketing/Forms now persist through Object Core. **Open:** OpenSearch full-text indexing on `object.created`/`object.updated` is still not wired.

**Goal:** Ship a working dynamic object engine that replaces hardcoded entity models.

**Tasks:**

1. **Object Core service** (`internal/object-core/`)
   - `object_types`, `objects`, `object_relations`, `object_events` tables
   - Full gRPC API (schema management + CRUD + relations + timeline)
   - System object types seeder per workspace template:
     - Support: Ticket, Conversation, Customer, SLA, Knowledge Article
     - Sales: Lead, Deal, Contact, Company, Activity
     - Marketing: Campaign, Audience, Form, Journey
     - Operations: Task, Project, Process, File

2. **Migrate Contacts and Teams to Object Core**
   - `Contact` becomes a system object type backed by Object Core
   - Legacy contact GORM model remains as a read-projection for backward compat
   - Event: `contact.created` → NATS → downstream consumers

3. **Custom fields**
   - Schema definition for custom fields on any object type
   - Supported field types: text, number, date, select, multi-select, user, relation, file

4. **Object search**
   - OpenSearch integration for full-text search across all object records
   - Index updated on every `object.created` / `object.updated` NATS event

5. **Views engine (server-side)**
   - Saved views: filter + sort + columns configuration stored in DB
   - Views belong to workspace and are shareable by role

**Deliverables:**
- Object Core service fully operational
- Contacts migrated to Object Core (backward compat kept)
- Full-text search working
- Views CRUD API

---

### Phase 3 — Inbox + Communication Layer (Weeks 15–22)  ✅ SUBSTANTIALLY COMPLETE
*The most user-facing and revenue-critical domain.*

> **Status:** `internal/communication/` ships Conversation/Message models, all five channel adapters (WhatsApp Cloud, Email IMAP/SMTP, SMS via Twilio + Africa's Talking, Voice, WebChat), an `AdapterRegistry` for outbound dispatch, inbound webhook handlers with tenant resolution, a routing engine with round-robin, an SLA engine driven by `events.Publisher`, and realtime streaming via `events.StreamBus` (NATS-backed). **Open:** skill-based routing beyond round-robin; web-chat React widget; full VoIP wiring (depends on Phase 5).

**Goal:** Unified inbox with multi-channel conversation management.

**Tasks:**

1. **Conversation + Message model** (`internal/communication/`)
   - `conversations` and `messages` tables
   - Conversation is an Object Core type → gets full timeline, relations, custom fields
   - Inbox gRPC service: list conversations, assign, update status, reply

2. **Channel integrations** (one at a time, in order of priority):
   - **WhatsApp Business API** — inbound webhook → conversation created → NATS event → inbox
   - **Email** — SMTP/IMAP bridge or Resend + Mailgun receive → conversation
   - **Web chat widget** — React embeddable widget using gRPC-Web streaming
   - **SMS** — via existing SMS proto/client
   - **VoIP** — wire existing telephony pb into communication layer (call → conversation)

3. **Realtime streaming**
   - NATS → gRPC server-side streaming for inbox live updates
   - Typing indicators, presence, message delivery status via NATS

4. **Channel routing**
   - Rule-based routing: route conversations to team/agent based on channel, keyword, contact segments
   - Round-robin and skill-based routing

5. **SLA engine**
   - SLA policies per workspace
   - Timer starts on conversation created, resets on reply
   - NATS event `sla.breaching` → automation trigger → notification

**Deliverables:**
- WhatsApp + Email inbox working end-to-end
- Real-time message streaming to React frontend
- SLA engine live

---

### Phase 4 — CRM + Ticketing (Weeks 23–30)  ✅ COMPLETE
*The structured customer operations layer.*

> **Status:** `internal/crm/` (pipelines, stages, deals as Object Core records — `MoveDeal` patches `data->>'stage_id'` via raw SQL); `internal/ticketing/` (rooms, threaded messages, macros); `internal/knowledge/` (categories + articles); `internal/forms/` (forms + public unauthenticated `SubmitForm`); `internal/projects/` (tasks + projects). Schema in `0006_phase4.sql` + `0007_projects.sql`. **Open:** Kanban/forecast pipeline views in frontend; CSAT-on-close automation (blocked on Phase 6); AI ticket auto-categorisation (blocked on Phase 6).

**Goal:** Implement the full CRM and Ticketing modules using Object Core as the foundation.

**Tasks:**

1. **CRM** (`internal/crm/`)
   - Lead, Deal, Company, Activity as Object Core system types
   - Pipeline service: stages, drag-drop ordering, probability weighting
   - Pipeline views (Kanban, Table, Forecast)
   - Deal-to-Conversation relation: link calls, emails, WhatsApp threads to deals

2. **Ticketing** (`internal/ticketing/`)
   - Ticket as Object Core type; all custom fields via Object Core schema
   - SLA assignment, escalation rules, macro execution
   - Ticket rooms: threaded internal notes + customer-facing replies
   - Assignment rules engine (leverage automation engine)
   - CSAT survey trigger on ticket close (automation)

3. **Knowledge Base** (`internal/knowledge/`)
   - Article CRUD, categories, search
   - AI-powered article suggestion when agent opens a ticket (Phase 4 AI stretch)
   - Customer-facing portal view

4. **Forms** (`internal/forms/`)
   - Form builder: field types, logic branching
   - Submission creates Object Core record of type `Form Submission`
   - Workflow triggers fire on submission

5. **Projects and Tasks** (`internal/projects/`)
   - Task as Object Core type
   - Project grouping, status boards (Kanban), timeline view
   - Task assignees, due dates, sub-tasks, comments

**Deliverables:**
- Full CRM pipeline usable by sales teams
- Ticketing covering existing platform customers
- Knowledge base with article editor
- Form builder + submission processing

---

### Phase 5 — Telephony Integration (Weeks 31–36)  🚧 SLICES 5A/5B/5C SHIPPED (mod_xml_curl + production FS config + slices 5d/5e remain)
*Self-hosted SIP via FreeSWITCH with WebRTC softphone support.*

> **Status (2026-05-29):** Architecture choice locked in: self-hosted SIP from day one (FreeSWITCH + coturn). Backend foundation shipped:
>
> **Infrastructure** (`deploy/docker-compose.yaml`): `freeswitch` (signalwire/freeswitch:1.10) exposes SIP/UDP 5060, SIP-over-WSS 7443 for browser softphones, ESL 8021 for Go control, RTP 16384-16484; `coturn` (4.6) exposes STUN/TURN 3478 + TURN/TLS 5349 + relay range 49160-49200.
>
> **Service layer** (`internal/telephony/`): `0010_telephony.sql` migration adds `calls` (id = FreeSWITCH UUID), `call_events` (per-call timeline), `sip_domains` (one default per org via partial unique index), `sip_extensions` (one per user, bcrypt-hashed SIP password), `sip_trunks` (write-only password field on the gRPC read path); new `telephony.proto` → `TelephonyService` gRPC (Originate/Hangup/Transfer/Hold/Resume + GetCall/ListCalls + AppendCallEvent/ListCallEvents + SIP admin + IssueSoftphoneToken).
>
> **PBX abstraction**: `PBXController` interface with `NoopPBX` fallback (binary boots without FS) and `FreeSWITCHController` skeleton that connects + authenticates ESL, subscribes to CHANNEL_CREATE/ANSWER/HANGUP_COMPLETE/SOFIA_REGISTER/RECORD_STOP, lifts events onto a buffered `CallEventStream`, and issues bgapi originate/uuid_kill/uuid_transfer/uuid_hold commands.
>
> **Event pipeline**: `EventBridge` consumes the ESL event stream → updates `calls` projection (ringing → answered → ended) + appends `call_events` rows + publishes `call.started`/`call.answered`/`call.ended` NATS events. The existing `communication.VoiceCallBridge` already subscribes to `call.ended` and auto-creates conversations — telephony plugs straight into the inbox.
>
> **WebRTC softphone bootstrap**: `IssueSoftphoneToken` returns an HS256 JWT signed with `JWT_SECRET_KEY` (bound to org/user/extension), the SIP-over-WSS URL, the SIP realm, and ICE servers (STUN + TURN credentials).
>
> **IVR engine (slice 5c, 2026-05-29)**: `0011_ivr.sql` adds `ivr_flows` (definition stored as JSONB so the visual builder can persist arbitrary node shapes), `ivr_did_mappings` (DID→flow routing), `ivr_runs` (per-call breadcrumb trail). `ivr.proto` → `IVRService` gRPC. `internal/telephony/ivr/executor.go` is the node walker: handles `play_audio`, `say`, `menu` (DTMF capture via play_and_get_digits), `transfer`, `record`, `hangup`, `goto`. The executor is driven by ESL events — `PlaybackStop` advances after `play_audio`/`say`; `DTMFCaptured` advances after `menu` with the captured digit fed to `node.branches[digit]`. The `telephony.IVRHook` interface decouples the EventBridge from the IVR package; `cmd/server/main.go` wires an `ivrBridgeAdapter` to satisfy it without an import cycle.
>
> **FreeSWITCH config skeleton (2026-05-29)**: `deploy/freeswitch/conf/` ships `vars.xml`, `autoload_configs/{event_socket,modules}.conf.xml`, `sip_profiles/{internal,webrtc}.xml`, `dialplan/default.xml`. docker-compose mounts these read-only over the image defaults. The internal profile listens on 5060 (UDP/TCP); the webrtc profile listens on 7443 for SIP-over-WSS with mandatory DTLS-SRTP and OPUS. The default dialplan parks calls so ESL takes over routing; echo test on extension `9999`. See `deploy/freeswitch/README.md` for the mount layout and verification steps.
>
> **Frontend softphone wiring (2026-05-29)**: `kylaFE/src/features/telephony/sip/` adds `SipClient` (SIP.js `SimpleUser` wrapper) and `useSipClient` hook (manages a module-level singleton with an off-screen `<audio>` element for remote audio). `useSipClient` fetches a softphone bootstrap via `services.telephony.issueSoftphoneToken()`, connects WSS to FreeSWITCH, REGISTERs, and mirrors SIP state into the existing softphone Zustand store. `Softphone.tsx` now dials/hangs up/mutes/sends DTMF through the SIP client instead of the legacy `useStartCallSession` mutation — the new TelephonyService creates `calls` rows from ESL events, so the frontend doesn't need to pre-create a session row. `sip.js@^0.21.2` added to `package.json` (run `pnpm install` to fetch).
>
> **Open in slice 5a/5b**: ESL bgapi job-correlation (originate currently fires command but doesn't await BACKGROUND_JOB — UUID generated locally and persisted optimistically); automatic ESL reconnect with backoff; mod_xml_curl integration for dynamic directory + dialplan + JWT validation served from the Go backend.
>
> **Open in slice 5c**: visual IVR flow builder (the `IvrFlowBuilder.tsx` scaffold exists; canvas + node palette + persistence not wired); inbound DID-to-org mapping outside of IVR (calls without a DID mapping still drop today); recording management via uuid_record.
>
> **Open in slices 5d-5e**: queues + routing engine + wallboard; recording S3 upload + transcription via the AI engine.
>
> **Open in slice 5f**: SIP admin pages (sip_trunks/sip_extensions/sip_domains gRPCs exist; UI not yet built).

**Goal:** Make VoIP a first-class communication channel in the platform.

**Tasks:**

1. **Telephony service** (`internal/telephony/`)
   - Implement the already-defined gRPC servers for: call sessions, IVR, queues, extensions, SIP, trunks, recording
   - Call → Conversation relation in Object Core
   - Call analytics → ClickHouse pipeline

2. **Agent softphone**
   - WebRTC integration in React frontend
   - Click-to-call from Contact / Deal / Ticket object pages

3. **IVR flow builder**
   - Node-based drag-drop IVR builder (similar to workflow engine nodes)
   - Backed by existing `call_ivr_flow.proto`

4. **Call recording + transcription**
   - Existing proto covers this; wire S3 upload + transcription service
   - Transcripts indexed in OpenSearch, linked to conversation

**Deliverables:**
- VoIP calls work through platform
- Calls linked to contacts and tickets
- IVR builder operational

---

### Phase 6 — Automation Engine + AI + Campaigns (Weeks 37–46)  ✅ COMPLETE
*The differentiation layer.*

> **Status (2026-05-29):** Automation engine + minimal AI + campaigns all shipped.
>
> **Automation engine** (`internal/automation/`): Temporal client + auto-setup + UI in `docker-compose.yaml`; `0008_automation.sql` migration; `workflow.proto` → `WorkflowService` gRPC (Create/Update/Get/List/Delete/GetRunHistory/TestRunWorkflow); `Store` with JSONB-indexed `FindMatchingWorkflows`; `Executor` with deterministic WorkflowID for NATS-redelivery dedup; `Consumer` on `kyla.*.>` using NATS queue group `kyla-automation` for horizontal scaling; in-process Temporal worker started from `cmd/server/main.go`. All 11 action types implemented: `delay`, `start_workflow` (inline via `workflow.Sleep` / `ExecuteChildWorkflow`), plus 9 activities (`update_object`, `assign_user`, `create_object`, `create_task`, `send_message`, `invoke_webhook`, `set_sla`, `send_notification`, `run_ai_skill`).
>
> **Minimal AI engine** (`internal/ai/`): `LLMProvider` interface with OpenAI (default) and Anthropic implementations, both via direct HTTP (no SDK deps); `NoopProvider` fallback so binary boots without keys; provider selection via `LLM_PROVIDER` env var; `AIService` gRPC (`ClassifyText`, `SummarizeText`, `GenerateReply`); in-process `ActivityAdapter` so the worker calls the provider directly without a gRPC hop. Per architectural decision: Temporal worker runs in same binary; versioning uses Temporal's `GetVersion()` patching; no JSONB definition snapshots.
>
> **Campaigns** (`internal/campaigns/`): `0009_campaigns.sql` migration adds `campaigns`, `campaign_recipients`, `whatsapp_templates` tables; `campaigns.proto` → `CampaignService` gRPC (CRUD + Launch/Pause/Cancel + ListRecipients + WhatsApp template registry); `CampaignExecutionWorkflow` resolves audience (object_query or explicit), fans out per-recipient send via `SendRecipientActivity` (reuses `communication.AdapterRegistry` so messages route through the same WA/SMS/Email/Voice/WebChat adapters used by the inbox), `FinaliseCampaignActivity` recomputes denormalised stats. Schedule modes: `immediate`, `scheduled_once` (workflow.Sleep), `recurring` (Temporal Schedules with cron). Campaigns worker is a separate Temporal worker on the same `kyla-automation` task queue — keeps a slow audience resolution from starving the automation worker pool.
>
> **Open:** Visual workflow builder in React (frontend `automation` feature dir scaffold landed but the React Flow canvas isn't built); richer AI (RAG, vector store, virtual agents) — deferred to a later AI-specific phase; autodialer campaigns (depend on Phase 5 telephony).

**Goal:** Deliver a visual workflow builder backed by Temporal for durable execution, and embedded AI capabilities.

---

#### Workflow Execution Architecture

The automation engine is split into two concerns:

| Layer | Technology | Responsibility |
|-------|-----------|----------------|
| **Event backbone** | NATS JetStream | Delivering domain events that trigger workflows |
| **Workflow execution** | Temporal | Durable, retryable execution of workflow steps |

NATS does not disappear — it remains the event bus. Temporal replaces the naive `go executeWorkflow()` goroutine with durable, observable, versioned workflow execution.

```
NATS event (e.g. ticket.created)
  └── automation/consumer.go (NATS subscriber)
        └── matches workflow definitions in DB
              └── client.ExecuteWorkflow(ctx, WorkflowRun, workflowDef, event)
                    └── Temporal Server (schedules + persists execution)
                          ├── Activities: UpdateObject, SendMessage, AssignUser...
                          ├── workflow.Sleep() for delay nodes
                          └── child workflows for start_workflow action
```

---

#### Why Temporal

| Requirement | Naive goroutine | Temporal |
|-------------|----------------|---------|
| Survive server restart mid-workflow | ❌ State lost | ✅ Durable — state replayed from history |
| Retry failed actions (HTTP/gRPC calls) | Manual | ✅ Activity retry policies, backoff, jitter |
| Delay 3 days then send email | `time.Sleep` in goroutine (broken) | ✅ `workflow.Sleep(72 * time.Hour)` — no goroutine held |
| Workflow run history + debuggability | Custom table | ✅ Temporal Web UI + history API |
| Versioned workflow logic | Not possible | ✅ `workflow.GetVersion()` patching |
| Child workflows / fan-out | goroutine leaks | ✅ `workflow.ExecuteChildWorkflow()` |
| Cron schedules | Custom cron runner | ✅ Temporal Schedules |
| Timeout enforcement on actions | Complex | ✅ Activity + workflow timeouts, deadline context |

---

#### Infrastructure additions

**Docker Compose (dev):**

```yaml
temporal:
  image: temporalio/auto-setup:1.24
  ports:
    - "7233:7233"
  environment:
    - DB=postgres12
    - DB_PORT=5432
    - POSTGRES_USER=${POSTGRES_USER}
    - POSTGRES_PWD=${POSTGRES_PASS}
    - POSTGRES_SEEDS=postgres
  depends_on:
    - postgres
  networks:
    - kyla_net

temporal-ui:
  image: temporalio/ui:2.26
  ports:
    - "8088:8080"
  environment:
    - TEMPORAL_ADDRESS=temporal:7233
  networks:
    - kyla_net
```

**New env vars:**
```
TEMPORAL_HOST_PORT=temporal:7233
TEMPORAL_NAMESPACE=kyla-default
TEMPORAL_TASK_QUEUE=kyla-automation
```

---

#### Implementation structure (`internal/automation/`)

```
internal/automation/
  model.go          ← Workflow, WorkflowRun GORM models (run history projection)
  store.go          ← FindMatchingWorkflows(event DomainEvent) []*Workflow
  server.go         ← gRPC: CreateWorkflow, UpdateWorkflow, ListWorkflows, GetRunHistory
  consumer.go       ← NATS subscriber → evaluates triggers → starts Temporal workflows
  worker.go         ← Temporal worker: registers all Workflow + Activity functions

  workflows/
    automation_workflow.go  ← Top-level Temporal Workflow function (runs all nodes in sequence)
    schedule_workflow.go    ← Temporal Schedule-backed cron workflow

  activities/
    update_object.go        ← UpdateObjectActivity
    create_object.go        ← CreateObjectActivity
    send_message.go         ← SendMessageActivity (WA/Email/SMS)
    assign_user.go          ← AssignUserActivity
    create_task.go          ← CreateTaskActivity
    invoke_webhook.go       ← InvokeWebhookActivity (HTTP POST, retries built-in)
    run_ai_skill.go         ← RunAISkillActivity (classify, summarize, reply)
    send_notification.go    ← SendNotificationActivity
    set_sla.go              ← SetSLAActivity
```

**Key contracts:**

```go
// internal/automation/workflows/automation_workflow.go

func AutomationWorkflow(ctx workflow.Context, def WorkflowDefinition, trigger events.DomainEvent) error {
    logger := workflow.GetLogger(ctx)

    for _, node := range def.Actions {
        switch node.Type {
        case "delay":
            dur := time.Duration(node.Params["duration_seconds"].(float64)) * time.Second
            _ = workflow.Sleep(ctx, dur)  // durable — survives restarts

        case "invoke_webhook":
            ao := workflow.ActivityOptions{
                StartToCloseTimeout: 30 * time.Second,
                RetryPolicy: &temporal.RetryPolicy{
                    MaximumAttempts: 5,
                    BackoffCoefficient: 2.0,
                },
            }
            _ = workflow.ExecuteActivity(workflow.WithActivityOptions(ctx, ao),
                activities.InvokeWebhookActivity, node.Params, trigger).Get(ctx, nil)

        case "run_ai_skill":
            ao := workflow.ActivityOptions{StartToCloseTimeout: 60 * time.Second}
            _ = workflow.ExecuteActivity(workflow.WithActivityOptions(ctx, ao),
                activities.RunAISkillActivity, node.Params, trigger).Get(ctx, nil)

        default:
            if err := executeStandardActivity(ctx, node, trigger); err != nil {
                logger.Error("action failed", "type", node.Type, "error", err)
                // non-fatal: continue to next node (configurable per workflow)
            }
        }
    }
    return nil
}
```

```go
// internal/automation/consumer.go

type AutomationConsumer struct {
    store    *AutomationStore
    temporal client.Client
    taskQueue string
}

func (c *AutomationConsumer) Start(ctx context.Context, bus events.Bus) error {
    return bus.Subscribe("kyla.*.>", func(event *events.DomainEvent) error {
        workflows, err := c.store.FindMatchingWorkflows(event)
        if err != nil || len(workflows) == 0 {
            return nil
        }
        for _, wf := range workflows {
            if !wf.IsActive { continue }
            _, err := c.temporal.ExecuteWorkflow(ctx,
                client.StartWorkflowOptions{
                    ID:        fmt.Sprintf("wf-%s-%s", wf.ID, event.ID),
                    TaskQueue: c.taskQueue,
                },
                AutomationWorkflow, wf.Definition(), *event,
            )
            if err != nil {
                log.Printf("[automation] start workflow %s failed: %v", wf.ID, err)
            }
        }
        return nil
    })
}
```

```go
// internal/automation/worker.go

func StartWorker(temporalClient client.Client, taskQueue string, deps ActivityDeps) {
    w := worker.New(temporalClient, taskQueue, worker.Options{})

    // Register workflows
    w.RegisterWorkflow(AutomationWorkflow)
    w.RegisterWorkflow(ScheduleWorkflow)

    // Register activities (inject domain store/client dependencies)
    w.RegisterActivity(&activities.UpdateObjectActivity{Store: deps.ObjectStore})
    w.RegisterActivity(&activities.SendMessageActivity{MessagingClient: deps.MessagingClient})
    w.RegisterActivity(&activities.AssignUserActivity{Store: deps.ObjectStore})
    w.RegisterActivity(&activities.InvokeWebhookActivity{HTTPClient: deps.HTTPClient})
    w.RegisterActivity(&activities.RunAISkillActivity{AIClient: deps.AIClient})
    w.RegisterActivity(&activities.SendNotificationActivity{NotifClient: deps.NotifClient})
    w.RegisterActivity(&activities.SetSLAActivity{Store: deps.ConversationStore})

    if err := w.Run(worker.InterruptCh()); err != nil {
        log.Fatalf("temporal worker error: %v", err)
    }
}
```

---

#### Database schema (run history projection)

```sql
-- Keep workflows table as the definition store.
-- Temporal owns the execution state; workflow_runs is a projection for our UI.
CREATE TABLE workflow_runs (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workflow_id     UUID NOT NULL REFERENCES workflows(id),
  temporal_run_id TEXT NOT NULL,          -- Temporal's RunID for deep-link to UI
  trigger_event_id TEXT,
  status          TEXT NOT NULL,          -- mirror of Temporal status for fast queries
  context         JSONB,
  error           TEXT,
  started_at      TIMESTAMPTZ,
  finished_at     TIMESTAMPTZ
);
```

> `temporal_run_id` lets the Kyla UI deep-link directly into Temporal Web for detailed workflow traces.

---

#### Built-in Action types

| Action ID | Description | Temporal notes |
|-----------|-------------|----------------|
| `update_object` | Update a field on any object | Synchronous activity |
| `create_object` | Create a related object | Synchronous activity |
| `send_message` | Send WhatsApp / SMS / Email | Async activity with 5-retry policy |
| `assign_user` | Assign to agent / team | Synchronous activity |
| `create_task` | Create a task object | Synchronous activity |
| `invoke_webhook` | HTTP POST to external URL | Async activity, exponential backoff, 5 retries |
| `run_ai_skill` | Call AI engine (classify, summarize, reply) | Async activity, 60s timeout |
| `start_workflow` | Chain into another workflow | `ExecuteChildWorkflow` |
| `send_notification` | In-app / push notification | Synchronous activity |
| `delay` | Time delay between nodes | `workflow.Sleep(duration)` — durable |
| `set_sla` | Set/reset SLA timer | Synchronous activity |
| `wait_for_event` | Pause until a NATS event matches | `workflow.GetSignalChannel` |
| `schedule` | Cron-triggered workflow | Temporal Schedule API |

---

**Tasks:**

1. **Automation engine** (`internal/automation/`)
   - NATS consumer that evaluates workflow trigger conditions
   - Temporal Workflow function (`AutomationWorkflow`) that executes node sequences
   - All built-in action types implemented as Temporal Activities
   - Temporal worker registered in `cmd/server/main.go` (runs in same binary for Phase 6; can be split later)
   - `workflow_runs` projection table updated via Temporal lifecycle hooks

2. **Temporal infrastructure**
   - Temporal server + UI added to `docker-compose.yaml`
   - Temporal namespace provisioned on startup
   - Go SDK: `go.temporal.io/sdk@v1.x`

3. **Visual workflow builder** (React)
   - Drag-drop node canvas (React Flow)
   - Trigger picker, condition builder, action configurator
   - Test-run button → calls `StartWorkflow` directly with a synthetic test event
   - Run history viewer with link to Temporal Web for detailed traces

4. **AI Engine** (`internal/ai/`)
   - Wire existing RAG agents, summarization, classification protobufs
   - `RunAISkillActivity` calls AI gRPC service internally
   - Ticket auto-categorization on creation
   - Conversation summarization on resolve
   - AI reply suggestion in inbox (based on KB + history)
   - Copilot chat panel (ask questions about your data)

5. **Campaigns** (`internal/campaigns/`)
   - WhatsApp campaign builder: template selection, audience segment, schedule
   - Campaigns use Temporal Schedules for timed sends
   - SMS campaign
   - Autodialer campaign
   - Campaign responses → conversations created → inbox

**Deliverables:**
- Temporal server running in Docker Compose
- Automation engine processing events end-to-end with durable execution
- Delay nodes survive server restarts
- AI reply suggestions live in inbox
- Campaign manager with WhatsApp + SMS

---

### Phase 7 — Analytics + Billing (Weeks 47–54)  ⏳ NOT STARTED
*Revenue instrumentation and data visibility.*

> **Status:** `internal/analytics/{server,store}` and `internal/billing/{events,model,server,store}` are empty placeholder dirs. ClickHouse not in compose. Billing protos exist in `pkg/pb/` but no server implements them yet.

**Goal:** Actionable reporting and subscription billing.

**Tasks:**

1. **Analytics pipeline**
   - ClickHouse setup for event-sourced analytics
   - All NATS events forwarded to ClickHouse consumer
   - Pre-built reports: ticket volume, SLA adherence, call analytics, conversion rates

2. **Dashboard builder**
   - Drag-drop dashboard with chart widgets
   - Charts backed by ClickHouse queries or PostgreSQL aggregates
   - Per-workspace dashboards, shareable

3. **Billing service** (`internal/billing/`)
   - Implement the extensively defined billing protobufs: accounts, wallets, subscriptions, payment methods, transactions
   - Integrate with payment provider (Stripe / Paystack depending on market)
   - Usage-based billing hooks: count conversations, users, AI tokens

**Deliverables:**
- Basic analytics dashboard live
- Subscription billing for multi-tenant

---

### Phase 8 — Marketplace + SDK (Weeks 55–65)  🚧 PARTIAL (webhooks only)
*Platform extensibility — the long-term moat.*

> **Status:** Outbound webhook registration is live via `internal/apps/webhook.go`, `webhook_handler.go`, `webhook_store.go` — REST endpoints `/api/v1/webhooks` for CRUD. Public REST API surface, OAuth 2.0 for third-party apps, npm SDK, developer portal — all pending.

**Goal:** Enable third-party developers to build on Kyla.

**Tasks:**

1. **Public REST API**
   - gRPC-Gateway transcoding or hand-crafted REST wrappers on critical endpoints
   - API versioning (`v1/`)
   - OAuth 2.0 for third-party app authorization

2. **Webhook system**
   - Webhook registration per org/workspace
   - NATS consumer → HTTP delivery with retry logic
   - Delivery logs in UI

3. **App SDK** (TypeScript + Go client libraries)
   - Published npm package (`@kyla/sdk`) for web integrations
   - Go client package

4. **Extension points (Phase 1)**
   - Sidebar widget: iframed app panel in conversation/object view
   - Workflow action node: custom webhook-backed action
   - Object panel: custom tab on object record page

5. **Developer portal**
   - App registration, OAuth credential management
   - Webhook testing console
   - Documentation site (auto-generated from protos)

**Deliverables:**
- Public API stable and documented
- Webhook system live
- 3 reference integrations (Slack, Google Sheets, Zapier)

---

## Immediate Next Steps (This Sprint)

> Phases 0–4 are complete. The next decision is **what to start next: Phase 5 (Telephony) or Phase 6 (Automation/AI)?**
>
> The strategic argument for **Phase 6 first**: every shipped Phase 3/4 module currently publishes NATS events with no consumer wired up to them — SLA breaches, ticket creation, deal stage changes, form submissions. Without the automation engine, those events go nowhere. CSAT-on-close, AI ticket categorisation, and campaign send-throttling all unblock the moment Temporal is in place. Telephony, by contrast, is a self-contained track that doesn't unblock anything else.
>
> The strategic argument for **Phase 5 first**: telephony pb is the largest unimplemented surface area (~30+ proto files) and is what enterprise customers ask for in pilots. It can be parallelised with another developer.
>
> **Recommendation:** Phase 6 first, then Phase 5 in parallel once the Temporal worker is registered.

### Week 1 actions (Phase 6 kickoff)

1. **Add Temporal + Temporal UI** to `deploy/docker-compose.yaml` (see Phase 6 compose snippet above)
2. **Add `go.temporal.io/sdk` to `go.mod`** and the three `TEMPORAL_*` env vars to `config/config.go`
3. **Create `0008_automation.sql` migration** — `workflows` + `workflow_runs` tables (the latter as a projection of Temporal state)
4. **Implement `internal/automation/store/`** — `FindMatchingWorkflows(event)`, CRUD on `workflows`
5. **Implement `internal/automation/server/`** — gRPC `CreateWorkflow`, `UpdateWorkflow`, `ListWorkflows`, `GetRunHistory`; register in `main.go`
6. **Implement `internal/automation/engine/consumer.go`** — NATS subscriber on `kyla.*.>` that evaluates triggers and calls `client.ExecuteWorkflow`
7. **Implement `internal/automation/engine/worker.go`** + the `workflows/` and `activities/` sub-packages; start with `update_object`, `send_message`, `assign_user`, `delay`, `invoke_webhook`
8. **Define `automation.proto`** in `kylaPB/` and run `make proto-go`

### Architectural decisions (locked 2026-05-26)

- **Worker topology:** Temporal worker runs in the same binary as the gRPC server. Started from `cmd/server/main.go` via `worker.New(...)` + `w.Start()` on its own goroutine alongside the existing gRPC and Gin servers. Can be split into `cmd/worker/` later if scaling pressure demands it.
- **Workflow versioning:** Use Temporal's `workflow.GetVersion()` patching API for breaking changes to workflow logic. Workflow definitions remain in the `workflows` table; the engine looks up the latest definition at start, but in-flight runs replay against the version they were started with via Temporal's event history. No JSONB snapshot.
- **AI bootstrap:** Phase 6 includes a minimal `internal/ai/` so `RunAISkillActivity` works end-to-end. Scope for the minimal cut: a single `AIServiceServer` exposing `ClassifyText`, `SummarizeText`, and `GenerateReply`, each backed by a single LLM provider call (Anthropic Claude via the existing API key infra). Vector store, RAG, and virtual agents stay deferred to a later AI-specific phase.

---

## Technology Additions Required

| Technology | Purpose | When |
|------------|---------|------|
| NATS JetStream | Event bus, realtime backbone | Phase 0 |
| Temporal | Durable workflow execution, retries, timers, scheduling | Phase 6 |
| OpenSearch | Full-text object + conversation search | Phase 2 |
| ClickHouse | Analytics, call/ticket metrics | Phase 7 |
| gRPC-Web Gateway (Envoy) | Already exists — validate config | Phase 1 |
| Redis Streams | Realtime inbox typing/presence | Phase 3 |
| S3 (already: AWS) | File attachments, call recordings | Phase 4 |
| WebRTC (mediasoup/livekit) | Browser-based VoIP softphone | Phase 5 |

---

## Proto Compilation Strategy

Currently all proto files are compiled into `pkg/pb/` as one flat package.

**Target:** Domain-namespaced Go packages.

```bash
# Before
import "kyla-be/pkg/pb"

# After
import "kyla-be/pb/crm"
import "kyla-be/pb/ticketing"
import "kyla-be/pb/identity"
```

To achieve this without a big-bang rewrite:

1. New proto files go in `proto/{domain}/` directories
2. They compile to `pb/{domain}/` packages
3. Old `pkg/pb/` kept as-is until each domain is migrated
4. Gradual migration: each Phase migrates one domain's protos

---

## Risk Register

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Object Core migration breaks existing data models | High | High | Keep parallel tables; use views/projections; migrate gradually |
| NATS operational complexity | Medium | Medium | Start embedded NATS in dev; managed NATS Cloud in production |
| Proto migration causing import churn | Medium | Medium | Domain namespace from new protos only; keep old pkg/pb intact |
| Feature scope creep delaying core | High | High | Strictly phase-gate; do not start Phase 2 until Phase 1 RBAC is hardened |
| Team size bottleneck | Medium | High | Domain packages map to team ownership; parallelizable after Phase 1 |

---

## Definition of Done (per Phase)

A phase is complete when:
- [ ] All gRPC endpoints for the domain are implemented and returning correct responses
- [ ] Database migrations are in `pkg/db/migrations/` and run clean
- [ ] Domain events are published to NATS on all mutations
- [ ] At least one automated integration test per service covers the happy path
- [ ] `go vet`, `go build` pass with zero errors
- [ ] React frontend has a working UI for the domain (can be basic)
- [ ] Postman collection or buf Studio verified gRPC endpoints

---

*Last updated: 29 May 2026 — Phase 6 complete (Automation + AI + Campaigns). Phase 5 slices 5a/5b/5c shipped (FreeSWITCH foundation + IVR engine + FS config skeleton + frontend SIP.js softphone). Phase 7 (Analytics/Billing) still pending.*
