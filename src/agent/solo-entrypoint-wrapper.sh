#!/bin/sh
set -e

cd /solo/agent
exec go run main.go "$@"
