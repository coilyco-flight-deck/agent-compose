#!/bin/sh
# Print the case list the current roster implies. Go exports the roster, Python
# derives the board, so adding a boundary changes this output on its own.
set -e
render_dir=$(mktemp -d)
cleanup() { rm -rf "$render_dir"; }
trap cleanup EXIT HUP INT TERM
go run ./cmd/agent-compose roster --out "$render_dir" >/dev/null
uv run python -m evalkit.matrix --roster "$render_dir/person.json" "$@"
