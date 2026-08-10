#!/bin/sh
# Compose one bundle per role and write its delivery as <out>/<role>.md, which
# is what evalkit.run sends as the system prompt. Frontier is the only tier
# every role supports, and model tier does not change selected context.
set -e
out=${1:-.evalkit/prompts}
mkdir -p "$out"

work=$(mktemp -d)
cleanup() { rm -rf "$work"; }
trap cleanup EXIT HUP INT TERM

go run ./cmd/agent-compose roster --out "$work/roster" >/dev/null
roles=$(python3 -c "import json,sys; print(' '.join(json.load(open(sys.argv[1]))['role_order']))" \
  "$work/roster/person.json")

for role in $roles; do
  printf 'compose {\n    role "%s"\n    delivery "compiled"\n    model-tier "frontier"\n}\n' \
    "$role" > "$work/request.kdl"
  go run ./cmd/agent-compose compose --out "$work/bundles/$role" "$work/request.kdl" >/dev/null
  compiled=$(find "$work/bundles/$role" -name compiled.md -type f | head -1)
  if [ -z "$compiled" ]; then
    echo "no compiled delivery for role $role" >&2
    exit 1
  fi
  cp "$compiled" "$out/$role.md"
  printf '%s\t%s words\n' "$role" "$(wc -w < "$out/$role.md" | tr -d ' ')"
done
