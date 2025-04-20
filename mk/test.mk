# --------------------------------------------------
# Test tooling
# --------------------------------------------------

# Determine the packages for coverage ignoring generated code
# COVERAGE_PACKAGES=$(shell go list ./... | grep -v "internal/api")

.PHONY: test
test: ## Run tests
	@go test -v -race ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@go test -race -coverprofile=$(APP_DIR)/coverage.out -covermode=atomic $(shell echo $(COVERAGE_PACKAGES) | tr '\n' ' ')
	@go tool cover -html=$(APP_DIR)/coverage.out -o $(APP_DIR)/coverage.html

.PHONY: test-no-cache
test-no-cache: ## Run tests without cache
	@go test -v -count=1 -race ./...

.PHONY: test-no-race
test-no-race: ## Run tests without race detector
	@go test -v ./...

.PHONY: test-no-cache-no-race
test-no-cache-no-race: ## Run tests without cache and without race detector
	@go test -v -count=1 ./...
