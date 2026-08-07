#!/bin/sh
# Re-earn the committed evaluation baseline: render packs, drive every active
# case, review the preserved responses in independent sessions, and write the
# canonical records. Each half runs in its own isolated home so the reviewer
# never inherits the driver's projected role context.
set -eu

require() {
  eval "value=\${$1:-}"
  if [ -z "${value}" ]; then
    echo "$1 must be set" >&2
    exit 2
  fi
}

require ACOMPOSE_EVAL_DRIVER_HOME
require ACOMPOSE_EVAL_REVIEWER_HOME
require ACOMPOSE_EVAL_ISSUE
require ACOMPOSE_EVAL_DATE

SEAT="${ACOMPOSE_EVAL_SEAT:-claude}"
TIER="${ACOMPOSE_EVAL_TIER:-frontier}"
DRIVER_MODEL="${ACOMPOSE_EVAL_DRIVER_MODEL:-claude-opus-5}"
DRIVER_EFFORT="${ACOMPOSE_EVAL_DRIVER_EFFORT:-medium}"
REVIEWER_MODEL="${ACOMPOSE_EVAL_REVIEWER_MODEL:-claude-opus-5}"
REVIEWER_EFFORT="${ACOMPOSE_EVAL_REVIEWER_EFFORT:-high}"
JOBS="${ACOMPOSE_EVAL_JOBS:-4}"
WORK="${ACOMPOSE_EVAL_WORK:-$(mktemp -d)}"

PACKS="${WORK}/packs"
RUNS="${WORK}/runs"
SESSIONS="${WORK}/sessions"
REVIEW="${WORK}/review.json"

mkdir -p "${WORK}"
rm -rf "${PACKS}"
mkdir -p "${PACKS}" "${RUNS}" "${SESSIONS}"

echo "==> rendering ${SEAT} packs into ${PACKS}" >&2
go run ./cmd/agent-compose evaluation \
  --all \
  --seat "${SEAT}" \
  --format yaml \
  --out-dir "${PACKS}"

echo "==> driving ${TIER} cases with ${DRIVER_MODEL} at ${DRIVER_EFFORT}" >&2
python3 scripts/evaluation_driver.py \
  --packs "${PACKS}" \
  --out "${RUNS}" \
  --sessions "${SESSIONS}" \
  --home "${ACOMPOSE_EVAL_DRIVER_HOME}" \
  --harness "${SEAT}" \
  --tier "${TIER}" \
  --jobs "${JOBS}" \
  --arm "baseline:${DRIVER_MODEL}:${DRIVER_EFFORT}"

echo "==> reviewing with ${REVIEWER_MODEL} at ${REVIEWER_EFFORT}" >&2
python3 scripts/evaluation_reviewer.py \
  --packs "${PACKS}" \
  --runs "${RUNS}/baseline.json" \
  --out "${REVIEW}" \
  --model "${REVIEWER_MODEL}" \
  --effort "${REVIEWER_EFFORT}" \
  --home "${ACOMPOSE_EVAL_REVIEWER_HOME}" \
  --jobs "${JOBS}"

echo "==> writing records into evaluations/latest" >&2
go run ./scripts/evaluation-record \
  --run "${RUNS}/baseline.json" \
  --review "${REVIEW}" \
  --out evaluations/latest \
  --seat "${SEAT}" \
  --issue "${ACOMPOSE_EVAL_ISSUE}" \
  --evaluated-at "${ACOMPOSE_EVAL_DATE}"

echo "==> refreshing the scorecard" >&2
go run ./cmd/agent-compose scorecard \
  --results evaluations/latest \
  --out docs/evaluation-scores.md

echo "work directory: ${WORK}" >&2
