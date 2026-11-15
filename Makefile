.DEFAULT_GOAL := help

BACKEND := backend
# Override when the local toolchain lacks the coverage tools: make check GOTEST_FLAGS=-race
GOTEST_FLAGS ?= -race -cover
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: help check doctor fmt fmt-check lint test build dev-backend \
        docs-check practices-check shell-lint actions-lint \
        backend-fmt backend-fmt-check backend-lint backend-test backend-build backend-deps-check

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Every gate. CI calls this same target (ADR-0004); each check is listed in docs/dev/harness.md.
check: fmt-check lint test build docs-check practices-check ## Run every gate

doctor: ## Compare local Go / golangci-lint / actionlint with .tool-versions
	bash scripts/doctor.sh

fmt: backend-fmt ## Format all code in place
fmt-check: backend-fmt-check ## Verify formatting without changing files
lint: backend-lint shell-lint actions-lint ## Run every linter
test: backend-test ## Run every test
build: backend-build ## Build the API server (backend/bin/server)

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

# --- documents and scripts ---------------------------------------------------

docs-check: ## Verify links, paths and the index in Markdown (scripts/check-docs.sh)
	bash scripts/check-docs.sh

practices-check: ## Verify the shape of the practice index (scripts/check-practices.sh)
	bash scripts/check-practices.sh

shell-lint: ## shellcheck the scripts (install: apt install shellcheck, or brew install shellcheck)
	@command -v shellcheck >/dev/null || { echo "shellcheck not found: apt install shellcheck, or brew install shellcheck"; exit 1; }
	shellcheck scripts/*.sh

actions-lint: ## actionlint the GitHub workflows, with a local actionlint or the pinned image
	@v=$$(awk '$$1 == "actionlint" { print $$2 }' .tool-versions); \
	if command -v actionlint >/dev/null; then actionlint; \
	elif command -v docker >/dev/null; then docker run --rm -v "$$PWD:/repo" -w /repo "rhysd/actionlint:$$v"; \
	else echo "actions-lint: needs actionlint $$v or docker"; exit 1; fi
