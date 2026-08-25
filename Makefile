.PHONY: help build run test lint fmt vet clean \
        spec-version fetch-spec format-spec update-spec-online docs-diff \
        tf-plan tf-apply tf-destroy

SPEC_URL ?= https://cloud.seqera.io/openapi/seqera-api-latest-flattened.yml
SPEC     := specs/seqera-api-cloud.yaml
SORT_CFG := specs/.openapi-format-sort.yaml
BINARY   := terraform-provider-seqera
TF_DIR   := test/hello-world-tests

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ---------------------------------------------------------------------------
# Build / test
# ---------------------------------------------------------------------------

build: ## Build the provider binary
	go build -o $(BINARY)

run: ## Regenerate provider code with Speakeasy, then build
	speakeasy run --skip-versioning
	go build -o $(BINARY)

test: ## Run tests
	go test ./...

fmt: ## Format Go sources
	gofmt -l -w internal main.go

vet: ## Run go vet
	go vet ./...

lint: ## Run linters (skipped if golangci-lint is unavailable)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

clean: ## Clean build artifacts
	rm -f $(BINARY)

# ---------------------------------------------------------------------------
# API spec
#
# See docs-internal/SPEC_UPDATE_RUNBOOK.md. `update-spec-online` handles the
# mechanical half of a bump; triaging what reached the Terraform surface is
# still a human step.
# ---------------------------------------------------------------------------

# Extracts `version:` from the info block. Uses awk rather than `grep -m1`
# because closing a curl pipe early makes curl exit 56 and print a spurious
# write error.
spec-version: ## Print the vendored spec version and the latest published one
	@echo "vendored: $$(awk '/^info:/{f=1} f && /^  version:/{sub(/.*version: */,""); print; exit}' $(SPEC))"
	@tmp=$$(mktemp); \
	if curl -fsSL -o "$$tmp" "$(SPEC_URL)"; then \
		echo "upstream: $$(awk '/^info:/{f=1} f && /^  version:/{sub(/.*version: */,""); print; exit}' "$$tmp")"; \
	else \
		echo "upstream: (fetch failed)"; \
	fi; \
	rm -f "$$tmp"

fetch-spec: ## Download the latest OpenAPI spec and sort it for a clean diff
	@if grep -q 'x-speakeasy' $(SPEC); then \
		echo "ERROR: $(SPEC) contains x-speakeasy annotations."; \
		echo "       Overwriting would lose them. Move them into overlays/ first."; \
		exit 1; \
	fi
	@old=$$(awk '/^info:/{f=1} f && /^  version:/{sub(/.*version: */,""); print; exit}' $(SPEC)); \
	tmp=$$(mktemp); \
	echo "Fetching $(SPEC_URL)"; \
	if ! curl -fsSL -o "$$tmp" "$(SPEC_URL)"; then \
		echo "ERROR: download failed; $(SPEC) left untouched."; rm -f "$$tmp"; exit 1; \
	fi; \
	cp -f "$$tmp" $(SPEC); rm -f "$$tmp"; \
	$(MAKE) --no-print-directory format-spec; \
	new=$$(awk '/^info:/{f=1} f && /^  version:/{sub(/.*version: */,""); print; exit}' $(SPEC)); \
	echo "Spec updated: $$old -> $$new"

format-spec: ## Sort OpenAPI spec for clean diffs
	npx --yes openapi-format $(SPEC) -o $(SPEC) --sortComponentsProps -s $(SORT_CFG) --yamlQuoteStyle single

update-spec-online: ## Fetch the latest spec, regenerate, build, and show what changed
	@$(MAKE) --no-print-directory fetch-spec
	@echo ""
	@echo "==> Spec diff"
	@git diff --stat -- $(SPEC) || true
	@echo ""
	@echo "==> Regenerating"
	@$(MAKE) --no-print-directory run
	@echo ""
	@$(MAKE) --no-print-directory docs-diff
	@echo ""
	@echo "Next: triage every added attribute (keep / ignore / validate / constrain /"
	@echo "sync-description) per docs-internal/SPEC_UPDATE_RUNBOOK.md, then commit."
	@echo "Nothing has been committed."

docs-diff: ## Show which resource docs changed (the spec-bump review surface)
	@echo "==> Generated docs diff"
	@git diff --stat -- docs/ || true
	@echo ""
	@echo "New/changed attributes:"
	@attrs=$$(git diff -U0 -- docs/ | grep -E '^\+- `' | sed 's/^+/  /'); \
	if [ -n "$$attrs" ]; then echo "$$attrs"; else echo "  (none)"; fi

# ---------------------------------------------------------------------------
# Terraform smoke tests
# ---------------------------------------------------------------------------

tf-plan: ## Run terraform plan in test directory
	cd $(TF_DIR) && terraform plan

tf-apply: ## Run terraform apply in test directory
	cd $(TF_DIR) && terraform apply

tf-destroy: ## Run terraform destroy in test directory
	cd $(TF_DIR) && terraform destroy
