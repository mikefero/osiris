# Used to source the .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

SHELL := /usr/bin/env bash

.DEFAULT_GOAL := help

# Use https://patorjk.com/software/taag/#p=display&f=ANSI%20Shadow&t=App-Name to generate
# ANSI Shadow title for application name
.PHONY: help
help: ## Display this help screen
	@echo ''
	@echo ' █████╗ ██████╗ ██████╗       ███╗   ██╗ █████╗ ███╗   ███╗███████╗'
	@echo '██╔══██╗██╔══██╗██╔══██╗      ████╗  ██║██╔══██╗████╗ ████║██╔════╝'
	@echo '███████║██████╔╝██████╔╝█████╗██╔██╗ ██║███████║██╔████╔██║█████╗'
	@echo '██╔══██║██╔═══╝ ██╔═══╝ ╚════╝██║╚██╗██║██╔══██║██║╚██╔╝██║██╔══╝'
	@echo '██║  ██║██║     ██║           ██║ ╚████║██║  ██║██║ ╚═╝ ██║███████╗'
	@echo '╚═╝  ╚═╝╚═╝     ╚═╝           ╚═╝  ╚═══╝╚═╝  ╚═╝╚═╝     ╚═╝╚══════'
	@echo ''
	@# Display top-level targets since they are the ones most developes will need.
	@grep -h -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort -k1 | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@# Now show hierarchical targets in separate sections.
	@grep -h -E '^[a-zA-Z0-9_-]+/[a-zA-Z0-9/_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk '{print $$1}' | \
		awk -F/ '{print $$1}' | \
		sort -u | \
	while read section ; do \
		echo; \
		grep -h -E "^$$section/[^:]+:.*?## .*$$" $(MAKEFILE_LIST) | sort -k1 | \
			awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' ; \
	done

APP_NAME := app-name
APP_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
APP_WORKDIR := $(shell pwd)
APP_PACKAGE := github.com/mikefero/go-template/internal/cmd

include $(APP_DIR)/mk/build.mk
include $(APP_DIR)/mk/common.mk
include $(APP_DIR)/mk/dev.mk
include $(APP_DIR)/mk/test.mk
include $(APP_DIR)/mk/tools.mk
