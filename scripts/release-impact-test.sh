#!/bin/sh
# Fixture coverage for the owning release-impact classifier.
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
classifier="$script_dir/release-impact.sh"
fixture_root=$(mktemp -d)
cleanup() {
  chmod -R u+w "$fixture_root" 2>/dev/null || true
  rm -rf "$fixture_root"
}
trap cleanup EXIT HUP INT TERM

git -C "$fixture_root" init -q
git -C "$fixture_root" config user.name Fixture
git -C "$fixture_root" config user.email fixture@example.test
git -C "$fixture_root" config commit.gpgSign false
git -C "$fixture_root" config tag.gpgSign false
mkdir -p "$fixture_root/cmd/tool" "$fixture_root/docs" "$fixture_root/evaluations/latest"
printf 'package main\n' >"$fixture_root/cmd/tool/main.go"
printf '# Guide\n' >"$fixture_root/docs/guide.md"
git -C "$fixture_root" add .
git -C "$fixture_root" commit -q -m base
base=$(git -C "$fixture_root" rev-parse HEAD)
git -C "$fixture_root" tag v1.0.0 "$base"

expect() {
  expected=$1
  event_name=$2
  before=$3
  after=$4
  actual=$(
    cd "$fixture_root"
    RELEASE_EVENT="$event_name" RELEASE_BASE="$before" RELEASE_HEAD="$after" \
      "$classifier"
  )
  if [ "$actual" != "$expected" ]; then
    echo "release-impact-test: got $actual, want $expected" >&2
    exit 1
  fi
}

printf 'More docs.\n' >>"$fixture_root/docs/guide.md"
git -C "$fixture_root" add docs/guide.md
git -C "$fixture_root" commit -q -m docs
docs_revision=$(git -C "$fixture_root" rev-parse HEAD)
expect false push "$base" "$docs_revision"

printf 'format: fixture\n' >"$fixture_root/evaluations/latest/result.yaml"
git -C "$fixture_root" add evaluations/latest/result.yaml
git -C "$fixture_root" commit -q -m results
results_revision=$(git -C "$fixture_root" rev-parse HEAD)
expect false push "$docs_revision" "$results_revision"

printf '// product change\n' >>"$fixture_root/cmd/tool/main.go"
git -C "$fixture_root" add cmd/tool/main.go
git -C "$fixture_root" commit -q -m product
product_revision=$(git -C "$fixture_root" rev-parse HEAD)
expect true push "$results_revision" "$product_revision"
expect true push 0000000000000000000000000000000000000000 "$product_revision"

printf 'format: recovery\n' >"$fixture_root/evaluations/latest/result.yaml"
git -C "$fixture_root" add evaluations/latest/result.yaml
git -C "$fixture_root" commit -q -m recovery-results
recovery_revision=$(git -C "$fixture_root" rev-parse HEAD)
expect true push "$product_revision" "$recovery_revision"
git -C "$fixture_root" tag v1.1.0 "$recovery_revision"

printf 'Released docs.\n' >>"$fixture_root/docs/guide.md"
git -C "$fixture_root" add docs/guide.md
git -C "$fixture_root" commit -q -m released-docs
released_docs_revision=$(git -C "$fixture_root" rev-parse HEAD)
expect false push "$recovery_revision" "$released_docs_revision"

printf 'v2.0.0\n' >"$fixture_root/.release-major"
printf '// held product change\n' >>"$fixture_root/cmd/tool/main.go"
git -C "$fixture_root" add .release-major cmd/tool/main.go
git -C "$fixture_root" commit -q -m held
held_revision=$(git -C "$fixture_root" rev-parse HEAD)
expect false push "$released_docs_revision" "$held_revision"
expect true workflow_dispatch "$held_revision" "$held_revision"

echo "release-impact-test: ok"
