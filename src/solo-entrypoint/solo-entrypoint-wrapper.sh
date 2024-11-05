#!/bin/sh
set -e

cd /solo/agent
go run main.go "$@"
