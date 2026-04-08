#!/usr/bin/env bash
set -euo pipefail

# Finds the previous tag in the same series for changelog generation.
#
# RC tags: find the nearest ancestor tag.
# Stable tags: skip past RC tags to find the previous stable release.
#
# Args: TAG
# Outputs: tag

TAG="$1"

case "$TAG" in
  *-rc*)
    PREV=$(git describe --tags --abbrev=0 --match "v*" \
             "${TAG}^" 2>/dev/null || true)
    ;;
  *)
    PREV=""
    COMMIT="${TAG}^"
    for _ in $(seq 1 50); do
      CANDIDATE=$(git describe --tags --abbrev=0 --match "v*" \
                    "$COMMIT" 2>/dev/null || true)
      [ -z "$CANDIDATE" ] && break
      case "$CANDIDATE" in
        *-rc*) COMMIT="${CANDIDATE}^" ;;
        *)     PREV="$CANDIDATE"; break ;;
      esac
    done
    ;;
esac

if [ -z "$PREV" ]; then
  echo "::notice::No previous tag found; changelog will cover full history"
else
  echo "::notice::Previous tag: $PREV"
fi
echo "tag=$PREV" >> "$GITHUB_OUTPUT"
