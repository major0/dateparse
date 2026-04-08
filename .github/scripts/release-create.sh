#!/usr/bin/env bash
set -euo pipefail

# Creates a GitHub release with auto-generated changelog.
#
# Requires: TAG, NAME, PREV, GH_TOKEN (env)

TAG="$1"
NAME="$2"
PREV="$3"

ARGS=(
  "$TAG"
  --title "${NAME} ${TAG}"
  --generate-notes
)

# If we found a previous tag, scope the changelog to that range
if [ -n "$PREV" ]; then
  ARGS+=(--notes-start-tag "$PREV")
fi

# Mark pre-release for rc tags
case "$TAG" in
  *-rc*) ARGS+=(--prerelease) ;;
esac

gh release create "${ARGS[@]}"
