.PHONY: setup
setup: ## Install tools
	@echo ">> install golangci-lint"
	curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.12.0

.PHONY: test
test: ## Run unit tests
	go test -race -count=1 -shuffle=on -timeout=12m ./...

.PHONY: lint
lint: ## Run golangci linters
	golangci-lint run --output.text.print-issued-lines=false

.PHONY: fmt
fmt: ## Run golangci formatters
	golangci-lint fmt

.PHONY: help
help: ## Show available targets
	@grep -E '^[a-z]+(-[a-z]+)*:.*?## .+$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS=":.*?## "} {printf "%-12s %s\n", $$1, $$2}'
