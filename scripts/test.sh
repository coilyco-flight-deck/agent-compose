#!/bin/sh
# Full repository validation: engine tests first, then the hook catalog.
set -e
test_home=$(mktemp -d)
cleanup() {
  chmod -R u+w "$test_home" 2>/dev/null || true
  rm -rf "$test_home"
}
trap cleanup EXIT HUP INT TERM

go test ./...
sh scripts/release-impact-test.sh
sh scripts/render-packaging-test.sh
env HOME="$test_home" sh scripts/palette-web.sh test
env HOME="$test_home" sh scripts/context-budget.sh
sh scripts/evalkit-check.sh
pre-commit run --all-files
