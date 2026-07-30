#!/bin/sh
set -eu

if [ -z "${AGENT_COMPOSE_EVALUATION_OUT:-}" ]; then
  echo "AGENT_COMPOSE_EVALUATION_OUT must name an empty output directory" >&2
  exit 2
fi

go run ./cmd/agent-compose evaluation \
  --all \
  --seat codex \
  --format yaml \
  --out-dir "${AGENT_COMPOSE_EVALUATION_OUT}"
