#!/usr/bin/env bash
set -euo pipefail

# Extracts coverage percentage from a coverage.out file.
# Outputs the numeric percentage (no % sign).
#
# Args: COVERAGE_FILE

COVERAGE_FILE="$1"

go tool cover -func="$COVERAGE_FILE" | grep "total:" | awk '{print $3}' | sed 's/%//'
