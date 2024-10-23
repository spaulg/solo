GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCOVER=$(GOCMD) tool cover

CLI_BINARY_NAME=solo
ENTRYPOINT_BINARY_NAME=solo-entrypoint
BUILD_DIR=build
SRC=./...
CLI_GO_FILES=$(shell find src/cli -name '*.go')
ENTRYPOINT_GO_FILES=$(shell find src/agent -name '*.go')
SHARED_PROTO_FILES=$(shell find src/shared -name '*.proto')

.PHONY: all build test-cli test-agent cover-cli cover-agent clean -build-shared -build-solo -build-entrypoint
all: build

build: -build-solo -build-entrypoint

-build-shared: $(SHARED_PROTO_FILES)
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative \
		src/shared/pkg/solo/grpc/services/workflow.proto

-build-solo: -build-shared $(CLI_GO_FILES)
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) -C src/cli -ldflags="-s -w" -o ../../$(BUILD_DIR)/$(CLI_BINARY_NAME) main.go

-build-entrypoint: -build-shared $(ENTRYPOINT_GO_FILES)
	mkdir -p $(BUILD_DIR)
	GOOS=linux CGO_ENABLED=0 $(GOBUILD) -C src/agent -ldflags="-s -w" -o ../../$(BUILD_DIR)/$(ENTRYPOINT_BINARY_NAME) main.go

test-cli:
	$(GOTEST) -C src/cli -coverprofile=coverage.out -v ./...

cover-cli:
	cd src/cli; $(GOCOVER) -html=coverage.out

test-agent:
	$(GOTEST) -C src/agent -coverprofile=coverage.out -v ./...

cover-agent:
	cd src/agent; $(GOCOVER) -html=coverage.out

install:
	install -m 0755 -o root -g admin $(BUILD_DIR)/$(CLI_BINARY_NAME) /usr/local/bin/
	install -m 0755 -o root -g admin $(BUILD_DIR)/$(ENTRYPOINT_BINARY_NAME) /usr/local/bin/

clean:
	rm -rf $(BUILD_DIR)
