
# Version variables
VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || echo "dev-$(shell git rev-parse --short HEAD)")
GIT_COMMIT ?= $(shell git rev-parse HEAD)
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build flags
LDFLAGS := -X 'github.com/spaulg/solo/internal/pkg/impl/common/domain/version.Version=$(VERSION)' \
           -X 'github.com/spaulg/solo/internal/pkg/impl/common/domain/version.GitCommit=$(GIT_COMMIT)' \
           -X 'github.com/spaulg/solo/internal/pkg/impl/common/domain/version.BuildDate=$(BUILD_DATE)'

GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOTESTSUM := gotestsum
GOCOVER := $(GOCMD) tool cover
GOLINT := golangci-lint

ROOT_DIR := $(shell pwd)
BUILD_ROOT := $(ROOT_DIR)/.build
BUILD_OUTPUT_DIR := $(BUILD_ROOT)/output
TEST_OUTPUT_DIR := $(BUILD_ROOT)/tests
SRC_DIR := $(ROOT_DIR)
PREFIX ?= /usr/local
BINDIR := $(PREFIX)/bin
IMPLDIR := ./internal/pkg/impl
TEST_FLAGS ?= $(IMPLDIR)/...

NATIVE_SERVICES := solo
LINUX_SERVICES := solo-entrypoint
SERVICES := $(NATIVE_SERVICES) $(LINUX_SERVICES)

GOOS_solo :=
GOOS_solo-entrypoint := linux

.PHONY: all build test install clean
all: build

protos:
	mkdir -p $(BUILD_OUTPUT_DIR)
	find $(SRC_DIR)/internal/pkg/impl/common/infra/grpc/services -name *.proto -exec \
		protoc --experimental_allow_proto3_optional --proto_path=$(SRC_DIR) --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative {} \;

bootstrap: protos

$(NATIVE_SERVICES): bootstrap
	cd $(SRC_DIR)/cmd/$@ && CGO_ENABLED=0 $(GOBUILD) -ldflags="$(LDFLAGS) -s -w" -o $(BUILD_OUTPUT_DIR)/$@

$(LINUX_SERVICES): bootstrap
	cd $(SRC_DIR)/cmd/$@ && GOOS=linux CGO_ENABLED=0 $(GOBUILD) -ldflags="$(LDFLAGS) -s -w" -o $(BUILD_OUTPUT_DIR)/$@

build: bootstrap $(NATIVE_SERVICES) $(LINUX_SERVICES) ## Build files

test:
	mkdir -p $(TEST_OUTPUT_DIR)
	@cd $(SRC_DIR) && $(GOTESTSUM) --format pkgname -- \
		$(TEST_FLAGS) \
		-coverprofile=$(TEST_OUTPUT_DIR)/coverage.txt

citest:
	mkdir -p $(TEST_OUTPUT_DIR)
	@cd $(SRC_DIR) && $(GOTEST) \
		$(TEST_FLAGS) \
		-coverprofile=$(TEST_OUTPUT_DIR)/coverage.txt \
		-json \
		| tee $(TEST_OUTPUT_DIR)/test.json \
		| go-junit-report -set-exit-code > $(TEST_OUTPUT_DIR)/junit.xml

cover: ## Open coverage report for the last test run
	cd $(SRC_DIR) && $(GOCOVER) -html=$(TEST_OUTPUT_DIR)/coverage.txt

lint: ## Run linters
	$(GOLINT) run

install: ## Install files to the system
	mkdir -p $(BINDIR)
	$(foreach srv, $(NATIVE_SERVICES), install -m 0755 $(BUILD_OUTPUT_DIR)/$(srv) $(BINDIR)/$(srv) || exit;)
	$(foreach srv, $(LINUX_SERVICES), install -m 0755 $(BUILD_OUTPUT_DIR)/$(srv) $(BINDIR)/$(srv) || exit;)

clean: ## Clean build artifacts
	rm -rf $(BUILD_ROOT)

help:
	@echo "Usage: make [target]"
	@echo
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "}; /^[a-zA-Z0-9_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
