#!/bin/sh
# Full repository validation: engine tests first, then the hook catalog.
set -e
go test ./...
go run ./cmd/agent-compose scorecard --results evaluations/latest --out docs/evaluation-scores.md --historical --check
sh scripts/palette-web.sh test
sh scripts/context-budget.sh
pre-commit run --all-files
