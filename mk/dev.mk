# --------------------------------------------------
# Development tooling
# --------------------------------------------------

define APP_LDFLAGS_DEV
-X $(APP_PACKAGE).AppName=$(APP_NAME) \
-X $(APP_PACKAGE).Version=$(APP_VERSION) \
-X $(APP_PACKAGE).Commit=dev \
-X $(APP_PACKAGE).OsArch=$(APP_OS_ARCH) \
-X $(APP_PACKAGE).GoVersion=$(APP_GO_VERSION) \
-X $(APP_PACKAGE).BuildDate=$(APP_BUILD_DATE)
endef

.PHONY: dump
dump: ## Run the dump command
	@CGO_ENABLED=0 go run -ldflags "$(APP_LDFLAGS_DEV)" "$(APP_DIR)" dump

.PHONY: license
license: ## Run the license command
	@CGO_ENABLED=0 go run -ldflags "$(APP_LDFLAGS_DEV)" "$(APP_DIR)" license

.PHONY: reset
reset: ## Run the reset command
	@CGO_ENABLED=0 go run -ldflags "$(APP_LDFLAGS_DEV)" "$(APP_DIR)" reset

.PHONY: version
version: ## Run the version command
	@CGO_ENABLED=0 go run -ldflags "$(APP_LDFLAGS_DEV)" "$(APP_DIR)" version
