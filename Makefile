# Main Makefile for the project

# ─── Colors ───────────────────────────────────────────────────────────────────
RESET   := \033[0m
BOLD    := \033[1m
GREEN   := \033[0;32m
CYAN    := \033[0;36m
YELLOW  := \033[0;33m
BLUE    := \033[0;34m
MAGENTA := \033[0;35m
GRAY    := \033[0;90m

# ─── Paths ────────────────────────────────────────────────────────────────────
PROTO_SRC := kylaPB

BE_PB_OUT := kylaBE/pkg/pb
FE_PB_OUT := kylaFE/src/pb
MB_PB_OUT := kylaMB/pb

# ─── Targets ──────────────────────────────────────────────────────────────────
.PHONY: proto proto-go proto-ts proto-dart help

## Generate all protobuf stubs (Go + TypeScript + Dart)
proto: proto-go proto-ts proto-dart
	@printf "$(BOLD)$(GREEN)✓ All protobuf stubs generated successfully.$(RESET)\n"

# ── Go (kylaBE) ───────────────────────────────────────────────────────────────
proto-go:
	@printf "$(CYAN)$(BOLD)❯ Generating Go protobuf stubs...$(RESET)\n"
	@mkdir -p $(BE_PB_OUT)
	@protoc \
		--proto_path=$(PROTO_SRC) \
		--go_out=$(BE_PB_OUT) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(BE_PB_OUT) \
		--go-grpc_opt=paths=source_relative \
		--experimental_allow_proto3_optional \
		$(PROTO_SRC)/*.proto
	@printf "$(GREEN)✓ Go stubs$(RESET)    → $(GRAY)$(BE_PB_OUT)$(RESET)\n"

# ── TypeScript (kylaFE) ───────────────────────────────────────────────────────
proto-ts:
	@printf "$(BLUE)$(BOLD)❯ Generating TypeScript protobuf stubs...$(RESET)\n"
	@mkdir -p $(FE_PB_OUT)
	@cd kylaFE && pnpm proto
	@printf "$(GREEN)✓ TypeScript stubs$(RESET) → $(GRAY)$(FE_PB_OUT)$(RESET)\n"

# ── Dart (kylaMB) ─────────────────────────────────────────────────────────────
proto-dart:
	@printf "$(MAGENTA)$(BOLD)❯ Generating Dart protobuf stubs...$(RESET)\n"
	@mkdir -p $(MB_PB_OUT)
	@protoc \
		--proto_path=$(PROTO_SRC) \
		--dart_out=grpc:$(MB_PB_OUT) \
		$(PROTO_SRC)/*.proto
	@printf "$(GREEN)✓ Dart stubs$(RESET)      → $(GRAY)$(MB_PB_OUT)$(RESET)\n"

# ─── Help ─────────────────────────────────────────────────────────────────────
help:
	@printf "\n$(BOLD)kylaCX — Protobuf Code Generation$(RESET)\n"
	@printf "$(GRAY)──────────────────────────────────────────────────$(RESET)\n"
	@printf "\n$(BOLD)Usage:$(RESET)  make $(CYAN)<target>$(RESET)\n\n"
	@printf "$(BOLD)Targets:$(RESET)\n"
	@printf "  $(CYAN)%-18s$(RESET) %s\n" "proto"       "Generate stubs for all targets (Go + TypeScript + Dart)"
	@printf "  $(CYAN)%-18s$(RESET) %s $(GRAY)→ $(BE_PB_OUT)$(RESET)\n" "proto-go"    "Generate Go stubs only"
	@printf "  $(CYAN)%-18s$(RESET) %s $(GRAY)→ $(FE_PB_OUT)$(RESET)\n" "proto-ts"    "Generate TypeScript stubs"
	@printf "  $(CYAN)%-18s$(RESET) %s $(GRAY)→ $(MB_PB_OUT)$(RESET)\n" "proto-dart"  "Generate Dart stubs"
	@printf "  $(CYAN)%-18s$(RESET) %s\n"  "help"        "Show this help message"
	@printf "\n$(BOLD)Proto source:$(RESET)  $(GRAY)$(PROTO_SRC)/$(RESET)\n"
	@printf "\n"