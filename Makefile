BIN_DIR=./bin

SHELL := env VERSION=$(VERSION) $(SHELL)
VERSION ?= $(shell git describe --tags $(git rev-list --tags --max-count=1))

APP_NAME?=ge-tax-calc
SHELL := env APP_NAME=$(APP_NAME) $(SHELL)

RELEASE_BRANCH?=master
SHELL := env RELEASE_BRANCH=$(RELEASE_BRANCH) $(SHELL)

COMPOSE_BOT_FILE=deployments/docker-compose/bot-docker-compose.yml
COMPOSE_BOT_CMD_BASE=docker compose -f $(COMPOSE_BOT_FILE)
BOT_ENV_FILE?=deployments/docker-compose/.env

GOVERSION:=1.27.3

SPECKIT_VERSION?=1.4.0
SPECKIT_COMPANION_VERSION?=0.27.0
SPECKIT_BROWNFIELD_VERSION?=1.0.0
SPECKIT_BUGFIX_VERSION?=1.1.0
SPECKIT_FEATURE_NUMBERING_VERSION?=1.0.0
# Space-separated Spec Kit agent integrations; the first one is the default.
SPECKIT_INTEGRATIONS?=claude
SPECKIT_ENV=SPECKIT_VERSION=$(SPECKIT_VERSION) \
	SPECKIT_COMPANION_VERSION=$(SPECKIT_COMPANION_VERSION) \
	SPECKIT_BROWNFIELD_VERSION=$(SPECKIT_BROWNFIELD_VERSION) \
	SPECKIT_BUGFIX_VERSION=$(SPECKIT_BUGFIX_VERSION) \
	SPECKIT_FEATURE_NUMBERING_VERSION=$(SPECKIT_FEATURE_NUMBERING_VERSION) \
	SPECKIT_INTEGRATIONS="$(SPECKIT_INTEGRATIONS)"

TARGET_MAX_CHAR_NUM=20

## Show help
help:
	${call colored, help is running...}
	@echo ''
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  %-$(TARGET_MAX_CHAR_NUM)s %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

## Build project.
build: sync-vendor generate compile-app
.PHONY: build

## Compile app.
compile-app:
	./scripts/build/app.sh
.PHONY: compile-app

## Test coverage report.
test-cover:
	./scripts/tests/coverage.sh
.PHONY: test-cover

prepare-cover-report: test-cover
	./scripts/tests/prepare-cover-report.sh
.PHONY: prepare-cover-report

## Open coverage report.
open-cover-report: prepare-cover-report
	./scripts/open-coverage-report.sh
.PHONY: open-cover-report

## Update readme coverage.
update-readme-cover: build prepare-cover-report
	./scripts/update-readme-coverage.sh
.PHONY: update-readme-cover

## Run tests.
test:
	./scripts/tests/run.sh
.PHONY: test

test-integration:
	./scripts/tests/integration.sh
.PHONY: test-integration

## Run regression tests.
test-regression: test
.PHONY: test-regression

## Sync vendor and install needed tools.
configure: sync-vendor install-tools

## Sync vendor with go.mod.
sync-vendor:
	./scripts/sync-vendor.sh
.PHONY: sync-vendor

## Fix imports sorting.
imports:
	./scripts/style/fix-imports.sh
.PHONY: imports

## Format code with go fmt.
fmt:
	./scripts/style/fmt.sh
.PHONY: fmt

## Format code and sort imports.
format-project: fmt imports
.PHONY: format-project

## Build tools pinned in tools/<name>/go.mod.
install-tools:
	./scripts/install/tools.sh
.PHONY: install-tools

## vet project
vet:
	./scripts/linting/run-vet.sh
.PHONY: vet

## Run full linting
lint-full:
	./scripts/linting/run-linters.sh
.PHONY: lint-full

## Run linting for build pipeline
lint-pipeline:
	./scripts/linting/golangci-pipeline.sh
.PHONY: lint-pipeline

## Run linting for sonar report
lint-sonar:
	./scripts/linting/golangci-sonar.sh
.PHONY: lint-sonar

## recreate all generated code and documentation.
codegen:
	./scripts/codegen/go-generate.sh
.PHONY: codegen

## recreate all generated code and swagger documentation and format code.
generate: codegen format-project vet
.PHONY: generate

## Release
release:
	./scripts/release/release.sh
.PHONY: release

## Release local snapshot
release-local-snapshot:
	./scripts/release/local-snapshot-release.sh
.PHONY: release-local-snapshot

## Check goreleaser config.
check-releaser:
	./scripts/release/check.sh
.PHONY: check-releaser

## Issue new release.
new-version: vet test-regression build
	./scripts/release/new-version.sh
.PHONY: new-release

bump-go-version:
	./scripts/bump-go.sh $(GOVERSION)
.PHONY: bump-go-version

## Build and start the Telegram bot locally via docker compose.
bot-up:
	$(COMPOSE_BOT_CMD_BASE) --env-file $(BOT_ENV_FILE) up --build -d
.PHONY: bot-up

## Stop the Telegram bot docker compose deployment.
bot-down:
	$(COMPOSE_BOT_CMD_BASE) down
.PHONY: bot-down

## Follow logs of the Telegram bot container.
bot-logs:
	$(COMPOSE_BOT_CMD_BASE) logs -f bot
.PHONY: bot-logs

## Install pinned Spec Kit, its extensions and agent skills.
speckit-install:
	$(SPECKIT_ENV) ./scripts/speckit/install.sh
.PHONY: speckit-install

## Check Spec Kit and extension updates without installing them.
speckit-update-check:
	$(SPECKIT_ENV) python3 ./scripts/speckit/check-updates.py
.PHONY: speckit-update-check

## Run network-free Spec Kit update-checker tests.
speckit-update-test:
	python3 ./scripts/speckit/test_check_updates.py
.PHONY: speckit-update-test

.DEFAULT_GOAL := help

