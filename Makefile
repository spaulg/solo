GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCOVER=$(GOCMD) tool cover

CLI_BINARY_NAME=solo
ENTRYPOINT_BINARY_NAME=solo-entrypoint
BUILD_DIR=build
SRC=./...
CLI_GO_FILES=$(shell find cli -name '*.go')
ENTRYPOINT_GO_FILES=$(shell find agent -name '*.go')

.PHONY: all build test-cli test-agent cover-cli cover-agent clean -build-solo -build-entrypoint
all: build

build: -build-solo -build-entrypoint

-build-solo: $(CLI_GO_FILES)
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 $(GOBUILD) -C cli -ldflags="-s -w" -o ../$(BUILD_DIR)/$(CLI_BINARY_NAME) main.go

-build-entrypoint: $(ENTRYPOINT_GO_FILES)
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux CGO_ENABLED=0 $(GOBUILD) -C agent -ldflags="-s -w" -o ../$(BUILD_DIR)/$(ENTRYPOINT_BINARY_NAME) main.go

test-cli:
	$(GOTEST) -C cli -coverprofile=coverage.out -v ./...

cover-cli:
	cd cli; $(GOCOVER) -html=coverage.out

test-agent:
	$(GOTEST) -C agent -coverprofile=coverage.out -v ./...

cover-agent:
	cd agent; $(GOCOVER) -html=coverage.out

clean:
	@rm -rf $(BUILD_DIR)
