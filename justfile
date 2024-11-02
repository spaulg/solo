GOCMD := "go"
GOBUILD := GOCMD + " build"
GOTEST := GOCMD + " test"
GOCOVER := GOCMD + " tool cover"

CLI_BINARY_NAME := "solo"
ENTRYPOINT_BINARY_NAME := "solo-entrypoint"
BUILD_DIR := "build"
SRC := "./..."
COMPONENTS := "cli agent shared"

build:
    find src/shared/pkg/solo/grpc/services -name *.proto -exec protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative {} \;
    mkdir -p {{BUILD_DIR}}
    CGO_ENABLED=0 {{GOBUILD}} -C src/cli -ldflags="-s -w" -o ../../{{BUILD_DIR}}/{{CLI_BINARY_NAME}} main.go
    GOOS=linux CGO_ENABLED=0 {{GOBUILD}} -C src/agent -ldflags="-s -w" -o ../../{{BUILD_DIR}}/{{ENTRYPOINT_BINARY_NAME}} main.go

test PACKAGE="":
    #!/usr/bin/env sh
    if [ -n "{{PACKAGE}}" ]; then
        cd `dirname "{{PACKAGE}}"`
        {{GOTEST}} -v ./...
    else
        for COMPONENT in {{COMPONENTS}}; do
            {{GOTEST}} -C "src/$COMPONENT" -coverprofile=coverage.out -v ./...
        done
    fi

test-coverage PATH:
    cd "{{PATH}}"; {{GOCOVER}} -html=coverage.out

install:
    install -m 0755 -o root -g admin {{BUILD_DIR}}/{{CLI_BINARY_NAME}} /usr/local/bin/
    install -m 0755 -o root -g admin {{BUILD_DIR}}/{{ENTRYPOINT_BINARY_NAME}} /usr/local/bin/

clean:
    #!/usr/bin/env sh
    rm -rf {{BUILD_DIR}}

    for COMPONENT in {{COMPONENTS}}; do
        rm -f "src/$COMPONENT/coverage.out"
    done
