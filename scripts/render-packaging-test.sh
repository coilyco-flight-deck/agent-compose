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

git -C "$fixture_root" init -q
git -C "$fixture_root" config user.name Fixture
git -C "$fixture_root" config user.email fixture@example.test
git -C "$fixture_root" config commit.gpgSign false
mkdir -p "$fixture_root/dist"
printf 'fixture\n' >"$fixture_root/release-input"
git -C "$fixture_root" add release-input
git -C "$fixture_root" commit -q -m release
git -C "$fixture_root" -c tag.gpgSign=false tag v2.0.0

for artifact in \
  agent-compose-darwin-arm64 \
  agent-compose-linux-amd64 \
  agent-compose-linux-arm64 \
  agent-compose-windows-amd64.exe
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

echo "render-packaging-test: ok"
