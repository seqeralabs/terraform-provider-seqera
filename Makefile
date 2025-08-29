.PHONY: help build test lint generate clean install tf-plan tf-apply

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the provider binary
	go build -o terraform-provider-seqera

run: ## Generate provider code using Speakeasy and build binary for testing
	speakeasy run --skip-versioning
	go build -o terraform-provider-seqera

test: ## Run tests
	go test ./...

lint: ## Run linters (if golangci-lint is available)
	@which golangci-lint > /dev/null && golangci-lint run || echo "golangci-lint not found, skipping..."

clean: ## Clean build artifacts
	rm -f terraform-provider-seqera

tf-plan: ## Run terraform plan in test directory
	cd test/hello-world-tests && terraform plan

tf-apply: ## Run terraform apply in test directory
	cd test/hello-world-tests && terraform apply

tf-destroy: ## Run terraform destroy in test directory
	cd test/hello-world-tests && terraform destroy
