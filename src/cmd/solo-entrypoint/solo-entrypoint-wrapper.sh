#!/bin/sh
set -e

cd /solo/cmd/solo-entrypoint
go run main.go "$@"
