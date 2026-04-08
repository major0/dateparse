#!/usr/bin/env bash
set -euo pipefail

# Checks for coverage regression between PR and main branch.
# Expects PR_COVERAGE and MAIN_COVERAGE environment variables.
# Fails if coverage drops by more than 1%.

if [ "${SKIP_REGRESSION:-}" = "true" ]; then
	echo "Skipping regression check due to main branch build failures"
	echo "PR branch coverage: ${PR_COVERAGE}%"
	exit 0
fi

echo "Main: ${MAIN_COVERAGE}%, PR: ${PR_COVERAGE}%"
COVERAGE_DIFF=$(echo "$MAIN_COVERAGE - $PR_COVERAGE" | bc -l)

if (( $(echo "$COVERAGE_DIFF > 1.0" | bc -l) )); then
	echo "Coverage regression: ${COVERAGE_DIFF}%"
	exit 1
fi
