#!/bin/sh
# Exercise the shipped acompose entry point without touching the operator's home.
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
smoke_root=$(mktemp -d "${TMPDIR:-/tmp}/agent-compose-smoke.XXXXXX")

cleanup() {
  rm -rf -- "$smoke_root"
}
trap cleanup EXIT HUP INT TERM

fail() {
  printf 'smoke: %s\n' "$1" >&2
  exit 1
}

assert_file() {
  test -f "$1" || fail "missing file $1"
}

assert_contains() {
  grep -F "$2" "$1" >/dev/null || fail "$1 does not contain: $2"
}

snapshot_file() {
  cp "$1" "$snapshot_dir/$2"
}

assert_unchanged() {
  cmp -s "$1" "$snapshot_dir/$2" || fail "$1 changed during the second convergence"
}

goos=$(go env GOOS)
case "$goos" in
  windows)
    command -v cygpath >/dev/null 2>&1 || fail "cygpath is required by the Windows sh environment"
    native_root=$(cygpath -m "$smoke_root")
    binary="$smoke_root/bin/acompose.exe"
    ;;
  *)
    native_root=$(CDPATH= cd -- "$smoke_root" && pwd -P)
    binary="$smoke_root/bin/acompose"
    ;;
esac

fixture_home="$smoke_root/home"
state_dir="$fixture_home/.agent-compose"
projects_dir="$smoke_root/projects"
provider_dir="$projects_dir/coilyco-flight-deck/agentic-os"
skill_dir="$provider_dir/.agents/skills/coding-go"
load_points="$smoke_root/load-points"
fixtures="$smoke_root/fixtures"
snapshot_dir="$smoke_root/snapshot"

mkdir -p "$smoke_root/bin" "$state_dir" "$skill_dir" "$load_points" "$fixtures" "$snapshot_dir"

cat >"$fixtures/AGENTS.COMPOSE.md" <<'EOF'
# Smoke doctrine

The smoke fixture exercises host convergence.
EOF

cat >"$skill_dir/SKILL.md" <<'EOF'
# Go

Use Go for the smoke fixture.
EOF

cat >"$fixtures/mcporter.json" <<'EOF'
{"imports":[],"mcpServers":{"reader":{"baseUrl":"https://mcp.example.test/mcp"}}}
EOF

cat >"$state_dir/agent-compose.yaml" <<EOF
sources:
  - $native_root/fixtures/AGENTS.COMPOSE.md
roots:
  - $native_root/home/.agent-compose/sources
roster_sources:
  - $native_root/projects/coilyco-flight-deck/agentic-os
skill_load_points:
  codex: $native_root/load-points/skills
mcp_inventory: $native_root/fixtures/mcporter.json
load_points:
  claude: $native_root/load-points/CLAUDE.md
  codex: null
EOF

cd "$repo_root"
go build -o "$binary" ./cmd/agent-compose

first_output="$smoke_root/first.txt"
if ! env HOME="$native_root/home" USERPROFILE="$native_root/home" \
  PROJECTS_ROOT="$native_root/projects" "$binary" >"$first_output" 2>&1; then
  cat "$first_output" >&2
  fail "first acompose convergence failed"
fi

assert_contains "$first_output" "roster  "
assert_contains "$first_output" "wrote"
assert_contains "$first_output" "cascade outputs=1 load-points=1 manifest=1 changed="
assert_contains "$first_output" "skills  managed=1 load-points=1"
assert_contains "$first_output" "mcp     servers=1 state=changed"

roster_table="$state_dir/sources/personality/AGENTS.COMPOSE.md"
roster_override="$state_dir/sources/personality/AGENTS.claude.md"
roster_body="$state_dir/sources/personality/personalities/curious.md"
composed="$state_dir/COMPOSED.md"
manifest="$state_dir/mount-eligibility.json"
skill_state="$state_dir/skill-mounts.json"
mcporter="$fixture_home/.mcporter/mcporter.json"
claude_mcp="$fixture_home/.claude.json"
codex_mcp="$fixture_home/.codex/config.toml"

for path in "$roster_table" "$roster_override" "$roster_body" "$composed" \
  "$manifest" "$skill_state" "$mcporter" "$claude_mcp" "$codex_mcp" \
  "$load_points/CLAUDE.md" "$load_points/skills/coding-go/SKILL.md"; do
  assert_file "$path"
done

assert_contains "$composed" "# Smoke doctrine"
assert_contains "$composed" "# Personality invariant"
assert_contains "$composed" "# Agent seats"
assert_contains "$composed" "opal engineer"
assert_contains "$mcporter" "\"reader\""
assert_contains "$claude_mcp" "\"reader\""
assert_contains "$codex_mcp" "[mcp_servers.\"reader\"]"

snapshot_file "$roster_table" roster-table
snapshot_file "$roster_override" roster-override
snapshot_file "$roster_body" roster-body
snapshot_file "$composed" composed
snapshot_file "$manifest" manifest
snapshot_file "$skill_state" skill-state
snapshot_file "$mcporter" mcporter
snapshot_file "$claude_mcp" claude-mcp
snapshot_file "$codex_mcp" codex-mcp

second_output="$smoke_root/second.txt"
if ! env HOME="$native_root/home" USERPROFILE="$native_root/home" \
  PROJECTS_ROOT="$native_root/projects" "$binary" >"$second_output" 2>&1; then
  cat "$second_output" >&2
  fail "second acompose convergence failed"
fi

if grep -F "wrote" "$second_output" >/dev/null; then
  cat "$second_output" >&2
  fail "second convergence rewrote current state"
fi
assert_contains "$second_output" "cascade outputs=1 load-points=1 manifest=1 changed=0"
assert_contains "$second_output" "skills  managed=1 load-points=1 verified=1 linked=0 removed=0 preserved=0"
assert_contains "$second_output" "mcp     servers=1 state=unchanged"

assert_unchanged "$roster_table" roster-table
assert_unchanged "$roster_override" roster-override
assert_unchanged "$roster_body" roster-body
assert_unchanged "$composed" composed
assert_unchanged "$manifest" manifest
assert_unchanged "$skill_state" skill-state
assert_unchanged "$mcporter" mcporter
assert_unchanged "$claude_mcp" claude-mcp
assert_unchanged "$codex_mcp" codex-mcp

printf 'smoke: acompose host convergence is healthy and idempotent\n'
