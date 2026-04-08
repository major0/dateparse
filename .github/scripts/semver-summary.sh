#!/usr/bin/env bash
set -euo pipefail

# Writes a semver summary to GITHUB_STEP_SUMMARY.
#
# Args: TAG TAG_TYPE

TAG="$1"
TAG_TYPE="$2"

{
	echo "### Semver Summary"
	echo "- **Tag**: ${TAG}"
	echo "- **Type**: ${TAG_TYPE}"
} >> "$GITHUB_STEP_SUMMARY"
