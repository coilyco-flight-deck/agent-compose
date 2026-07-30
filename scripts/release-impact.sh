#!/bin/sh
# Decide whether one validated revision should publish a product release.
set -eu

event="${RELEASE_EVENT:-push}"
base="${RELEASE_BASE:-}"
head="${RELEASE_HEAD:-HEAD}"

if [ "$event" = "workflow_dispatch" ]; then
  echo true
  exit 0
fi

if [ -f .release-major ]; then
  echo "release-impact: automatic publication held by .release-major" >&2
  echo false
  exit 0
fi

case "$base" in
  ""|0000000000000000000000000000000000000000)
    echo true
    exit 0
    ;;
esac

if ! git cat-file -e "${base}^{commit}" 2>/dev/null; then
  echo "release-impact: base revision is unavailable, publishing fail closed" >&2
  echo true
  exit 0
fi

if git diff --quiet "$base" "$head" -- \
  cmd \
  internal \
  go.mod \
  go.sum \
  scripts/release-build.sh \
  scripts/render-packaging.sh
then
  echo false
else
  status=$?
  if [ "$status" -ne 1 ]; then
    exit "$status"
  fi
  echo true
fi
