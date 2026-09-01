#!/bin/sh
# Render package metadata at a fixture major tag and inspect its public contract.
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
renderer="$script_dir/render-packaging.sh"
fixture_root=$(mktemp -d)
cleanup() {
  chmod -R u+w "$fixture_root" 2>/dev/null || true
  rm -rf "$fixture_root"
}
trap cleanup EXIT HUP INT TERM

# git -C wants a native path, and an MSYS mktemp path is not one, so each call
# changes directory in a subshell instead.
fixture_git() {
  (cd "$fixture_root" && git "$@")
}

fixture_git init -q
fixture_git config user.name Fixture
fixture_git config user.email fixture@example.test
fixture_git config commit.gpgSign false
mkdir -p "$fixture_root/dist"
printf 'fixture\n' >"$fixture_root/release-input"
fixture_git add release-input
fixture_git commit -q -m release
fixture_git -c tag.gpgSign=false tag v2.0.0

for artifact in \
  agent-compose-darwin-arm64 \
  agent-compose-linux-amd64 \
  agent-compose-linux-arm64 \
  agent-compose-windows-amd64.exe \
  agent-compose-roster.tar.gz
do
  printf '%s\n' "$artifact" >"$fixture_root/dist/$artifact"
done

(
  cd "$fixture_root"
  sh "$renderer"
)

formula="$fixture_root/dist/agent-compose.rb"
manifest="$fixture_root/dist/agent-compose.json"
for expected in \
  'Core Roster context composition for native agent harnesses' \
  'v2.0.0/agent-compose' \
  '2.0.0'
do
  if ! grep -F "$expected" "$formula" "$manifest" >/dev/null; then
    echo "render-packaging-test: missing $expected" >&2
    exit 1
  fi
done
if ! grep -F '"agent-compose-windows-amd64.exe", "acompose", "compose"' "$manifest" >/dev/null; then
  echo "render-packaging-test: Scoop alias contract is missing" >&2
  exit 1
fi
for expected in 'agent-compose-roster.tar.gz' 'share/"agent-compose"'
do
  if ! grep -F "$expected" "$formula" "$manifest" >/dev/null; then
    echo "render-packaging-test: seed roster is not installed: $expected" >&2
    exit 1
  fi
done

echo "render-packaging-test: ok"
