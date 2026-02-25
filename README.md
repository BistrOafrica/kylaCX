# kylaCX

A full-stack customer experience platform built as a monorepo. kylaCX provides the backend services, web frontend, mobile application, and shared protobuf definitions that power a modern CX product.

---

## Monorepo Structure

```
kylaCX/
├── kylaBE/   — Go gRPC + REST backend
├── kylaFE/   — TypeScript / React 19 web frontend
├── kylaMB/   — Dart / Flutter mobile app
├── kylaEX/   — Extensions & integrations (TBD)
├── kylaRM/   — Remote / edge module (TBD)
├── kylaPB/   — Shared Protobuf source of truth
└── Makefile  — Root-level commands (proto generation)
```

---

## Prerequisites

| Tool | Version | Notes |
|---|---|---|
| Go | 1.23+ | Backend |
| Node.js | 20+ | Frontend (via pnpm) |
| pnpm | 9+ | Frontend package manager |
| Dart / Flutter | 3+ | Mobile |
| protoc | 3+ | Protobuf compiler |
| protoc-gen-go | latest | `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest` |
| protoc-gen-go-grpc | latest | `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest` |
| protoc-gen-ts_proto | via pnpm | installed in `kylaFE/node_modules` |
| protoc-gen-dart | latest | `dart pub global activate protoc_plugin` |
| Air | latest | Live reload — `go install github.com/air-verse/air@latest` |
| Docker + Compose | latest | Local infra |

---

## Getting Started

### 1. Clone the repo

```sh
git clone <repo-url> kylaCX
cd kylaCX
```

### 2. Generate Protobuf stubs

All `.proto` files live in `kylaPB/`. Generated stubs are written to each project's `pb/` directory and **must not be edited manually**.

```sh
make proto          # Generate for all targets (Go + TypeScript + Dart)
make proto-go       # Go only   → kylaBE/pkg/pb/
make proto-ts       # TS only   → kylaFE/src/pb/
make proto-dart     # Dart only → kylaMB/pb/
```

### 3. Configure the backend

Copy the example env file and fill in your values:

```sh
cp kylaBE/envs/.env.example kylaBE/envs/.env
```

| Variable | Description |
|---|---|
| `ENVIRONMENT` | `local` \| `development` \| `staging` \| `production` |
| `PORT` | gRPC server port (e.g. `50050`) |
| `POSTGRES_HOST` | PostgreSQL host |
| `POSTGRES_PORT` | PostgreSQL port |
| `POSTGRES_DB` | Database name |
| `POSTGRES_USER` | Database user |
| `POSTGRES_PASS` | Database password |
| `JWT_SECRET_KEY` | Secret used to sign JWT tokens |
| `AUTH_SVC_URL` | Auth service endpoint |
| `RESEND_API_KEY` | Resend email API key |
| `RESEND_FROM_EMAIL` | Sender address |
| `RESEND_SUPPORT_EMAIL` | Support address |
| `RESEND_BASE_URL` | Base URL for email links |
| `FB_CREDENTIALS` | Firebase service account JSON (stringified) |
| `WEB_AUTHN_RP_ID` | WebAuthn relying party ID |
| `WEB_AUTHN_RP_ORIGIN` | WebAuthn relying party origin |
| `WEB_AUTHN_RP_DISPLAY_NAME` | WebAuthn display name |
| `REDIS_ADDR` | Redis address (e.g. `localhost:6379`) |
| `REDIS_PASSWORD` | Redis password |
| `REDIS_DB` | Redis database index |
| `AWS_REGION` | AWS region for SQS |
| `AWS_ACCESS_KEY` | AWS access key |
| `AWS_SECRET_KEY` | AWS secret key |

---

## Running the Backend (kylaBE)

### With live reload (recommended for development)

```sh
cd kylaBE
air
```

Air watches `.go` and `.proto` files and rebuilds automatically. Config is in `deploy/.air.toml`.

### Without live reload

```sh
cd kylaBE
go run cmd/server/main.go
```

### With Docker Compose

Starts the gRPC server and an Envoy proxy sidecar:

```sh
cd kylaBE
docker compose -f deploy/docker-compose.yaml up
```

| Service | Port |
|---|---|
| gRPC server | `50050` |
| Envoy proxy (gRPC-web) | `8000` |
| Envoy admin | `9903` |
| Gin REST API | `8085` |

### Build binary

```sh
cd kylaBE
go build -o bin/server cmd/server/main.go
```

---

## Running the Frontend (kylaFE)

```sh
cd kylaFE
pnpm install    # first time only
pnpm dev        # start Vite dev server
pnpm build      # production build
pnpm preview    # preview production build locally
pnpm lint       # run ESLint
```

---

## Running the Mobile App (kylaMB)

```sh
cd kylaMB
flutter pub get
flutter run
```

---

## Architecture

### Backend service pattern

Each domain in `kylaBE/pkg/service/` follows a strict four-file layout:

```
{domain}.go           ← GORM model structs
{domain}_store.go     ← database access layer
{domain}_server.go    ← gRPC server (implements pb interface)
{domain}_utils.go     ← converters between model and proto types
```

Database logic stays in `_store.go`. gRPC logic stays in `_server.go`. Cross-contamination between the two is a bug.

### Auth

- **JWT** — validated on every request via `AuthInterceptor` (gRPC unary + stream)
- **Firebase Auth** — social and device login flows
- **WebAuthn** — passkey credentials stored on the `Passkey` model
- **TOTP MFA** — secrets and recovery codes stored on the `User` model

### Protobuf & gRPC-web

The Envoy sidecar translates gRPC-web requests from the browser into gRPC calls to the Go backend. Frontend clients talk to Envoy on port `8000`; the backend only exposes gRPC on `50050`.

### TLS

Certificates are loaded from `kylaBE/certs/` depending on `ENVIRONMENT`:

| Value | Credential |
|---|---|
| `local` | Insecure (no TLS) |
| `development` | `certs/dev/` |
| `staging` | `certs/staging/` |
| `production` | `certs/prod/` |

---

## Adding a New Domain

### Backend checklist

- [ ] Define the service in `kylaPB/<domain>.proto`
- [ ] Run `make proto-go`
- [ ] Create `kylaBE/pkg/service/<domain>.go` — GORM model
- [ ] Create `kylaBE/pkg/service/<domain>_store.go` — DB layer
- [ ] Create `kylaBE/pkg/service/<domain>_server.go` — gRPC impl
- [ ] Create `kylaBE/pkg/service/<domain>_utils.go` — model ↔ proto converters
- [ ] Register the server in `kylaBE/cmd/server/main.go`
- [ ] Add `db.DB.AutoMigrate(&service.YourModel{})` in the migration block

### Frontend checklist

- [ ] Define the service in `kylaPB/<domain>.proto`
- [ ] Run `make proto-ts`
- [ ] Import generated types from `src/pb/<domain>` — never redeclare manually
- [ ] Add page component under `src/pages/`
- [ ] Register route in `src/routes/`

---

## Key Conventions

- **Never edit anything inside a `pb/` directory** — all files are machine-generated
- **Proto source of truth** is `kylaPB/` — changes go in `.proto` files, not generated stubs
- DB logic lives exclusively in `_store.go` files
- gRPC handler logic lives exclusively in `_server.go` files
- Frontend proto types are imported from `src/pb/`, never re-typed manually
