# Tollgate developer entry points. Run `make help` for the list.
SHELL := bash
.SHELLFLAGS := -euo pipefail -c
.DEFAULT_GOAL := help
MAKEFLAGS += --no-print-directory

GO ?= go
GOTOOL := $(GO) tool -modfile=tools/go.mod
PRE_COMMIT ?= pre-commit

.PHONY: help
help: ## Show available targets
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_-]+:.*## / {printf "  %-14s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: dev-setup
dev-setup: ## Install pinned dev tools (needs git, Go and uv; no Homebrew)
	@scripts/dev-setup.sh

.PHONY: fmt
fmt: ## Format Go code
	$(GOTOOL) golangci-lint fmt ./...

.PHONY: lint
lint: ## Run every linter (Go, vulnerabilities, workflows, shell, YAML, Markdown, editorconfig)
	$(GOTOOL) golangci-lint run ./...
	$(GOTOOL) govulncheck ./...
	$(PRE_COMMIT) run --all-files --show-diff-on-failure

.PHONY: test
test: ## Run Go tests with the race detector
	$(GO) test -race -count=1 ./...

.PHONY: bench-micro
bench-micro: ## Run Go micro-benchmarks
	$(GO) test -run '^$$' -bench . -benchmem ./...

.PHONY: sdk-compat
sdk-compat: ## Run the official OpenAI and Anthropic SDKs (Python, TypeScript, Go) against the simulators
	TARGET=$(or $(TARGET),sim) scripts/sdk-compat.sh

.PHONY: up
up: ## Start the local Compose profile
	PROFILE=$(or $(PROFILE),local) scripts/stack.sh up

.PHONY: down
down: ## Stop the local Compose profile and remove its volumes
	PROFILE=$(or $(PROFILE),local) scripts/stack.sh down

.PHONY: smoke
smoke: ## Exercise non-streaming and streaming requests through the local gateway
	scripts/smoke.sh

.PHONY: e2e
e2e: ## Run end-to-end tests against the running local stack
	TOLLGATE_E2E=1 $(GO) test ./test/e2e/... -count=1
