## Prerequisites

* Go
* Docker compose API v3 compatible container provisioner

## Development Setup

Install protoc [for your system](https://grpc.io/docs/protoc-installation/)

Install protobuf grpc Go plugins  

`go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`

`go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`

## Building

`make`

## Testing/Linting/Code Coverage

`make test`

`make lint`

`make cover`

## Installation

`make install`

## Build cleanup

`make clean`
