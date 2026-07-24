#!/bin/sh
# Generate canonical palette data, install pinned web tools, then run one mode.
set -eu

mode=${1:?usage: palette-web.sh build|test|dev}
repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
site_dir="$repo_dir/web/personality-palette"

cd "$repo_dir"
go run ./cmd/agent-compose palette-data --out "web/personality-palette/public/palette.json"

cd "$site_dir"
npm ci --prefer-offline --ignore-scripts --no-audit --no-fund

case "$mode" in
  build)
    exec npm run build
    ;;
  test)
    exec npm test
    ;;
  dev)
    exec npm run dev -- --host 127.0.0.1
    ;;
  *)
    printf '%s\n' "unknown palette web mode: $mode" >&2
    exit 2
    ;;
esac
