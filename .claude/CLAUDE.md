# CLAUDE.md — kylaCX

> Context file for AI-assisted development. Keep this file up to date as the project evolves.

---

## Project Overview

**kylaCX** is a customer experience platform monorepo. It is composed of five sub-projects:

| Directory | Role | Language / Runtime |
|-----------|------|--------------------|
| `kylaBE/` | Core gRPC + REST backend | Go 1.23 |
| `kylaFE/` | Web frontend | TypeScript · React 19 · Vite |
| `kylaMB/` | Mobile application | Dart / Flutter |
| `kylaEX/` | Extensions / integrations | TBD |
| `kylaRM/` | Remote / edge module | TBD |
| `kylaPB/` | Shared Protobuf definitions | Proto3 |

---

## Repository Structure

```
kylaCX/
├── CLAUDE.md               ← you are here
├── Makefile                ← root build commands (proto generation, etc.)
├── kylaPB/                 ← single source of truth for all .proto files
├── kylaBE/
│   ├── cmd/server/         ← entrypoint (main.go)
│   ├── config/             ← Viper-based config structs + loader
│   ├── internal/           ← domain-scoped packages (one sub-directory per domain)
│   │   ├── agentops/       ← agent status model / store / server
│   │   ├── apps/           ← API apps model / store / server
│   │   ├── auth/           ← auth wrapper store + JWT/DBAuthStore type aliases
│   │   ├── authctx/        ← request metadata types (shared, no domain logic)
│   │   ├── branch/         ← branch model / store / server
│   │   ├── contact/        ← contact + group model / store / servers
│   │   ├── department/     ← department model / store / server
│   │   ├── invitation/     ← invitation model / store / server
│   │   ├── label/          ← label store / server
│   │   ├── leave/          ← leave model / store / server
│   │   ├── middleware/     ← gRPC auth + response + session interceptors
│   │   ├── nats/           ← NATS client + EventPublisher
│   │   ├── onboarding/     ← onboarding model / store / server / handlers
│   │   ├── organisation/   ← organisation model / store / server
│   │   ├── rbac/           ← role model / store / server
│   │   ├── sharing/        ← resource sharing model / store / server
│   │   ├── shift/          ← shift + schedule + break model / stores / servers
│   │   ├── tag/            ← tag store / server
│   │   ├── team/           ← team model / store / server
│   │   └── user/           ← user + passkey model / store / server
│   ├── pkg/
│   │   ├── db/             ← GORM database initialisation
│   │   ├── pb/             ← generated Go protobuf stubs (do not edit)
│   │   ├── service/        ← auth infrastructure (JWT, Firebase, WebAuthn, interceptors)
│   │   ├── utils/          ← email (Resend), SQS, shared helpers
│   │   └── k/              ← constants / shared keys
│   ├── deploy/             ← Dockerfiles, docker-compose, ECS configs, Envoy
│   ├── envs/               ← environment variable files (.env)
│   └── certs/              ← TLS certificates per environment
├── kylaFE/
│   ├── src/
│   │   ├── app/            ← page-level components
│   │   ├── components/     ← reusable UI (shadcn/ui based)
│   │   ├── hooks/          ← custom React hooks
│   │   ├── lib/            ← utility functions
│   │   ├── pages/          ← route-level pages
│   │   ├── routes/         ← React Router route definitions
│   │   └── pb/             ← generated TypeScript protobuf stubs (do not edit)
│   └── public/
└── kylaMB/
    └── pb/                 ← generated Dart protobuf stubs (do not edit)
```

---

## Tech Stack

### kylaBE (Go)
- **Framework**: Gin (REST) + gRPC (`google.golang.org/grpc`)
- **ORM**: GORM with PostgreSQL (`gorm.io/driver/postgres`)
- **Auth**: JWT (`golang-jwt/jwt`) + Firebase Auth + WebAuthn (passkeys)
- **MFA**: TOTP via `pquerna/otp`
- **Cache**: Redis (`redis/go-redis/v9`)
- **Config**: Viper (`spf13/viper`) — reads from `envs/.env`
- **Messaging**: NATS (`nats-io/nats.go`) — EventPublisher in `internal/nats/`
- **Queue**: AWS SQS (`aws/aws-sdk-go-v2`)
- **Proxy**: Envoy sidecar (see `envoy.yaml`)
- **Live reload**: Air (`deploy/.air.toml`)
- **UUIDs**: `google/uuid`

### kylaFE (TypeScript)
- **Framework**: React 19 + React Router v7
- **Build**: Vite 7
- **UI**: shadcn/ui components (Radix UI primitives + Tailwind CSS v4)
- **Tables**: TanStack Table v8
- **Icons**: Tabler Icons + Lucide React
- **Forms/Validation**: Zod v4
- **Charts**: Recharts
- **Package manager**: pnpm

### kylaPB (Proto3)
- Well-known types: `google.protobuf.Timestamp`, optional fields via `--experimental_allow_proto3_optional`
- All `.proto` files live exclusively in `kylaPB/` — never copy them elsewhere

---

## Common Commands

### Protobuf Generation
```sh
make proto          # Generate stubs for all targets (Go + TypeScript + Dart)
make proto-go       # Go only   → kylaBE/pkg/pb/
make proto-ts       # TS only   → kylaFE/src/pb/
make proto-dart     # Dart only → kylaMB/pb/
make help           # Show all make targets with descriptions
```

### Backend (kylaBE)
```sh
cd kylaBE

# Run with live reload
air

# Run directly
go run cmd/server/main.go

# Build binary
go build -o bin/server cmd/server/main.go

# Tidy dependencies
go mod tidy

# Run with Docker Compose (local)
docker compose -f deploy/docker-compose.yaml up
```

### Frontend (kylaFE)
```sh
cd kylaFE

pnpm dev        # Start dev server (Vite)
pnpm build      # Production build
pnpm lint       # ESLint
pnpm preview    # Preview production build
```

---

## Architecture Patterns

### Backend — gRPC Service Pattern
Each domain lives in its own package under `internal/{domain}/` and follows a consistent four-file layout:

```
internal/{domain}/
  model.go      ← GORM model struct(s)
  store.go      ← database access layer (CRUD against DB)
  server.go     ← gRPC server implementation (implements pb interface)
  utils.go      ← helpers / converters (model ↔ proto)
```

Example for `user`:
- `internal/user/model.go` — `User`, `Passkey` structs (GORM models)
- `internal/user/store.go` — `UserStore` with DB methods
- `internal/user/server.go` — `UserServer` implementing `pb.UserServiceServer`
- `internal/user/utils.go` — conversion functions

**Never** put database logic in the server file. **Never** put gRPC logic in the store file.

### Interface-at-Boundary Pattern
Each domain server defines a minimal local `AuthGateway` interface describing only the auth methods it actually uses. `internal/auth.AuthStore` satisfies all of them without requiring domain packages to import each other. This prevents import cycles.

```go
// Example: internal/branch/server.go
type AuthGateway interface {
    ScopeCheck(ctx context.Context, scopeID string) (bool, *authctx.RequestMetadata, error)
    GetServiceAuthMetadata(ctx context.Context) (*authctx.RequestMetadata, error)
}
```

`auth.AuthStore` is constructed in `main.go` by wrapping a `*service.AuthStore` (the auth infrastructure layer in `pkg/service/`) together with the domain `*rbac.RbacStore`. Domain servers never import `pkg/service` directly.

### Event Publishing (NATS)
The NATS `EventPublisher` lives in `internal/nats/`. Inject it into servers that need to emit events:

```go
natsClient, _ := nats.Connect(configs.EnvConfigs.NatsURL)
pub := nats.NewEventPublisher(natsClient)
// pass pub to org/auth/invitation servers
```

Available subjects: `org.created`, `user.created`, `auth.login`, `invitation.sent`.

### gRPC Server Registration (main.go)
1. Load config via `config.LoadConfig()`
2. Initialise DB via `db.InitDB()`
3. Instantiate stores (`NewXxxStore(db.DB)`)
4. Instantiate servers (`NewXxxServer(store, ...)`)
5. Register servers on the gRPC instance (`pb.RegisterXxxServiceServer(grpcServer, server)`)
6. Run gRPC and Gin servers concurrently via `sync.WaitGroup`

### Auth
- JWT is validated in `AuthInterceptor` (gRPC unary + stream interceptors)
- Firebase Auth is used for social / device login flows
- WebAuthn passkeys are stored in `Passkey` model linked to `User`
- MFA uses TOTP secrets stored on the `User` model

### Environment / Config
- Config is loaded by Viper from `envs/.env`
- Different TLS credentials are loaded based on the `ENVIRONMENT` value:
  `local` → insecure, `development` / `staging` / `production` → TLS from `certs/`

### Frontend — Component Conventions
- UI primitives live in `src/components/ui/` (shadcn/ui generated, minimal modifications)
- Page-level layout in `src/app/`, route pages in `src/pages/`
- Custom hooks in `src/hooks/`, shared utilities in `src/lib/`
- Proto-generated types in `src/pb/` are imported directly for gRPC-web calls — never redeclare proto types manually

---

## Generated Code Rules

> **Do not manually edit files inside any `pb/` directory.**

All files under `kylaBE/pkg/pb/`, `kylaFE/src/pb/`, and `kylaMB/pb/` are machine-generated from `kylaPB/*.proto`. Changes must be made to the `.proto` source and then regenerated with `make proto`.

---

## Environment Variables

Key variables expected in `kylaBE/envs/.env`:

```
ENVIRONMENT=local|development|staging|production
PORT=<grpc port>
POSTGRES_HOST / POSTGRES_PORT / POSTGRES_DB / POSTGRES_USER / POSTGRES_PASS
JWT_SECRET_KEY
AUTH_SVC_URL
RESEND_API_KEY / RESEND_FROM_EMAIL / RESEND_SUPPORT_EMAIL / RESEND_BASE_URL
FB_CREDENTIALS
WEB_AUTHN_RP_ID / WEB_AUTHN_RP_ORIGIN / WEB_AUTHN_RP_DISPLAY_NAME
REDIS_ADDR / REDIS_PASSWORD / REDIS_DB
AWS_REGION / AWS_ACCESS_KEY / AWS_SECRET_KEY
NATS_URL                     # defaults to nats://localhost:4222
```

---

## Adding a New Domain (Backend Checklist)

- [ ] Define the service in `kylaPB/<domain>.proto`
- [ ] Run `make proto-go` to regenerate Go stubs
- [ ] Create `internal/<domain>/model.go` (GORM model)
- [ ] Create `internal/<domain>/store.go` (DB layer)
- [ ] Create `internal/<domain>/server.go` (gRPC impl; define a local `AuthGateway` interface)
- [ ] Create `internal/<domain>/utils.go` (converters)
- [ ] Register the server in `cmd/server/main.go`
- [ ] Add DB auto-migration in the migration block in `main.go`

## Adding a New Domain (Frontend Checklist)

- [ ] Define the service in `kylaPB/<domain>.proto`
- [ ] Run `make proto-ts` to regenerate TypeScript stubs
- [ ] Import types from `src/pb/<domain>` — never redeclare
- [ ] Add page component under `src/pages/`
- [ ] Register route in `src/routes/`
