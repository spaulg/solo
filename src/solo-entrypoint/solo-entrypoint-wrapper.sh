#!/bin/sh
set -e

cd /solo/solo-entrypoint
go run main.go "$@"
