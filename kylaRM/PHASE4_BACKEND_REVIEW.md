# Phase 4 Backend Implementation Review

## Scope
- Reviewed backend only, per request.
- Reviewed implemented domains under kylaBE/internal: crm, ticketing, knowledge, forms.
- Compared implementation against Phase 4 roadmap in kylaRM/ROADMAP.md.
- Validated compile status with:
  - go test ./internal/crm ./internal/ticketing ./internal/knowledge ./internal/forms
  - go test ./cmd/server

## Executive Summary
Phase 4 backend is partially implemented. Core CRUD foundations exist for CRM pipelines, ticket rooms/macros, KB categories/articles, and forms/submissions, but there are major completeness and readiness gaps.

Decision: Not ready to proceed to Phase 5 yet.

## Roadmap Coverage (Backend)
| Area | Expected in Roadmap | Observed Status |
|---|---|---|
| CRM | Lead/Deal/Company/Activity object-core model, pipeline + views, deal-conversation linking | Partial |
| Ticketing | Ticket object-core + SLA/escalation/macro/rooms/assignment/CSAT trigger | Partial |
| Knowledge | Article CRUD + categories + search (+ AI stretch) | Partial-to-Good |
| Forms | Builder logic + submissions create object-core record + workflow trigger | Partial |
| Projects/Tasks | internal/projects domain and APIs | Missing |

## Findings

### Critical

1. Projects and Tasks backend is not implemented
- Evidence:
  - Roadmap explicitly requires internal/projects in Phase 4: kylaRM/ROADMAP.md:626
  - No projects backend package exists: kylaBE/internal/projects (no files)
  - Project gRPC service exists in proto: kylaPB/project.proto:69
  - Project service is not registered in server startup (only CRM/Ticketing/Knowledge/Forms are): kylaBE/cmd/server/main.go:502
- Impact:
  - Phase 4 scope is incomplete against committed roadmap deliverables.
  - Any frontend/client integration for projects/tasks cannot be completed reliably.
- Recommendation:
  - Add minimal internal/projects domain now with model/store/server and gRPC registration.
  - Deliver at least CRUD + list + archive endpoints for ProjectService, then iterate to task boards.

2. Authorization is coarse; scope-level access checks are not enforced in Phase 4 servers
- Evidence:
  - Each Phase 4 server defines ScopeCheck in AuthGateway but uses only GetServiceAuthMetadata checks:
    - kylaBE/internal/crm/server.go:19
    - kylaBE/internal/ticketing/server.go:18
    - kylaBE/internal/knowledge/server.go:16
    - kylaBE/internal/forms/server.go:16
  - Calls consistently gate on RequestAuth only (403 on failure) without workspace or record-level scope checks, for example: kylaBE/internal/crm/server.go:38
- Impact:
  - Cross-workspace access risks in a multi-tenant hierarchy.
  - Difficult to guarantee least-privilege behavior before telephony and additional channels are added in Phase 5.
- Recommendation:
  - Enforce ScopeCheck on org/workspace/object for all mutating and read endpoints.
  - Centralize authorization helper per domain (authorizeOrg, authorizeWorkspace, authorizeObject).
  - Add audit entries for allow/deny decisions on sensitive mutations.

### High

3. Cursor pagination is incorrect in multiple endpoints
- Evidence:
  - Ticket room messages paginate by id cursor while ordering by created_at desc:
    - kylaBE/internal/ticketing/store.go:77
    - kylaBE/internal/ticketing/store.go:79
  - KB articles use id > page_token while ordering by created_at desc:
    - kylaBE/internal/knowledge/store.go:134
    - kylaBE/internal/knowledge/store.go:136
  - Form submissions use id > page_token while ordering by created_at desc:
    - kylaBE/internal/forms/store.go:110
    - kylaBE/internal/forms/store.go:112
- Impact:
  - Missing, duplicated, or unstable result pages under normal data growth.
  - UI inconsistencies and hard-to-debug list behavior.
- Recommendation:
  - Switch to stable composite cursor strategy: (created_at, id).
  - For DESC ordering use cursor predicate: (created_at, id) < (cursor_created_at, cursor_id).
  - Add pagination correctness tests with seeded out-of-order UUIDs.

4. Domain event publication is incomplete for Phase 4 mutations
- Evidence:
  - Event publication is implemented only in CRM server paths:
    - kylaBE/internal/crm/server.go:58
    - kylaBE/internal/crm/server.go:219
    - kylaBE/internal/crm/server.go:278
  - No corresponding publish flow observed in ticketing, knowledge, and forms servers.
- Impact:
  - Automation/notification/analytics consumers cannot react to most Phase 4 changes.
  - Conflicts with the platform event-driven architecture and Definition of Done intent.
- Recommendation:
  - Publish mutation events for create/update/delete across ticketing, knowledge, forms.
  - Standardize subject naming and payload contracts by entity/action.
  - Add integration test asserting event emission per mutation path.

5. Migration strategy mismatch can leave production DB without intended indexes/constraints
- Evidence:
  - SQL migration file includes explicit Phase 4 schema and FTS indexes: kylaBE/pkg/db/migrations/0006_phase4.sql:1
  - Runtime startup relies on AutoMigrate for Phase 4 tables: kylaBE/cmd/server/main.go:163
  - DB initializer only opens connection and does not run SQL migration files: kylaBE/pkg/db/db.go:15
- Impact:
  - Environment drift between developer DBs and deployed DBs.
  - Performance degradation if FTS/index definitions from SQL are not applied.
- Recommendation:
  - Adopt one migration path as source of truth (preferred: SQL migrations).
  - Add migration runner into startup or release pipeline.
  - Validate required indexes post-migration in CI.

### Medium

6. CRM implementation is narrower than Phase 4 roadmap outcomes
- Evidence:
  - Roadmap expects Lead/Deal/Company/Activity object types and deal-conversation relations: kylaRM/ROADMAP.md:603
  - Current CRM proto and server focus on pipeline/stage/move-board only:
    - kylaPB/crm.proto:159
    - kylaBE/internal/crm/server.go:35
- Impact:
  - Sales workflows remain incomplete and integration with communication history is limited.
- Recommendation:
  - Add relation APIs and object-core seeding for CRM system types.
  - Implement deal-conversation linkage endpoint first as a thin integration slice.

7. Forms submission endpoint is intentionally public but lacks hardening controls
- Evidence:
  - Public submission path with no auth check: kylaBE/internal/forms/server.go:132
- Impact:
  - Increased abuse/spam risk and potential load amplification.
- Recommendation:
  - Add rate limiting, optional CAPTCHA/HMAC form token, payload size limits, and input schema validation.
  - Emit submission event for workflow trigger compatibility.

8. Test coverage is effectively absent for Phase 4 backend domains
- Evidence:
  - go test reports no test files for crm, ticketing, knowledge, forms, and cmd/server.
- Impact:
  - High regression risk while entering Phase 5 telephony complexity.
- Recommendation:
  - Add table-driven unit tests for store logic and auth/error behavior.
  - Add at least one integration happy-path test per Phase 4 service.

## What Is Working Well
- All four implemented Phase 4 services are wired into gRPC startup:
  - kylaBE/cmd/server/main.go:502
- Data models for CRM/Ticketing/Knowledge/Forms are present and compile.
- Basic CRUD behavior for KB and forms exists and is coherent.

## Required Steps Before Phase 5

1. Close Phase 4 scope blocker
- Implement internal/projects backend domain and register ProjectService.

2. Enforce scoped authorization
- Add and validate org/workspace/object scope checks in all Phase 4 handlers.

3. Fix cursor pagination correctness
- Replace id-only cursor logic with stable created_at+id cursor semantics.

4. Complete event emission coverage
- Emit domain events for all mutating operations in ticketing, knowledge, forms, and missing CRM mutations.

5. Unify migration strategy
- Ensure 0006_phase4.sql (and related SQL migrations) is actually executed in non-local environments.

6. Add minimum regression test gate
- Unit tests for each Phase 4 domain store/server.
- Integration happy-path tests for each service.
- CI gate: go test for Phase 4 packages must pass.

## Phase 5 Readiness Gate (Suggested)
Proceed only when all are true:
- ProjectService implemented and registered.
- Scope-based auth checks enforced and tested.
- Pagination cursor bugs fixed and tested.
- Event emission coverage implemented for all Phase 4 mutations.
- Migration execution path verified across environments.
- Minimum test coverage gate established for Phase 4 domains.
