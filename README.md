# Development Setup

### Install the Protobuf compiler

Linux with apt: `apt install -y protobuf-compiler`

MacOS with brew: `brew install protobuf`

### Install GRPC/protobuf compiler plugins

`go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`

`go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`

# Make commands

### Building

`make`

### Testing/Linting/Code Coverage

`make test`

`make lint`

`make cover`

### Installation

`make install`

### Build cleanup

`make clean`
