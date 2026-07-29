#!/bin/sh
set -eu

budget_root=$(mktemp -d "${TMPDIR:-/tmp}/agent-compose-context-budget.XXXXXX")
trap 'rm -rf "$budget_root"' EXIT HUP INT TERM

go run ./cmd/agent-compose roster --out "$budget_root" >/dev/null
go run ./cmd/agent-compose compose testdata/contracts/native.kdl \
    --out "$budget_root/bundles" >/dev/null

for artifact in AGENTS.COMPOSE.md AGENTS.claude.md; do
    bytes=$(wc -c <"$budget_root/$artifact" | tr -d ' ')
    tokens=$(( (bytes + 3) / 4 ))
    printf '%s bytes=%s approximate_tokens=%s\n' "$artifact" "$bytes" "$tokens"
done

set -- "$budget_root"/bundles/*/content/instructions.md
bytes=$(wc -c <"$1" | tr -d ' ')
tokens=$(( (bytes + 3) / 4 ))
printf 'assigned-engineer-instructions.md bytes=%s approximate_tokens=%s\n' "$bytes" "$tokens"
