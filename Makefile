.DEFAULT_GOAL := help
.PHONY: help check docs-check shell-lint actions-lint

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

# Every gate. CI calls this same target (ADR-0004); each check is listed in docs/dev/harness.md.
check: docs-check shell-lint actions-lint ## Run every gate

docs-check: ## Verify links, paths and the index in Markdown (scripts/check-docs.sh)
	bash scripts/check-docs.sh

shell-lint: ## shellcheck the scripts (install: apt install shellcheck, or brew install shellcheck)
	@command -v shellcheck >/dev/null || { echo "shellcheck not found: apt install shellcheck, or brew install shellcheck"; exit 1; }
	shellcheck scripts/*.sh

actions-lint: ## actionlint the GitHub workflows, with a local actionlint or the pinned image
	@v=$$(awk '$$1 == "actionlint" { print $$2 }' .tool-versions); \
	if command -v actionlint >/dev/null; then actionlint; \
	elif command -v docker >/dev/null; then docker run --rm -v "$$PWD:/repo" -w /repo "rhysd/actionlint:$$v"; \
	else echo "actions-lint: needs actionlint $$v or docker"; exit 1; fi
