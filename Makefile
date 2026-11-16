.DEFAULT_GOAL := help

BACKEND := backend
FRONTEND := frontend
# Override when the local toolchain lacks the coverage tools: make check GOTEST_FLAGS=-race
GOTEST_FLAGS ?= -race -cover
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
# The images take the Go and Node versions from .tool-versions (ADR-0010).
GO_VERSION := $(shell awk '$$1 == "golang" { print $$2 }' .tool-versions)
NODE_VERSION := $(shell awk '$$1 == "nodejs" { print $$2 }' .tool-versions)
WEB_PORT ?= 8080
COMPOSE := GO_VERSION=$(GO_VERSION) NODE_VERSION=$(NODE_VERSION) VERSION=$(VERSION) WEB_PORT=$(WEB_PORT) docker compose

.PHONY: help setup check doctor fmt fmt-check lint test build dev-backend dev-frontend \
        image up down logs smoke dockerfile-lint pr-size \
        docs-check practices-check standards-check bundle-budget shell-lint actions-lint \
        backend-fmt backend-fmt-check backend-lint backend-test backend-build backend-deps-check \
        frontend-fmt frontend-fmt-check frontend-typecheck frontend-lint frontend-test frontend-build frontend-deps-check e2e chromium-check

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Every gate. CI calls this same target (ADR-0004); each check is listed in docs/dev/harness.md.
check: fmt-check lint test build bundle-budget e2e docs-check practices-check standards-check ## Run every gate

setup: ## Install the dependencies (Go modules, npm packages)
	cd $(BACKEND) && go mod download
	cd $(FRONTEND) && npm ci

doctor: ## Compare local Go / Node / golangci-lint / actionlint with .tool-versions
	bash scripts/doctor.sh

fmt: backend-fmt frontend-fmt ## Format all code in place
fmt-check: backend-fmt-check frontend-fmt-check ## Verify formatting without changing files
lint: backend-lint frontend-lint shell-lint actions-lint ## Run every linter
test: backend-test frontend-test ## Run every test
build: backend-build frontend-build ## Build the API server (backend/bin/server) and the screen (frontend/dist)

dev-backend: ## Run the API server on PORT (default 8080); settings come only from the environment
	cd $(BACKEND) && go run ./cmd/server

# --- backend -----------------------------------------------------------------

backend-fmt: ## gofmt in place
	cd $(BACKEND) && gofmt -w .

backend-fmt-check: ## Fail if any Go file is not gofmt-ed
	@cd $(BACKEND) && out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

backend-lint: ## go vet + golangci-lint (the config is verified first)
	cd $(BACKEND) && go vet ./... && golangci-lint config verify && golangci-lint run ./...

backend-test: ## Go tests with the race detector and coverage (GOTEST_FLAGS)
	cd $(BACKEND) && go test $(GOTEST_FLAGS) ./...

backend-build: ## Compile the API server to backend/bin/server
	cd $(BACKEND) && go build -ldflags "-X github.com/hassaku63/chord-rambler/backend/internal/httpapi.Version=$(VERSION)" -o bin/server ./cmd/server

backend-deps-check: ## go.mod / go.sum are tidy (needs the network; CI runs it)
	cd $(BACKEND) && go mod tidy -diff

# --- frontend ----------------------------------------------------------------

dev-frontend: ## Run the Vite dev server on :5173; /api goes to API_UPSTREAM (default http://localhost:8080)
	cd $(FRONTEND) && npm run dev

frontend-fmt: ## prettier --write
	cd $(FRONTEND) && npm run format

frontend-fmt-check: ## prettier --check
	cd $(FRONTEND) && npm run format:check

frontend-typecheck: ## tsc -b, after checking that the generated API types match the contract
	bash scripts/check-api-types.sh
	cd $(FRONTEND) && npm run typecheck

frontend-lint: frontend-typecheck ## eslint, after the type check
	cd $(FRONTEND) && npm run lint

frontend-test: ## vitest
	cd $(FRONTEND) && npm test

frontend-build: ## Type check and build the screen to frontend/dist
	cd $(FRONTEND) && npm run build

frontend-deps-check: ## Production npm dependencies have no high vulnerabilities (needs the network; CI runs it)
	cd $(FRONTEND) && npm audit --omit=dev --audit-level=high

# Needs Chromium: `cd frontend && npx playwright install chromium`, or PW_CHROMIUM_PATH (docs/dev/harness.md).
e2e: chromium-check backend-build ## The screen end to end in Chromium: the scenario, axe-core and the layout (ADR-0017)
	cd $(FRONTEND) && npx playwright test

# Stops with the way to get Chromium when there is none. With PW_CHROMIUM_PATH set, Playwright uses that path as it is.
chromium-check:
	@cd $(FRONTEND) && if [ -z "$$PW_CHROMIUM_PATH" ] && ! [ -x "$$(node -p "require('@playwright/test').chromium.executablePath()")" ]; then \
		echo "e2e: no Chromium for Playwright. Get one with 'cd frontend && npx playwright install chromium'" >&2; \
		echo "     (on Linux without the OS libraries: 'npx playwright install --with-deps chromium')," >&2; \
		echo "     or point PW_CHROMIUM_PATH at a Chromium on this machine (docs/dev/harness.md)." >&2; \
		exit 1; \
	fi

# --- containers (need Docker; CI runs them) ------------------------------------

image: ## Build the api and web images (Docker)
	$(COMPOSE) build

up: ## Build and start the stack on http://localhost:$(WEB_PORT) (Docker)
	$(COMPOSE) up -d --build --wait

down: ## Stop the stack
	$(COMPOSE) down

logs: ## Print the stack's logs
	$(COMPOSE) logs

smoke: ## Smoke test the running stack (scripts/smoke.sh)
	bash scripts/smoke.sh http://localhost:$(WEB_PORT) "$(VERSION)"

dockerfile-lint: ## hadolint the Dockerfile, with a local hadolint or the pinned image (needs one of them; CI runs it)
	@v=$$(awk '$$1 == "hadolint" { print $$2 }' .tool-versions); \
	if command -v hadolint >/dev/null; then hadolint Dockerfile; \
	elif docker info >/dev/null 2>&1; then docker run --rm -i "hadolint/hadolint:v$$v" < Dockerfile; \
	else echo "dockerfile-lint: needs hadolint $$v or a running Docker"; exit 1; fi

# --- size of the change ------------------------------------------------------

PR_BASE ?= origin/main
pr-size: ## Count the lines this branch changes against PR_BASE (default origin/main); warn over 600, fail over 1000
	bash scripts/check-pr-size.sh $(PR_BASE)

# --- size of what loads first ------------------------------------------------

bundle-budget: ## gzip size of what index.html loads up front is within the budget (needs frontend/dist; scripts/check-bundle-size.sh)
	bash scripts/check-bundle-size.sh

# --- documents and scripts ---------------------------------------------------

docs-check: ## Verify links, paths and the index in Markdown (scripts/check-docs.sh)
	bash scripts/check-docs.sh

practices-check: ## Verify the shape of the practice index (scripts/check-practices.sh)
	bash scripts/check-practices.sh

standards-check: ## Verify the shape of the standards index (scripts/check-standards.sh)
	bash scripts/check-standards.sh

shell-lint: ## shellcheck the scripts (install: apt install shellcheck, or brew install shellcheck)
	@command -v shellcheck >/dev/null || { echo "shellcheck not found: apt install shellcheck, or brew install shellcheck"; exit 1; }
	shellcheck scripts/*.sh

actions-lint: ## actionlint the GitHub workflows, with a local actionlint or the pinned image
	@v=$$(awk '$$1 == "actionlint" { print $$2 }' .tool-versions); \
	if command -v actionlint >/dev/null; then actionlint; \
	elif docker info >/dev/null 2>&1; then docker run --rm -v "$$PWD:/repo" -w /repo "rhysd/actionlint:$$v"; \
	else echo "actions-lint: needs actionlint $$v or a running Docker"; exit 1; fi
