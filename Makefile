
# Version variables
VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || echo "dev-$(shell git rev-parse --short HEAD)")
GIT_COMMIT ?= $(shell git rev-parse HEAD)
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build flags
LDFLAGS := -X 'github.com/spaulg/solo/internal/pkg/impl/common/version.Version=$(VERSION)' \
           -X 'github.com/spaulg/solo/internal/pkg/impl/common/version.GitCommit=$(GIT_COMMIT)' \
           -X 'github.com/spaulg/solo/internal/pkg/impl/common/version.BuildDate=$(BUILD_DATE)'

GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOCOVER := $(GOCMD) tool cover
GOLINT := golangci-lint

ROOT_DIR := $(shell pwd)
BUILD_DIR := $(ROOT_DIR)/.build
SRC_DIR := $(ROOT_DIR)
PREFIX ?= /usr/local
BINDIR := $(PREFIX)/bin
IMPLDIR := ./internal/pkg/impl
TEST_FLAGS ?= $(IMPLDIR)/...

FIND_IMPL_PACKAGES := find $(IMPLDIR) -name "*.go" | grep -vE ".*\.pb\.go" | grep -v ".*_testsuite\.go" | \
	xargs -n1 dirname | sort -u | paste -sd, -

NATIVE_SERVICES := solo
LINUX_SERVICES := solo-entrypoint
SERVICES := $(NATIVE_SERVICES) $(LINUX_SERVICES)

GOOS_solo :=
GOOS_solo-entrypoint := linux

.PHONY: all build test install clean
all: build

protos:
	mkdir -p $(BUILD_DIR)
	find $(SRC_DIR)/internal/pkg/impl/common/grpc/services -name *.proto -exec \
		protoc --experimental_allow_proto3_optional --proto_path=$(SRC_DIR) --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative {} \;

bootstrap: protos

$(NATIVE_SERVICES): bootstrap
	cd $(SRC_DIR)/cmd/$@ && CGO_ENABLED=0 $(GOBUILD) -ldflags="$(LDFLAGS) -s -w" -o $(BUILD_DIR)/$@

$(LINUX_SERVICES): bootstrap
	cd $(SRC_DIR)/cmd/$@ && GOOS=linux CGO_ENABLED=0 $(GOBUILD) -ldflags="$(LDFLAGS) -s -w" -o $(BUILD_DIR)/$@

build: bootstrap $(NATIVE_SERVICES) $(LINUX_SERVICES) ## Build files

test: ## Run tests; pass the flag TEST_FLAGS="flags for go test" to override default test flags
	@cd $(SRC_DIR) && $(GOTEST) \
		$(TEST_FLAGS) \
		-v \
		-coverprofile=coverage.txt \
		-coverpkg=$(shell $(FIND_IMPL_PACKAGES))

cover: ## Open coverage report for the last test run
	cd $(SRC_DIR) && $(GOCOVER) -html=coverage.txt

lint: ## Run linters
	$(GOLINT) run

install: ## Install files to the system
	mkdir -p $(BINDIR)
	$(foreach srv, $(NATIVE_SERVICES), install -m 0755 $(BUILD_DIR)/$(srv) $(BINDIR)/$(srv) || exit;)
	$(foreach srv, $(LINUX_SERVICES), install -m 0755 $(BUILD_DIR)/$(srv) $(BINDIR)/$(srv) || exit;)

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)

help:
	@echo "Usage: make [target]"
	@echo
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "}; /^[a-zA-Z0-9_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
