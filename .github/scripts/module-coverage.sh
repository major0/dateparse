#!/usr/bin/env bash
set -euo pipefail

# Runs tests with coverage and generates HTML report.
#
# Args: [MODULE_DIR]
# Defaults to current directory if not specified.

MODULE_DIR="${1:-.}"

(
	cd "$MODULE_DIR"
	go mod tidy
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
)
