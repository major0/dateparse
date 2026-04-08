#!/usr/bin/env bash
set -euo pipefail

# Installs Go development tools (goimports, golangci-lint).
# Appends GOPATH/bin to GITHUB_PATH if running in CI.

go install golang.org/x/tools/cmd/goimports@latest
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" latest

if [ -n "${GITHUB_PATH:-}" ]; then
	echo "$(go env GOPATH)/bin" >> "$GITHUB_PATH"
fi
