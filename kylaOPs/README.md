# kylaOPs

Infra for the kylaCX monorepo — Docker Compose, Dockerfiles, nginx, envoy,
and the Makefile that drives the whole local/staging deploy story.

```
kylaOPs/
├── Makefile                    primary entrypoint — `make help`
├── compose/
│   └── docker-compose.yml      full stack: db, redis, nats, temporal,
│                               freeswitch, coturn, backend, envoy,
│                               frontend nginx, webhook nginx
├── docker/
│   ├── backend/                Go service image (prod + dev)
│   ├── frontend/               Vite build → nginx (SPA + /grpc proxy)
│   ├── nginx-webhook/          public webhook ingress (isolated network)
│   ├── envoy/                  gRPC-web gateway config
│   ├── redis/                  redis.conf
│   └── freeswitch/             SIP/WebRTC config tree
├── envs/                       *.env.example templates (committed)
│                               *.env actual files (gitignored, made by init)
└── scripts/
    └── init.sh                 copies *.env.example → *.env
```

## Quickstart

```sh
cd kylaOPs
make init          # seed envs/*.env from *.env.example
# edit envs/*.env with real secrets (Resend, AWS, Firebase, etc.)
make deploy        # build images + start the stack
make logs          # tail everything
make ps            # show status
```

The SPA is then reachable on `http://localhost`, webhook providers post to
`http://<host>:8080/webhooks/...`, and the Temporal UI (dev profile) on
`http://localhost:8088`.

## Network model

Two bridge networks, deliberately separated:

| Network        | Members                                                              | Public ports |
|----------------|----------------------------------------------------------------------|--------------|
| `kyla`         | postgres, redis, nats, temporal, temporal-postgres, freeswitch, coturn, backend, envoy, frontend | `80` (frontend), SIP/STUN/TURN |
| `kyla_webhook` | nginx-webhook, **backend** (dual-homed)                              | `8080` (webhook) |

The webhook ingress is isolated so a misconfigured or compromised webhook
endpoint cannot reach the internal mesh laterally. Only the backend
container itself bridges the two networks.

## env files

Per-service `env_file` directives (project convention — no inline
`environment:` blocks).

| Service              | env file                       |
|----------------------|--------------------------------|
| postgres             | `envs/postgres.env`            |
| redis                | `envs/redis.env`               |
| nats                 | `envs/nats.env` (empty)        |
| temporal-postgres    | `envs/temporal-postgres.env`   |
| temporal, temporal-ui| `envs/temporal.env`            |
| backend              | `envs/backend.env`             |
| freeswitch           | `envs/freeswitch.env`          |
| coturn               | `envs/coturn.env`              |
| frontend (build-time)| `envs/.env` + build args       |

The frontend is the one exception: Vite bakes `VITE_*` vars into the bundle
at build time, so those come from `envs/.env` and feed `compose build.args`.
Run `make frontend-rebuild` after changing them.

## Makefile cheatsheet

```sh
make init                 # copy *.env.example → *.env
make check                # verify env files exist
make up                   # start stack
make deploy               # build + start
make down                 # stop + remove containers (keeps volumes)
make logs                 # tail all
make logs-backend         # tail one service (or: make logs svc=backend)
make ps
make build [svc=backend]
make rebuild [svc=...]    # --no-cache
make restart [svc=...]
make shell-backend        # interactive shell in a service
make psql                 # psql into the kyla DB
make redis-cli            # redis-cli against the live redis
make temporal-ui          # spin up the Temporal UI on :8088
make clean                # remove containers + local images (keep data)
make nuke                 # also wipe named volumes (asks for confirmation)
```

## Production notes

- `Dockerfile` builds the production-grade backend image. The dev variant
  (`Dockerfile.dev` with `air`) is wired only when you swap it into compose
  manually — the default compose uses production.
- `certs/` under `kylaBE/` is baked into the backend image. Override per
  environment via build args or a derived image.
- The frontend nginx serves `/grpc → envoy:8081` same-origin to dodge
  CORS preflight for gRPC-web. Override `VITE_API_URL` if you front the
  API on a different origin.
- For production, terminate TLS at a public LB / ingress in front of the
  frontend nginx and the webhook nginx. Neither container speaks TLS today.
