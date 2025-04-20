# --------------------------------------------------
# Tools tooling
# --------------------------------------------------

GOLANGCI_LINT_VERSION ?= v2.0.2

# Ensure curl and gofumpt are available
ifeq (, $(shell which curl 2> /dev/null))
$(error "'curl' is not installed or available in PATH")
endif

.PHONY: deadcode
deadcode: ## Run deadcode check
	@if [ -x "$(APP_DIR)/bin/deadcode" ]; then \
		"$(APP_DIR)/bin/deadcode" -test ./...; \
	else \
		echo "'deadcode' is not installed, run 'make install-tools'"; \
		exit 1; \
	fi

.PHONY: format
format: ## Format the source code
	@if [ -x "$(APP_DIR)/bin/gofumpt" ]; then \
		"$(APP_DIR)/bin/gofumpt" -l -w .; \
	else \
		echo "'gofumpt' is not installed, run 'make install-tools'"; \
		exit 1; \
	fi

.PHONY: install-tools
install-tools: ## Install required tools
	@mkdir -p "$(APP_DIR)/bin"
	@curl -sSfL \
		"https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh" \
		| sh -s -- -b "$(APP_DIR)/bin" "$(GOLANGCI_LINT_VERSION)"
	@go get
	@GOBIN="$(APP_DIR)/bin" cat tools/tools.go | \
		grep _ | \
			awk -F'"' '{print $$2}' | \
				xargs -tI % sh -c 'GOBIN="$(APP_DIR)/bin" go install %'

.PHONY: lint
lint: ## Lint the source code
	@if [ -x "$(APP_DIR)/bin/golangci-lint" ]; then \
		"$(APP_DIR)/bin/golangci-lint" run ./... --verbose; \
	else \
		echo "'golangci-lint' is not installed, run 'make install-tools'"; \
		exit 1; \
	fi
