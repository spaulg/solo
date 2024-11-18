GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCOVER=$(GOCMD) tool cover
GOLINT=golangci-lint

ROOT_DIR := $(shell pwd)
BUILD_DIR=$(ROOT_DIR)/build
SRC_DIR=$(ROOT_DIR)

NATIVE_SERVICES := solo
LINUX_SERVICES := solo-entrypoint
SERVICES := $(NATIVE_SERVICES) $(LINUX_SERVICES)

GOOS_solo :=
GOOS_solo-entrypoint := linux

.PHONY: all build test install clean
all: build

build: shared $(NATIVE_SERVICES) $(LINUX_SERVICES)

shared:
	mkdir -p $(BUILD_DIR)
	find $(SRC_DIR)/internal/pkg/shared/grpc/services -name *.proto -exec \
		protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative {} \;

$(NATIVE_SERVICES):
	cd $(SRC_DIR)/cmd/$@ && CGO_ENABLED=0 $(GOBUILD) -ldflags="-s -w" -o $(BUILD_DIR)/$@

$(LINUX_SERVICES):
	cd $(SRC_DIR)/cmd/$@ && GOOS=linux CGO_ENABLED=0 $(GOBUILD) -ldflags="-s -w" -o $(BUILD_DIR)/$@

test:
	cd $(SRC_DIR) && $(GOTEST) -coverprofile=coverage.out ./...

lint:
	$(foreach srv, $(SERVICES), cd $(SRC_DIR)/$(srv) && $(GOLINT) run || exit;)

cover:
	cd $(SRC_DIR) && $(GOCOVER) -html=coverage.out

install:
	$(foreach srv, $(NATIVE_SERVICES), install -m 0755 -o root -g admin $(BUILD_DIR)/$(srv) /usr/local/bin/ || exit;)
	$(foreach srv, $(LINUX_SERVICES), install -m 0755 -o root -g admin $(BUILD_DIR)/$(srv) /usr/local/bin/ || exit;)

clean:
	rm -rf $(BUILD_DIR)
