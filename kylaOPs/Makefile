# kylaCX — infra Makefile.
#
# All Docker/compose lifecycle commands live here. Run from kylaOPs/.

SHELL := /usr/bin/env bash

# ─── Paths ────────────────────────────────────────────────────────────────────
OPS_DIR   := $(abspath $(dir $(firstword $(MAKEFILE_LIST))))
ROOT_DIR  := $(abspath $(OPS_DIR)/..)
COMPOSE_FILE := $(OPS_DIR)/compose/docker-compose.yml
ENV_FILE     := $(OPS_DIR)/envs/.env

# ─── Compose invocation ──────────────────────────────────────────────────────
# --env-file points compose at our top-level .env (host ports, build args).
# We deliberately do NOT pass --project-directory: compose then resolves
# build contexts and env_files relative to the compose file location
# (kylaOPs/compose/), which is what every path in docker-compose.yml assumes.
COMPOSE := docker compose \
	--env-file $(ENV_FILE) \
	-f $(COMPOSE_FILE)

# ─── Colours ──────────────────────────────────────────────────────────────────
RESET   := \033[0m
BOLD    := \033[1m
GREEN   := \033[0;32m
CYAN    := \033[0;36m
YELLOW  := \033[0;33m
GRAY    := \033[0;90m
RED     := \033[0;31m

# ─── Default ─────────────────────────────────────────────────────────────────
.DEFAULT_GOAL := help

# ─── Phony targets ───────────────────────────────────────────────────────────
.PHONY: help init env check up down deploy restart stop start build rebuild \
	pull ps status logs logs-% top exec-% shell-% sh-% \
	psql redis-cli nats-info temporal-ui \
	clean nuke prune networks volumes \
	frontend-rebuild backend-rebuild

# ─── Help ─────────────────────────────────────────────────────────────────────
help: ## Show this help
	@printf "\n$(BOLD)kylaCX — infra commands$(RESET)\n"
	@printf "$(GRAY)─────────────────────────────────────────────────────────$(RESET)\n\n"
	@printf "$(BOLD)Usage:$(RESET)  make $(CYAN)<target>$(RESET)\n\n"
	@awk 'BEGIN {FS = ":.*?## "} \
		/^[a-zA-Z_-]+:.*?## / { printf "  $(CYAN)%-22s$(RESET) %s\n", $$1, $$2 } \
		/^# ── / { printf "\n$(BOLD)%s$(RESET)\n", substr($$0, 5) }' \
		$(MAKEFILE_LIST)
	@printf "\n"

# ── Bootstrap ────────────────────────────────────────────────────────────────
init: ## Copy every envs/*.env.example → envs/*.env (idempotent)
	@$(OPS_DIR)/scripts/init.sh

check: ## Verify all required .env files exist
	@missing=0; \
	for f in .env postgres.env redis.env nats.env temporal.env \
	         temporal-postgres.env backend.env freeswitch.env coturn.env; do \
		if [ ! -f "$(OPS_DIR)/envs/$$f" ]; then \
			printf "$(RED)missing$(RESET)  envs/$$f\n"; \
			missing=$$((missing + 1)); \
		fi; \
	done; \
	if [ "$$missing" -gt 0 ]; then \
		printf "\n$(RED)$$missing env file(s) missing — run 'make init'.$(RESET)\n"; \
		exit 1; \
	fi; \
	printf "$(GREEN)✓ all env files present$(RESET)\n"

env: ## Show resolved compose config (after env interpolation)
	@$(COMPOSE) config

# ── Lifecycle ────────────────────────────────────────────────────────────────
up: check ## Start the full stack in the background
	@$(COMPOSE) up -d --remove-orphans
	@printf "$(GREEN)✓ stack up$(RESET) — $(GRAY)make ps to inspect, make logs to tail$(RESET)\n"

deploy: check build ## Build images and start the full stack
	@$(COMPOSE) up -d --remove-orphans
	@printf "$(GREEN)✓ deploy complete$(RESET)\n"

down: ## Stop and remove containers (preserves volumes)
	@$(COMPOSE) down --remove-orphans

stop: ## Stop containers without removing them
	@$(COMPOSE) stop

start: ## Start previously stopped containers
	@$(COMPOSE) start

restart: ## Restart all services (or one: make restart svc=backend)
ifeq ($(svc),)
	@$(COMPOSE) restart
else
	@$(COMPOSE) restart $(svc)
endif

# ── Build ────────────────────────────────────────────────────────────────────
build: check ## Build all images (or one: make build svc=backend)
ifeq ($(svc),)
	@$(COMPOSE) build
else
	@$(COMPOSE) build $(svc)
endif

rebuild: ## Build with --no-cache (or one: make rebuild svc=backend)
ifeq ($(svc),)
	@$(COMPOSE) build --no-cache
else
	@$(COMPOSE) build --no-cache $(svc)
endif

frontend-rebuild: ## Rebuild and restart the frontend container (VITE vars baked in)
	@$(COMPOSE) build frontend
	@$(COMPOSE) up -d frontend

backend-rebuild: ## Rebuild and restart the backend container
	@$(COMPOSE) build backend
	@$(COMPOSE) up -d backend

pull: ## Pull all remote images
	@$(COMPOSE) pull

# ── Inspect ──────────────────────────────────────────────────────────────────
ps: ## Show container status
	@$(COMPOSE) ps

status: ps ## Alias for ps

top: ## Show running processes per container
	@$(COMPOSE) top

logs: ## Tail logs for all services (or one: make logs svc=backend)
ifeq ($(svc),)
	@$(COMPOSE) logs -f --tail=200
else
	@$(COMPOSE) logs -f --tail=200 $(svc)
endif

logs-%: ## Tail logs for a specific service: make logs-backend
	@$(COMPOSE) logs -f --tail=200 $*

# ── Shells / clients ─────────────────────────────────────────────────────────
exec-%: ## Run a one-off command in a service: make exec-backend cmd="ls /app"
	@$(COMPOSE) exec $* $(or $(cmd),sh)

shell-%: ## Open an interactive shell in a service: make shell-backend
	@$(COMPOSE) exec $* sh

sh-%: shell-%  ## Alias for shell-%

psql: ## Open psql against the application database
	@$(COMPOSE) exec postgres psql -U $$(grep ^POSTGRES_USER $(OPS_DIR)/envs/postgres.env | cut -d= -f2) \
	                                -d $$(grep ^POSTGRES_DB   $(OPS_DIR)/envs/postgres.env | cut -d= -f2)

redis-cli: ## Open redis-cli against the running redis container
	@$(COMPOSE) exec redis sh -c 'redis-cli -a "$$REDIS_PASSWORD"'

nats-info: ## Print NATS JetStream health / monitoring info
	@$(COMPOSE) exec nats wget -qO- http://localhost:8222/varz | head -40

temporal-ui: ## Bring up the Temporal Web UI on :8088
	@$(COMPOSE) --profile dev up -d temporal-ui
	@printf "$(GREEN)✓ temporal-ui$(RESET) → $(CYAN)http://localhost:8088$(RESET)\n"

# ── Cleanup ──────────────────────────────────────────────────────────────────
clean: ## Stop containers and remove anonymous volumes (keeps named volumes)
	@$(COMPOSE) down --remove-orphans -v --rmi local
	@printf "$(YELLOW)✓ clean$(RESET) — named volumes preserved (use 'make nuke' to wipe data)\n"

nuke: ## DANGER — stop, remove containers AND all named volumes (wipes data)
	@printf "$(RED)This will delete postgres, temporal, redis, nats, and freeswitch data.$(RESET)\n"
	@read -p "Type 'nuke' to confirm: " confirm; \
	if [ "$$confirm" = "nuke" ]; then \
		$(COMPOSE) down --remove-orphans -v --rmi local; \
		docker volume rm -f kyla_postgres_data kyla_temporal_pg_data kyla_redis_data \
		                   kyla_nats_data kyla_freeswitch_recordings 2>/dev/null || true; \
		printf "$(GREEN)✓ nuked$(RESET)\n"; \
	else \
		printf "$(GRAY)aborted$(RESET)\n"; \
	fi

prune: ## Prune unused Docker resources (cluster-wide, not just kyla)
	@docker system prune -f

networks: ## List Docker networks
	@docker network ls --filter name=kyla

volumes: ## List Docker volumes
	@docker volume ls --filter name=kyla
