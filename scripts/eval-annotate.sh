#!/bin/sh
# Grade the filtered dataset by hand. Renders the roster first so the per-role
# header carries purpose, boundaries, adjacency reasons, and personalities.
set -e
dataset=${EVAL_DATASET:-.evalkit/dataset.yaml}
out=${EVAL_ANNOTATIONS:-.evalkit/annotations.yaml}

if [ ! -f "$dataset" ]; then
  echo "no dataset at $dataset. Run evalkit-filter first." >&2
  exit 2
fi

render_dir=$(mktemp -d)
cleanup() { rm -rf "$render_dir"; }
trap cleanup EXIT HUP INT TERM
go run ./cmd/agent-compose roster --out "$render_dir" >/dev/null

uv run python -m evalkit.annotate \
  --dataset "$dataset" \
  --out "$out" \
  --roster "$render_dir/person.json" \
  "$@"
