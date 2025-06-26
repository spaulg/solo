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

FIND_IMPL_PACKAGES := find $(IMPLDIR) -name "*.go" | grep -vE ".*\.pb\.go" | grep -v ".*_testsuite\.go" | \
	xargs -n1 dirname | sort -u | paste -sd, -

NATIVE_SERVICES := solo
LINUX_SERVICES := solo-entrypoint
SERVICES := $(NATIVE_SERVICES) $(LINUX_SERVICES)

GOOS_solo :=
GOOS_solo-entrypoint := linux

.PHONY: all build test install clean
all: build

build: protos $(NATIVE_SERVICES) $(LINUX_SERVICES)

protos:
	mkdir -p $(BUILD_DIR)
	find $(SRC_DIR)/internal/pkg/impl/common/grpc/services -name *.proto -exec \
		protoc --experimental_allow_proto3_optional --proto_path=$(SRC_DIR) --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative {} \;

$(NATIVE_SERVICES): protos
	cd $(SRC_DIR)/cmd/$@ && CGO_ENABLED=0 $(GOBUILD) -ldflags="-s -w" -o $(BUILD_DIR)/$@

$(LINUX_SERVICES): protos
	cd $(SRC_DIR)/cmd/$@ && GOOS=linux CGO_ENABLED=0 $(GOBUILD) -ldflags="-s -w" -o $(BUILD_DIR)/$@

test:
	@cd $(SRC_DIR) && $(GOTEST) \
		-coverprofile=coverage.out \
		-coverpkg=$(shell $(FIND_IMPL_PACKAGES)) \
		$(IMPLDIR)/... | sed -E 's/of statements in .*/of statements/' && \
			cat coverage.out | \
			grep -vE ".*\.pb\.go" | \
			grep -v ".*_testsuite\.go" > filtered.coverage.out && \
	go tool cover -func=filtered.coverage.out | tail -1 | awk '{print "Total:", $$3}'

lint:
	$(GOLINT) run

cover:
	cd $(SRC_DIR) && $(GOCOVER) -html=filtered.coverage.out

install:
	mkdir -p $(BINDIR)
	$(foreach srv, $(NATIVE_SERVICES), install -m 0755 $(BUILD_DIR)/$(srv) $(BINDIR)/$(srv) || exit;)
	$(foreach srv, $(LINUX_SERVICES), install -m 0755 $(BUILD_DIR)/$(srv) $(BINDIR)/$(srv) || exit;)

clean:
	rm -rf $(BUILD_DIR)
