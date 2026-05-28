# Frontend Overhaul Plan — Kyla Agent Workbench

> **Status:** Plan v1 · Awaiting approval before F0 implementation begins.
> **Author:** Discovery + planning conducted 2026-05-26.
> **Companion docs:** `ROADMAP.md` (backend phases), `PHASE4_BACKEND_REVIEW.md` (backend tech debt to plan around).

---

## Part 1 — Vision

### What we are building

**Kyla Agent Workbench** — a high-end, AI-native operations console for customer-experience teams. It is the *daily driver* for an agent, supervisor, or admin. Optimize for an 8-hour shift, not a demo.

### Mental model

Not "CRM with pages." A **multi-pane workbench** in the lineage of:

| Reference | What we steal |
|---|---|
| **Linear** | Keyboard-first ergonomics; every action reachable in ≤2 keystrokes; dense rows; saved views; ⌘K as connective tissue |
| **Intercom Fin** | AI-in-the-flow: copilot as a persistent surface, not a popup |
| **Cursor** | Right-rail agent panel that is context-aware to the current object |
| **Stripe Dashboard** | Information density done with breathing room; restrained accent use |
| **Vercel / Resend** | Modern dark-first aesthetic, monochrome chrome, single accent |
| **Front / Plain** | Omnichannel inbox feel — the inbox is the agent's homepage |

### Positioning vs the current frontend

The existing `kylaFE/` is a *visual prototype* on top of a production backend:
- ~30 well-built shadcn primitives, a polished data-table, a working shell
- **Zero real backend integration** — all 9 routes use hardcoded data
- 40+ gRPC service clients wired in `GlobalClients.ts` but never called from any page
- Demo-token auth bypass, vestigial `"recoil-persist"` localStorage key, hardcoded localhost endpoints
- ~3 of 15+ backend domains have any UI footprint at all

The overhaul is **not** a refactor — it's **building the real app on top of keepable design primitives**.

---

## Part 2 — Locked Decisions Log

| # | Decision | Choice | Notes |
|---|---|---|---|
| 1 | Migration approach | **In-place evolution** | Refactor `kylaFE/` incrementally; quarantine legacy in `_legacy/` as features replace it |
| 2 | Aesthetic & density | **Linear-dense** | 32px row baseline, hairline borders, monochrome chrome |
| 3 | Brand accent | **Emerald** via shadcn + Tailwind theming | Linear/Vercel/Cursor neighborhood feel |
| 4 | First PR | **Full F0** | Tokens + shell + ⌘K + auth + Query/Zustand + i18n + Storybook + Vitest + Playwright |
| 5 | Auth storage | **Investigated → Hardened localStorage + auto-refresh** | Backend has refresh tokens but no cookies/CORS-credentials; gap-fill deferred to a later hardening phase |
| 6 | AI scope in F0 | **Rail scaffold only** | UI primitives + Zustand store + `⌘J` toggle; real AIService wired in F1 |
| 7 | Typography | **Geist Sans + Geist Mono** | OFL, variable, designed for product UI |
| 8 | Device scope | **Desktop-first responsive** | Optimized for ≥1280px; graceful collapse to ≥768px; <768px shows a "use mobile app" banner |
| 9 | Icon library | **Tabler primary + Lucide selectively** | Keep current Tabler usage; add Lucide where shadcn primitives expect it |
| 10 | Testing | **Vitest + Storybook + Playwright** | All three in F0 |
| 11 | i18n | **Full + RTL + multi-locale in F0** | English default; **proposed locales: en, ar (RTL), fr, sw** — confirm during F0 kickoff |
| 12 | Next deliverable | **This plan document** | Approve before code |

---

## Part 3 — Backend Reality Check

### Phase status (verified against `main` 2026-05-26)
- **Phases 0–4: Complete** — auth, orgs, workspaces, Object Core, omnichannel inbox, CRM, ticketing, KB, forms, projects (stub)
- **Phase 6 core: Complete** — Temporal automation engine, 11 action types, minimal AI (OpenAI/Anthropic/Noop)
- **Phase 5 (Telephony): Not started** — proto layer extensive, servers empty
- **Phase 7 (Analytics/Billing): Not started**
- **Campaigns (within Phase 6): Not started**

### Backend issues frontend must plan around

| Issue | Impact on FE | FE workaround |
|---|---|---|
| Pagination cursor bug (ticketing/KB/forms): `id` cursor used while ordering by `created_at DESC` | Lists may show out-of-order rows; pagination may skip/duplicate items | Use small page sizes (25); flag visible duplicates; coordinate with BE fix |
| `ScopeCheck` defined but not enforced on many Phase-4 endpoints | FE could call across scopes and get data it shouldn't | Pass scope IDs deliberately in every call; don't rely on backend filtering |
| NATS events only emitted from CRM (ticketing/KB/forms/projects silent) | No realtime updates on ticket/KB mutations | Use polling fallback in F3; coordinate with BE for emission |
| `ProjectService` not registered in `main.go` | `/projects` route would 404 | Defer Projects UI until BE wires the service |
| `0006_phase4.sql` not run by `db.InitDB()` | Local dev may lack Phase 4 tables | Document setup; surface "table missing" errors clearly |

### Auth backend findings (from investigation 2026-05-26)

**Tokens:**
- `LoginResponse` returns `access_token` + `refresh_token` (both JWT, HS256, **10-day TTL each** — note: long, worth shortening later)
- `RefreshToken` RPC exists at `kylaBE/pkg/service/auth_server.go:137`
- Frontend currently reads token from `localStorage["recoil-persist"]` (vestigial; Recoil not installed)

**Cookies & CORS:**
- Backend never sets cookies (zero `SetCookie` calls anywhere)
- Envoy CORS lacks `allow_credentials: true` and does not expose `set-cookie` in `expose_headers`
- gRPC-web auth flows through `authorization` metadata header

**WebAuthn:**
- Backend fully implemented (`RegisterPasskey`, `LoginWithPasskey` in `auth_server.go:599-801`, sessions in Redis with 5-min TTL)
- Frontend has stub utilities (`kylaFE/src/api/service/utils/webAuthn.ts`) — not wired

**MFA:**
- TOTP via `pquerna/otp` (30s, 6 digits, SHA1)
- Multi-step flow: `Login()` returns tokens immediately, OR `LoginWithMFA()` accepts code + user_id
- `MFASetup()` provides QR + secret + recovery codes
- `VerifyMFA()` updates `LastMFALogin` without re-issuing tokens

**Verdict:** **Hardened localStorage + auto-refresh interceptor** for F0. HttpOnly-cookie migration is a backend project (set-cookie support + Envoy CORS update); track as separate hardening task post-F1.

---

## Part 4 — Design Language

### 4.1 Color system

Two-tier token architecture so re-theming is a palette swap, not a refactor.

#### Palette layer (OKLch via Tailwind defaults)

```
Brand          Emerald (shadcn/Tailwind)
  emerald.50   #ECFDF5
  emerald.100  #D1FAE5
  emerald.300  #6EE7B7
  emerald.500  #10B981   ← primary accent
  emerald.600  #059669
  emerald.700  #047857
  emerald.900  #064E3B
  emerald.950  #022C22

Neutral        Zinc (cool, Linear-adjacent)
  zinc.50      #FAFAFA
  zinc.100     #F4F4F5
  zinc.200     #E4E4E7
  zinc.300     #D4D4D8
  zinc.400     #A1A1AA
  zinc.500     #71717A
  zinc.600     #52525B
  zinc.700     #3F3F46
  zinc.800     #27272A
  zinc.900     #18181B
  zinc.950     #09090B

Status
  emerald.500  success
  amber.500    warning      #F59E0B
  rose.500     danger       #F43F5E
  sky.500      info         #0EA5E9

Channels (badges only — no chrome bleed)
  emerald.500  WhatsApp
  sky.500      Email
  violet.500   SMS
  amber.500    Voice
  rose.500     WebChat
  fuchsia.500  Instagram
  blue.500     Messenger
```

#### Semantic layer (the only tokens components read)

```
bg.canvas         page background
bg.surface        cards, panels, surfaces raised once
bg.elevated       popovers, dialogs, dropdowns
bg.subtle         hover row, secondary fill
bg.muted          disabled, placeholder

border.default    hairline borders (1px)
border.strong     inputs, selected
border.focus      focus ring (2px outline-offset)

text.primary      headings, body
text.secondary    metadata, labels
text.muted        placeholder, disabled
text.inverse      on accent.solid

accent.solid      emerald.600 (light) / emerald.500 (dark)
accent.subtle     emerald.50 (light) / emerald.950 (dark)
accent.fg         text on accent.solid (zinc.50)

status.success.{solid,subtle,fg}
status.warn.{solid,subtle,fg}
status.danger.{solid,subtle,fg}
status.info.{solid,subtle,fg}

focus.ring        accent.solid @ 35% alpha
selection         accent.solid @ 25% alpha
```

#### Light/dark mapping (excerpt)

```css
:root {
  --bg-canvas: var(--zinc-50);
  --bg-surface: #FFFFFF;
  --bg-subtle: var(--zinc-100);
  --border-default: var(--zinc-200);
  --text-primary: var(--zinc-900);
  --text-secondary: var(--zinc-600);
  --accent-solid: var(--emerald-600);
}

.dark {
  --bg-canvas: var(--zinc-950);
  --bg-surface: var(--zinc-900);
  --bg-subtle: var(--zinc-800);
  --border-default: var(--zinc-800);
  --text-primary: var(--zinc-50);
  --text-secondary: var(--zinc-400);
  --accent-solid: var(--emerald-500);
}
```

### 4.2 Typography

**Geist Sans + Geist Mono** (variable, OFL, self-hosted via `geist` npm package or local woff2).

```
Font families
  --font-sans: "Geist", "Inter", system-ui, sans-serif
  --font-mono: "Geist Mono", "JetBrains Mono", ui-monospace, monospace

Scale (4px-based, line-height optimized)
  text.xs    11px / 16  — micro labels, badges
  text.sm    12px / 18  — table cells, secondary text
  text.base  13px / 20  — body default (NOTE: dense; not 14/16)
  text.md    14px / 20  — buttons, inputs
  text.lg    16px / 24  — section titles
  text.xl    18px / 26  — page titles
  text.2xl   22px / 30  — major headings
  text.3xl   28px / 36  — empty states / hero

Weights
  400 regular
  500 medium      — UI default
  600 semibold    — emphasis, headings
  700 bold        — rare, hero only

Tracking
  Default -0.005em (slightly tight)
  Mono    0em
  Headings -0.015em
```

### 4.3 Spacing & radius

```
Base: 4px grid
  space.0  0
  space.1  4
  space.2  8
  space.3  12
  space.4  16
  space.5  20
  space.6  24
  space.8  32
  space.10 40
  space.12 48
  space.16 64

Radius (tight; Linear-adjacent)
  radius.xs   2   — kbd, micro-badges
  radius.sm   4   — inputs, buttons, tags
  radius.md   6   — cards, panels
  radius.lg   8   — dialogs, drawers, popovers
  radius.xl   12  — toast, large containers

Row heights (density)
  row.compact 28  — data tables max density
  row.default 32  — default inbox/list rows
  row.cozy    40  — settings, forms
```

### 4.4 Elevation

Borders over shadows. Use shadows only for floating surfaces (popovers, dropdowns, dialogs).

```
elevation.0   border 1px solid border.default
elevation.1   0 1px 2px rgb(0 0 0 / 0.04)         — subtle card lift
elevation.2   0 4px 12px rgb(0 0 0 / 0.08)        — popovers
elevation.3   0 16px 32px rgb(0 0 0 / 0.16)       — dialogs, full-screen panels
elevation.4   0 24px 64px rgb(0 0 0 / 0.24)       — command palette (over backdrop)
```

### 4.5 Motion

- Default duration: **150ms** for color, **200ms** for transform, **300ms** for layout
- Easing: `cubic-bezier(0.2, 0, 0, 1)` (Material standard) for most; spring physics for drawers and command palette
- Framer Motion used for orchestrated transitions only — never for hover states (CSS owns those)
- All animations respect `prefers-reduced-motion: reduce`

### 4.6 Iconography

- **Primary:** Tabler Icons React (current dominant usage — keep)
- **Secondary:** Lucide React (where shadcn primitives expect it)
- **Channel & brand icons:** custom SVG components under `src/components/icons/channels/` (WhatsApp, Email, SMS, WebChat, Voice, Instagram, Messenger)
- Default size 16px in dense rows, 18px in buttons, 20px in nav. Stroke 1.5px.

### 4.7 Density and pixel-perfect specs

Inbox row reference (Linear-dense):

```
┌────────────────────────────────────────────────────────────────────┐
│ ● Maria Khalil     WA  2m  #1284  SLA 28m  "Order not arrived..."  │
│ ● Tom Wilson       ✉   4m  #1283  ─        "Cancel order"          │
│ ○ Aisha Bello      SMS 7m  #1282  ─        "Refund question"       │
│ ○ Jordan Park      WA  12m #1281  SLA 1h   "Reschedule delivery"   │
└────────────────────────────────────────────────────────────────────┘
  32px row · 12px row gutter · 1px hairline · 13px text · 11px badges
```

---

## Part 5 — Information Architecture

### 5.1 App shell

```
┌─────────────────────────────────────────────────────────────────────┐
│ [W ▾] [⌘K]  Search anything                  [✨] [🔔]  [● Available ▾]│  ← Top bar (40px)
├──────────┬──────────────────────────────────────────┬───────────────┤
│ Sidebar  │  Primary surface                          │  AI rail      │
│ (208px)  │                                           │  (320px,      │
│          │  Multi-pane:                              │   toggle ⌘J)  │
│ ▸ Inbox 12│   list (320px) + detail (flex)            │              │
│ ▸ CRM    │                                           │  Contextual   │
│ ▸ Tickets│                                           │  to current   │
│ ▸ KB     │                                           │  surface.     │
│ ▸ Forms  │                                           │              │
│ ▸ Auto   │                                           │              │
│ ▸ Calls  │                                           │              │
│ ▸ Camps  │                                           │              │
│ ▸ Stats  │                                           │              │
│ ────     │                                           │              │
│ ▸ Admin  │                                           │              │
├──────────┴──────────────────────────────────────────┴───────────────┤
│ [● Available ▾]   SLA risk: 2   Active: 5   Queue: 14   v1.0.0      │  ← Status bar (28px)
└─────────────────────────────────────────────────────────────────────┘
```

**Responsive collapse:**
- `≥1440px` — full 4-pane (sidebar + list + detail + AI rail)
- `≥1024px` — 3-pane (AI rail becomes a drawer)
- `≥768px` — 2-pane (sidebar collapses to 48px icon rail)
- `<768px` — minimum-viable "use mobile app" banner; only `/login`, `/signup`, `/otp` remain usable

### 5.2 Navigation surfaces

| Surface | Role |
|---|---|
| **Sidebar** | Primary domain nav with live counts (Inbox 12, Tickets 4, etc.) |
| **Top bar** | Workspace switcher (left), ⌘K palette (center), AI toggle / notifications / profile (right) |
| **Right AI rail** | Context-aware copilot; collapsed via ⌘J |
| **Status bar** | Agent presence (AgentOps), SLA breach counter, active items, queue depth, build version |
| **Command palette (⌘K)** | The connective tissue — navigate, run any action, ask AI, switch workspace, jump to any object |
| **Quick switcher (⌘P)** | Object-search variant of ⌘K (just objects, no actions) |
| **Help & shortcuts (?)** | Full keyboard map; searchable |

### 5.3 Multi-tab work (stretch, F1+)

Like an IDE — multiple conversations / deals open as tabs above the primary surface. Persisted across sessions per agent. Defers to F1.

### 5.4 Keyboard shortcuts (initial map)

```
Global
  ⌘K           Command palette
  ⌘P           Quick object switcher
  ⌘J           Toggle AI rail
  ⌘\           Toggle sidebar
  ⌘/           Toggle theme
  ⌘,           Settings
  ⌘B           Back / close detail
  g i          Go to Inbox
  g c          Go to CRM
  g t          Go to Tickets
  g k          Go to Knowledge
  g a          Go to Automation
  ?            Shortcuts overlay

Inbox (F1)
  j / k        Next / prev conversation
  e            Resolve
  s            Snooze
  a            Assign to…
  l            Label
  r            Reply
  n            Internal note
  ⌘Enter       Send
```

---

## Part 6 — Tech Stack

### 6.1 Keep

| Tool | Version | Notes |
|---|---|---|
| React | 19.2 | |
| Vite | 7.2 | |
| TypeScript | 5.9 | Enable `strict: true` if not already |
| React Router | 7.13 | |
| Tailwind CSS | 4.1 | v4 native CSS engine |
| shadcn/ui primitives | current | Refactored to read semantic tokens only |
| TanStack Table | 8.21 | |
| dnd-kit | 6.3 | |
| Sonner | 2.0 | Toasts |
| Recharts | 2.15 | Defer to dashboard phase; consider visx later |
| next-themes | 0.4 | Light/dark switch |
| Tabler Icons React | 3.36 | Primary icons |
| date-fns | 4.1 | |
| Zod | 4 | Used in earnest now (forms, API contracts) |
| @protobuf-ts/grpcweb-transport | current | gRPC-web transport |

### 6.2 Add

| Tool | Version | Purpose |
|---|---|---|
| **@tanstack/react-query** | ^5 | Data fetching, cache, mutations, optimistic updates |
| **@tanstack/react-query-devtools** | ^5 | Dev cache inspector |
| **zustand** | ^5 | Global UI state (auth, workspace, palette, AI rail) |
| **react-hook-form** | ^7 | Form management |
| **@hookform/resolvers** | ^3 | Zod integration |
| **cmdk** | ^1 | ⌘K palette |
| **framer-motion** | ^11 | Orchestrated transitions |
| **react-hotkeys-hook** | ^4 | Layered keyboard shortcuts |
| **@tanstack/react-virtual** | ^3 | Virtualized inbox/list |
| **lucide-react** | ^0.4 | Secondary icons (shadcn-default) |
| **geist** | ^1 | Self-hosted Geist Sans + Mono |
| **react-i18next + i18next** | latest | i18n + RTL |
| **i18next-browser-languagedetector** | latest | Locale detection |
| **vitest** | ^2 | Unit tests |
| **@testing-library/react** | ^16 | Component tests |
| **@testing-library/jest-dom** | ^6 | DOM matchers |
| **@storybook/react-vite** | ^8 | Component docs |
| **@playwright/test** | ^1 | E2E |
| **@sentry/react** | ^8 | Error reporting (configured but not wired until deployment) |

### 6.3 Remove

| Tool | Why |
|---|---|
| `little-date` | date-fns already covers it |
| References to `"recoil-persist"` localStorage key | Vestigial; replace with proper auth store |
| Hardcoded `localhost:8082`, `localhost:8083` in messaging service URLs | Move to env vars |

### 6.4 Target file structure (post-F0)

```
kylaFE/
├── src/
│   ├── app/
│   │   ├── shell/                    NEW
│   │   │   ├── AppShell.tsx
│   │   │   ├── TopBar.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   ├── AIRail.tsx
│   │   │   ├── StatusBar.tsx
│   │   │   └── CommandPalette.tsx
│   │   ├── routes/                   refactored
│   │   │   ├── _authenticated.tsx    (layout: requires auth + shell)
│   │   │   ├── _public.tsx           (layout: login/signup, no shell)
│   │   │   ├── login.tsx
│   │   │   ├── signup.tsx
│   │   │   ├── otp.tsx
│   │   │   └── index.tsx             (redirect → /inbox in F1)
│   │   └── providers/                NEW
│   │       ├── QueryProvider.tsx
│   │       ├── ThemeProvider.tsx
│   │       ├── I18nProvider.tsx
│   │       └── AuthProvider.tsx
│   ├── design-system/                NEW (replaces components/ui/ over time)
│   │   ├── tokens/
│   │   │   ├── palette.css
│   │   │   ├── semantic.css
│   │   │   └── tokens.ts             (typed accessors)
│   │   ├── primitives/               (refactored shadcn)
│   │   │   ├── Button.tsx
│   │   │   ├── Input.tsx
│   │   │   ├── Card.tsx
│   │   │   ├── Badge.tsx
│   │   │   ├── Kbd.tsx               NEW
│   │   │   └── ...
│   │   ├── patterns/                 NEW (composed primitives)
│   │   │   ├── ListRow.tsx
│   │   │   ├── EmptyState.tsx
│   │   │   ├── ErrorState.tsx
│   │   │   ├── LoadingSkeleton.tsx
│   │   │   ├── PageHeader.tsx
│   │   │   ├── SidePanel.tsx
│   │   │   └── DataTable.tsx         (refactored from current monolith)
│   │   ├── icons/                    NEW
│   │   │   ├── channels/
│   │   │   ├── brand/
│   │   │   └── index.ts
│   │   └── index.ts                  (single import surface)
│   ├── features/                     NEW (per-domain feature modules, populated F1+)
│   │   ├── inbox/                    F1
│   │   ├── crm/                      F2
│   │   ├── service-desk/             F3
│   │   ├── automation/               F4
│   │   ├── admin/                    F5
│   │   ├── analytics/                F6
│   │   ├── telephony/                F7
│   │   └── campaigns/                F8
│   ├── lib/
│   │   ├── rpc/                      NEW
│   │   │   ├── client.ts             (gRPC clients with interceptors)
│   │   │   ├── interceptors.ts       (auth metadata, error toast, Sentry)
│   │   │   ├── stream.ts             (server-streaming helpers)
│   │   │   └── errors.ts             (gRPC error → typed app error)
│   │   ├── query/                    NEW
│   │   │   ├── client.ts             (QueryClient)
│   │   │   ├── keys.ts               (query-key factory)
│   │   │   └── mutations.ts          (shared mutation helpers)
│   │   ├── auth/                     NEW
│   │   │   ├── store.ts              (zustand auth slice)
│   │   │   ├── api.ts                (login, refresh, logout, mfa, passkey)
│   │   │   ├── interceptor.ts        (401 → refresh → retry)
│   │   │   └── guards.ts             (route guards)
│   │   ├── i18n/                     NEW
│   │   │   ├── config.ts
│   │   │   └── direction.ts          (RTL switching)
│   │   ├── workspace/                NEW
│   │   │   └── store.ts              (current workspace + switcher)
│   │   ├── shortcuts/                NEW
│   │   │   └── registry.ts           (scoped hotkey definitions)
│   │   ├── command/                  NEW
│   │   │   └── registry.ts           (⌘K command registry)
│   │   └── ai/                       NEW (scaffold only)
│   │       ├── rail-store.ts
│   │       └── streaming.ts          (placeholder)
│   ├── locales/                      NEW
│   │   ├── en.json
│   │   ├── ar.json                   (RTL)
│   │   ├── fr.json
│   │   └── sw.json
│   ├── pb/                           generated (do not edit)
│   └── _legacy/                      NEW — quarantined old code, deleted as v2 absorbs
├── .storybook/                       NEW
├── tests/
│   ├── unit/                         (Vitest)
│   └── e2e/                          (Playwright)
├── playwright.config.ts              NEW
├── vitest.config.ts                  NEW
└── package.json
```

---

## Part 7 — Phased Roadmap (F0 → F8)

| Phase | Goal | Weeks | Depends on |
|---|---|---|---|
| **F0** | Foundation: tokens, shell, ⌘K, auth, Query, i18n, AI rail scaffold, Storybook, tests | 2–3 | — |
| **F1** | Inbox + Conversations + first real AI copilot | 3–4 | F0 |
| **F2** | CRM + Object Core (lists/detail/timeline/pipeline kanban) | 3 | F0 |
| **F3** | Service Desk (Tickets + KB + Forms) | 2–3 | F0 |
| **F4** | Automation Studio (visual workflow builder on React Flow) + AI Studio | 3 | F0 |
| **F5** | Admin & Settings (org/branches/depts/teams/RBAC/apps/audit) | 1–2 | F0 |
| **F6** | Analytics & Dashboards (BE Phase 7 dependent) | 2–3 | BE Phase 7 |
| **F7** | Telephony / Softphone (BE Phase 5 dependent) | 3 | BE Phase 5 |
| **F8** | Campaigns (BE Phase 6 Campaigns dependent) | 2 | BE Campaigns |

Each phase ships dogfood-able functionality. F1–F4 are parallelizable in pairs once F0 patterns are proven.

---

## Part 8 — F0 Detailed Implementation Plan

### 8.1 Scope statement

F0 lands the foundation that everything else rides on. **No domain functionality.** Exit state: an agent can sign in (real auth), see the shell, navigate via sidebar + ⌘K, toggle the AI rail (empty state), switch language including RTL, and see the design system documented in Storybook — but no real Inbox/CRM/etc surfaces exist yet.

### 8.2 Module-by-module plan

#### M1 — Design system tokens

**Files:**
- `src/design-system/tokens/palette.css` — raw palette as CSS custom properties
- `src/design-system/tokens/semantic.css` — semantic mappings with `:root` and `.dark`
- `src/design-system/tokens/tokens.ts` — typed accessors (`token('bg.canvas')` etc.)
- `src/index.css` — imports the above, removes old hardcoded color block

**Acceptance:**
- Every existing `bg-*`, `text-*`, `border-*` Tailwind class in components either resolves through a semantic token OR is migrated to use one
- Storybook page "Tokens" renders both light and dark swatches for every semantic token
- Theme toggle flips all surfaces correctly with no FOUC

#### M2 — Typography

**Files:**
- `src/design-system/tokens/typography.css` — font-face, scale, weights
- Self-host Geist via `geist` npm package
- Tailwind config maps `text-*` utilities to the new scale

**Acceptance:**
- Geist Sans loads with `font-display: swap`, no FOIT >100ms
- Geist Mono used in all monospaced surfaces (IDs, kbd, code)
- RTL locales fall back to system Arabic font (`"SF Arabic", "Geeza Pro", sans-serif`)

#### M3 — Refactored primitives

Migrate `src/components/ui/*` → `src/design-system/primitives/*`, each component reading only semantic tokens.

**In scope (F0):**
- Button (variants: primary, secondary, ghost, danger, link · sizes: sm, md, lg)
- Input, Textarea, Select (with proper focus rings)
- Checkbox, Switch, Radio
- Card, Separator, Badge, Kbd (NEW), Avatar
- Dialog, Drawer, Sheet, Popover, Tooltip, DropdownMenu
- Tabs, Breadcrumb, Skeleton
- Toast (Sonner wrapper with semantic colors)

**Deferred to F1+:** DataTable (refactor from 796-line monolith), Calendar specialization, Chart wrappers.

**Acceptance:**
- Every primitive has a Storybook story (default, all variants, all states)
- Every primitive has at least one Vitest test
- Focus states visible and consistent across all interactive primitives

#### M4 — App shell

**Files:**
- `src/app/shell/AppShell.tsx` — 4-pane layout container, responsive
- `src/app/shell/TopBar.tsx`
- `src/app/shell/Sidebar.tsx` — collapsible to 48px rail
- `src/app/shell/AIRail.tsx` — empty state in F0, controlled by Zustand
- `src/app/shell/StatusBar.tsx` — placeholder counts in F0

**Acceptance:**
- Sidebar items defined in a typed registry; counts can be wired in later phases
- ⌘\ toggles sidebar collapse
- ⌘J toggles AI rail
- Responsive collapse rules verified at 1440 / 1024 / 768 / 375 widths

#### M5 — Command palette (⌘K)

**Files:**
- `src/app/shell/CommandPalette.tsx` — cmdk-based
- `src/lib/command/registry.ts` — command definitions with scopes (`global`, `inbox`, `crm`, …)
- `src/lib/command/use-commands.ts` — hook for per-route command registration

**F0 command set:**
- Navigation: "Go to Inbox", "Go to CRM", "Go to Settings", etc.
- Theme: "Toggle theme", "Set theme to light/dark/system"
- Locale: "Change language" → sub-list of locales
- Workspace: "Switch workspace" → sub-list
- Account: "Sign out"
- Help: "Open keyboard shortcuts"

**Acceptance:**
- Opens with ⌘K, closes with Escape
- Fuzzy search across all registered commands
- Keyboard nav (↑↓ + Enter) feels Linear-fast
- Recently-used commands surface to top

#### M6 — Auth

Based on backend investigation: **hardened localStorage + auto-refresh**.

**Files:**
- `src/lib/auth/store.ts` — Zustand slice: `{user, accessToken, refreshToken, status: 'idle' | 'authenticating' | 'authenticated' | 'mfa_required'}`
- `src/lib/auth/api.ts` — wrappers for `Login`, `LoginWithMFA`, `RegisterPasskey`, `LoginWithPasskey`, `RefreshToken`, `Logout`, `SignUp`, `ActivateUserAccount`
- `src/lib/auth/interceptor.ts` — gRPC interceptor that catches `Unauthenticated` errors, calls `RefreshToken`, retries the original call once
- `src/lib/auth/guards.ts` — `<RequireAuth>` route wrapper
- `src/app/routes/login.tsx` — wired to real `Login` RPC
- `src/app/routes/signup.tsx` — wired to real `SignUp` RPC
- `src/app/routes/otp.tsx` — wired to MFA + ActivateUserAccount paths

**Storage strategy:**
- Tokens in `localStorage` under a new namespaced key (e.g., `kyla.auth.v1`)
- Tokens NEVER read directly from components — only via `useAuthStore()`
- On 401: try refresh; on refresh failure, clear store, redirect to `/login`
- Background refresh at the 80% mark of token TTL (8 days for current 10d TTL)

**Stretch (F0 if time, else F1):**
- Wire passkey login (backend ready; frontend stub exists; just needs final integration)

**Acceptance:**
- Real login → see shell, real user data
- Reload preserves session
- Token expiry triggers refresh transparently
- Logout clears all auth state and Query cache
- "demo-token" bypass deleted
- `"recoil-persist"` localStorage key reads deleted

#### M7 — Query + state layer

**Files:**
- `src/lib/query/client.ts` — QueryClient with sensible defaults (`staleTime: 30s`, `gcTime: 5m`, `retry: 1`, error → toast)
- `src/lib/query/keys.ts` — typed query-key factory: `qk.conversations.list(workspaceId)`, `qk.conversations.detail(id)`, etc.
- `src/lib/query/mutations.ts` — shared mutation helpers (optimistic update pattern)
- `src/app/providers/QueryProvider.tsx` — provider + devtools (dev only)

**Acceptance:**
- React Query Devtools available in dev
- Standard pattern documented in Storybook MDX: "How to fetch", "How to mutate", "How to invalidate"
- Errors route to toast + Sentry (Sentry stubbed for now)

#### M8 — i18n + RTL

**Files:**
- `src/lib/i18n/config.ts` — i18next setup with browser-languagedetector
- `src/lib/i18n/direction.ts` — manages `<html dir="rtl|ltr">` based on locale
- `src/app/providers/I18nProvider.tsx`
- `src/locales/en.json` — base
- `src/locales/ar.json` — RTL test locale (Arabic)
- `src/locales/fr.json` — French
- `src/locales/sw.json` — Swahili

**Proposed locales** (confirm at F0 kickoff): English (default), Arabic (RTL), French, Swahili. Aligned with the platform's apparent African market focus (BistrOafrica, Africa's Talking SMS integration, WhatsApp-first).

**Acceptance:**
- Every UI string in F0 surfaces is in `en.json` — no inline English literals
- Language switcher in profile menu changes locale + direction live
- RTL mirrors layout correctly (sidebar swaps to right, ⌘K still works, focus rings)
- Linter rule (or pre-commit) flags new untranslated strings
- Date/number formatting uses `Intl` with current locale

#### M9 — AI rail scaffold

**Files:**
- `src/app/shell/AIRail.tsx`
- `src/lib/ai/rail-store.ts` — Zustand: `{ isOpen, currentContext, history }`
- `src/lib/ai/streaming.ts` — placeholder for SSE / gRPC server-streaming reader (real wiring in F1)
- `src/design-system/patterns/AISuggestion.tsx` — card primitive for AI output with accept/reject
- `src/design-system/patterns/StreamingText.tsx` — character-by-character reveal (mocked in F0)

**Acceptance:**
- ⌘J toggles rail
- Empty state reads: "Open a conversation, contact, or workflow to talk to Kyla."
- AISuggestion + StreamingText documented in Storybook with mock data
- `useAISession()` hook exists with stubbed responses for F1 to swap in

#### M10 — Workspace context

**Files:**
- `src/lib/workspace/store.ts` — current workspace + member list cache
- `src/app/shell/WorkspaceSwitcher.tsx` — top-bar dropdown
- Loads on auth resolve; persists last selection per user

**Acceptance:**
- Switching workspace invalidates Query cache for workspace-scoped keys
- Direct URL access to a workspace-scoped route validates membership

#### M11 — Storybook

**Files:**
- `.storybook/main.ts`, `.storybook/preview.tsx` — Vite + addon-a11y + addon-themes
- `src/design-system/**/*.stories.tsx` — one per primitive
- MDX docs for tokens, typography, density, motion

**Acceptance:**
- `pnpm storybook` runs and shows all primitives + token reference
- Light/dark toggle in toolbar
- RTL toggle in toolbar (for layout testing)
- a11y addon passing on all stories

#### M12 — Testing infra

**Files:**
- `vitest.config.ts` + jsdom setup
- `tests/unit/lib/*.test.ts` — auth store, query keys, command registry, shortcuts registry
- `playwright.config.ts`
- `tests/e2e/auth.spec.ts` — login → shell renders → sign out
- `tests/e2e/shell.spec.ts` — ⌘K opens, navigation works, theme toggle persists
- `tests/e2e/i18n.spec.ts` — language switch flips direction, key strings translate
- GitHub Actions workflow: lint + typecheck + unit + e2e on every PR

**Acceptance:**
- All three test types runnable locally with one command (`pnpm test`)
- CI passes on `main` on day 1 of F0 merge

#### M13 — Cleanup

Quarantine legacy code into `src/_legacy/` (not deleted yet — those pages still need to demo until F1 ships the inbox replacement):
- Move `src/components/data-table.tsx`, `call-widget.tsx`, demo support pages, etc.
- Add `// eslint-disable-next-line` block import allowing `_legacy/` only from `_legacy/`
- Remove vestigial:
  - `"recoil-persist"` references
  - hardcoded `localhost:8082/8083` (move to env)
  - dead `webAuthn.ts` stubs (will replace properly in F0 stretch or F1)

### 8.3 Deliverable checklist

```
[ ] M1  · Design tokens (palette + semantic)
[ ] M2  · Geist typography
[ ] M3  · Refactored primitives (15+ components)
[ ] M4  · App shell (4-pane, responsive)
[ ] M5  · Command palette ⌘K
[ ] M6  · Real auth (login + signup + MFA + token refresh)
[ ] M7  · TanStack Query + Zustand + interceptors
[ ] M8  · i18n + RTL with 4 locales
[ ] M9  · AI rail scaffold + AISuggestion + StreamingText
[ ] M10 · Workspace context + switcher
[ ] M11 · Storybook with all primitives + token docs
[ ] M12 · Vitest + Playwright + CI workflow
[ ] M13 · Quarantine legacy + cleanup vestigial code
```

### 8.4 Exit criteria

F0 is done when **all** of the following are true:

1. A real user can sign in with email/password (and MFA where enabled) using actual backend `Login` / `LoginWithMFA` RPCs
2. Reload preserves session; expired access tokens are refreshed transparently
3. The shell renders with: top bar, sidebar (with at least one nav target per future phase), AI rail (empty state), status bar
4. ⌘K opens the command palette with ≥15 working commands
5. Theme toggle (light/dark) persists and applies to all surfaces
6. Language switcher supports English + Arabic (RTL) + French + Swahili (or final agreed set)
7. Every UI string is externalized — no inline English literals in F0 code
8. Storybook documents every refactored primitive + the token reference + density specs
9. CI runs lint + typecheck + Vitest + Playwright on every PR and passes on main
10. No `demo-token`, `"recoil-persist"`, or hardcoded `localhost:808x` references in non-legacy code
11. Lighthouse a11y score ≥95 on `/login` and the empty shell

### 8.5 Risks & mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Backend `RefreshToken` RPC behavior different than spec | Medium | High | Write integration test against local backend before relying on it |
| RTL layout regressions in shadcn primitives (Radix) | Medium | Medium | Storybook RTL toggle catches early; test every primitive |
| Tailwind v4 + Geist + variable font subset mismatch | Low | Low | Use `geist` npm package; subset via `unicode-range` if needed |
| Storybook 8 + Vite 7 compatibility | Low | Medium | Pin versions; test on day 1 |
| Long token TTL (10 days) is a security smell; can't fix in F0 | Certain | Low | Document; raise with backend team for separate hardening |
| Existing `kylaFE` pages break during refactor | High | Medium | Quarantine in `_legacy/`; replace one route at a time |
| Translation strings drift between English and other locales | Medium | Low | Pre-commit script checks key-set equality |

### 8.6 Estimated timeline (2 senior FE engineers)

```
Week 1   M1, M2, M3 (tokens + typography + primitives)
         M11 (Storybook scaffold in parallel)
Week 2   M4, M5 (shell + command palette)
         M6 starts (auth investigation already done; wiring begins)
Week 3   M6 (auth) finishes
         M7, M8, M9, M10 (query, i18n, AI scaffold, workspace)
         M12 (tests fill in)
         M13 (cleanup)
         Storybook docs + acceptance walk-through
```

Single PR or feature-flagged multi-PR per team preference. Recommendation: **multi-PR behind a `?v2=1` query flag** so reviewers can see incremental progress.

### 8.7 Open questions for F0 kickoff

These don't block plan approval but need answers before code lands:

1. **Locale set confirmation** — proposed: en, ar, fr, sw. Confirm or amend.
2. **Sentry DSN** — provision now, or stub and add post-F0?
3. **Backend env URL strategy** — single `VITE_API_URL` or keep the current multi-hostname split? Recommend collapsing to one Envoy gateway URL.
4. **Lighthouse / a11y targets** — confirm ≥95 a11y, ≥90 performance acceptable for F0 (no real data yet, so easy to hit).
5. **Feature-flag mechanism** — query param `?v2=1`, env-driven build flag, or proper feature flagging service?
6. **CI provider** — GitHub Actions assumed; confirm.

---

## Part 9 — Beyond F0 (high-level deliverables per phase)

### F1 — Inbox + Conversations (3–4 weeks)
- Omnichannel list with channel/status/assignee/priority/SLA filters, saved views
- Streaming via `StreamConversationUpdates` (gRPC server-streaming)
- Conversation detail: thread view, Tiptap composer, attachments, channel-aware rendering
- Side panel: contact card, related deals/tickets, object timeline
- **AI copilot wired (first real integration):** summarize thread, suggested replies (streaming), sentiment + intent, translate, draft from KB
- Macros quick-insert
- Assign / transfer / snooze / resolve actions
- SLA countdown badges
- Bulk actions

### F2 — CRM + Object Core (3 weeks)
- Contacts / Companies / Leads / Deals list + detail (Object Core powered, schema-aware)
- Pipeline kanban board with drag-to-stage
- Deal detail: `object_events` timeline, related conversations/tickets, notes
- Saved views (shared filters)
- CSV import/export
- Schema editor for admins
- Reusable ObjectTimeline + ObjectRelations components

### F3 — Service Desk (Tickets + KB + Forms) (2–3 weeks)
- Tickets list/detail with room threads + macro management
- KB article editor (Tiptap) + categories + draft/publish + search
- Public help center theme (embeddable)
- Form builder with conditional logic + embed code generator
- AI: "suggest KB article" while composing ticket replies; auto-categorize incoming tickets

### F4 — Automation Studio + AI Studio (3 weeks)
- React Flow visual workflow builder with 11 action nodes + branching
- Test-run UI with step-by-step execution view (via `TestRunWorkflow`)
- Run history + replay
- AI playground: Classify / Summarize / Generate Reply with token cost estimates
- Templates gallery

### F5 — Admin & Settings (1–2 weeks)
- Org → branches → departments → teams hierarchy editor
- Users, invitations, role assignments
- RBAC matrix editor (Casbin-backed)
- Apps & webhook management
- Audit log viewer
- Workspace settings, object schema management
- AgentOps status switcher in status bar

### F6 — Analytics (BE Phase 7 dependent, 2–3 weeks)
- Dashboard builder (widget grid)
- Pre-built reports: SLA, volume, agent performance, channel mix, deal velocity
- Drill-down
- Scheduled exports

### F7 — Telephony / Softphone (BE Phase 5 dependent, 3 weeks)
- WebRTC softphone replacing the current mockup
- Call list, recording playback, transcript with AI summary
- IVR builder (reuse React Flow canvas from F4)
- Click-to-call from any contact/deal/ticket
- Supervisor wallboard

### F8 — Campaigns (BE Campaigns dependent, 2 weeks)
- Audience segmentation
- Channel + content composer
- WhatsApp approved-template picker
- Campaign analytics + responses fed into inbox
- A/B test scaffolding

---

## Part 10 — Cross-cutting Principles

1. **AI as platform primitive, not a feature** — one `<AISurface>` contract used everywhere. Streaming text, citation chips, accept/reject — shared.
2. **Object Core is the data spine** — every domain UI is a typed wrapper over the same Object Core list/detail/timeline/relations components. Avoid one-offs.
3. **Streaming first** — anything that can stream from BE consumes a stream. No polling unless explicitly degrading.
4. **Saved Views everywhere** — single filter/sort/columns primitive, persisted via BE views layer.
5. **Workspace-scoped everything** — top-bar switcher invalidates Query cache; zero cross-workspace bleed.
6. **One way to do things** — single primitive each for data table, kanban, list-detail, side panel, modal/drawer, form, empty state, error state, loading skeleton.
7. **Keyboard parity** — every mouse action has a keyboard equivalent surfaced in ⌘K and `?` overlay.
8. **Density is a feature** — never default to comfortable when dense reads fast enough.
9. **Accessibility baked in** — WCAG AA, focus rings, ARIA, live regions, prefers-reduced-motion.
10. **Mobile is responsive, not native-first** — kylaMB (Flutter) owns true mobile.

---

## Appendix A — Reference: Backend Data Anchors

| FE module | Backend surface |
|---|---|
| Auth | `AuthService` (Login, LoginWithMFA, RefreshToken, LoginWithPasskey, MFASetup, VerifyMFA, SignUp, ActivateUserAccount) |
| Inbox | `CommunicationService` + `ConversationService` (Create/Get/List/Assign/Update/Resolve/SendMessage/ListMessages/StreamConversationUpdates) |
| CRM | `CRMService` (pipelines + stages) + `ObjectCoreService` (objects with type_slug=deal/contact/company/lead) |
| Tickets | `TicketingService` (rooms + messages + macros) |
| Knowledge | `KnowledgeService` (categories + articles + search + publish) |
| Forms | `FormsService` (CRUD + SubmitForm public + ListSubmissions) |
| Automation | `WorkflowService` (Create/Update/Get/List/Delete/GetRunHistory/TestRunWorkflow) |
| AI (F1+) | `AIService` (ClassifyText, SummarizeText, GenerateReply) |
| Admin/Identity | `OrganisationService`, `BranchService`, `DepartmentService`, `TeamService`, `UserService`, `RBACService`, `InvitationService`, `AppService`, `AuditService` |
| AgentOps | `AgentStatusService` (status changes, history, availability) |

---

## Appendix B — Approval

| Reviewer | Status | Date |
|---|---|---|
| Product owner | ☐ Pending | |
| FE lead | ☐ Pending | |
| Design lead | ☐ Pending | |

Once approved, F0 implementation kicks off with the auth integration test against local backend, followed by Week 1 deliverables (tokens, typography, primitives, Storybook scaffold).
