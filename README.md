# Solo

![Project status](https://img.shields.io/badge/project_status-alpha-orange)
![Latest Version](https://img.shields.io/github/v/tag/spaulg/solo?label=latest%20version&color=green)
![Codecov](https://img.shields.io/codecov/c/github/spaulg/solo)

Solo simplifies the creation of containerized development environments by wrapping Docker Compose and executing 
workflow commands for events like starting, stopping, and rebuilding containers, plus tooling support.

Inspired by [lando.dev](https://lando.dev).

## Installation

### Using Homebrew (MacOS or Linux)

Install using my [Homebrew tap](https://github.com/spaulg/homebrew-tap) with the command:

```shell
brew install spaulg/tap/solo
```

## Development Setup

### Using Homebrew (MacOS or Linux)

```shell
brew install go protobuf protoc-gen-go protoc-gen-go-grpc
```

### Debian Linux

```shell
apt update && apt install unzip git make curl
```

Install Go (see https://go.dev/doc/install)

Install protoc (see https://protobuf.dev/installation/)

```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.10.1
export PATH="$PATH:$HOME/go/bin"
```

## Make commands

### Building

`make` or `make build`

### Testing/Linting/Code Coverage

`make test`

`make lint`

`make cover`

### Installation

`make install`

### Build cleanup

`make clean`
