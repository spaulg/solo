GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCOVER=$(GOCMD) tool cover
GOLINT=golangci-lint

ROOT_DIR := $(shell pwd)
BUILD_DIR=$(ROOT_DIR)/.build
SRC_DIR=$(ROOT_DIR)
PREFIX ?= /usr/local
BINDIR = $(PREFIX)/bin

NATIVE_SERVICES := solo
LINUX_SERVICES := solo-entrypoint
SERVICES := $(NATIVE_SERVICES) $(LINUX_SERVICES)

ifeq ($(HOMEBREW_BUILD),1)
INSTALL_OWNER =
else
INSTALL_OWNER = -o root
endif

GOOS_solo :=
GOOS_solo-entrypoint := linux

.PHONY: all build test install clean
all: build

build: protos $(NATIVE_SERVICES) $(LINUX_SERVICES)

protos:
	mkdir -p $(BUILD_DIR)
	find $(SRC_DIR)/internal/pkg/common/grpc/services -name *.proto -exec \
		protoc --experimental_allow_proto3_optional --proto_path=$(SRC_DIR) --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative {} \;

$(NATIVE_SERVICES): protos
	cd $(SRC_DIR)/cmd/$@ && CGO_ENABLED=0 $(GOBUILD) -ldflags="-s -w" -o $(BUILD_DIR)/$@

$(LINUX_SERVICES): protos
	cd $(SRC_DIR)/cmd/$@ && GOOS=linux CGO_ENABLED=0 $(GOBUILD) -ldflags="-s -w" -o $(BUILD_DIR)/$@

test:
	cd $(SRC_DIR) && $(GOTEST) -coverprofile=coverage.out ./...

lint:
	$(GOLINT) run

cover:
	cd $(SRC_DIR) && $(GOCOVER) -html=coverage.out

install:
	$(foreach srv, $(NATIVE_SERVICES), install -m 0755 $(INSTALL_OWNER) $(BUILD_DIR)/$(srv) $(BINDIR) || exit;)
	$(foreach srv, $(LINUX_SERVICES), install -m 0755 $(INSTALL_OWNER) $(BUILD_DIR)/$(srv) $(BINDIR) || exit;)

clean:
	rm -rf $(BUILD_DIR)
