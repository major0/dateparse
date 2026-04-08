#!/usr/bin/env bash
set -euo pipefail

# Ensures the pushed tag is annotated. If it's lightweight, replaces it
# with an annotated tag pointing at the same commit.
#
# Args: TAG
# Requires: GH_TOKEN or git push credentials

TAG="$1"
TYPE=$(git cat-file -t "$TAG")

if [ "$TYPE" = "tag" ]; then
  echo "::notice::$TAG is already annotated"
  exit 0
fi

echo "::warning::$TAG is a lightweight tag — converting to annotated"

COMMIT=$(git rev-parse "$TAG^{commit}")

# Delete the lightweight tag locally and remotely
git tag -d "$TAG"
git push origin ":refs/tags/$TAG"

# Re-create as annotated and push
git tag -a "$TAG" "$COMMIT" -m "$TAG"
git push origin "$TAG"

echo "::notice::$TAG converted to annotated tag on $COMMIT"
