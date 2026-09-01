#!/bin/sh
# Cross-compile release binaries into dist/, stamped with the current tag.
set -e
VERSION="$(git describe --tags --exact-match 2>/dev/null || git describe --tags --always)"
mkdir -p dist
for target in darwin/arm64 linux/amd64 linux/arm64 windows/amd64; do
    GOOS="${target%/*}"
    GOARCH="${target#*/}"
    out="dist/agent-compose-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        out="${out}.exe"
    fi
    GOOS="$GOOS" GOARCH="$GOARCH" CGO_ENABLED=0 \
        go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" \
        -o "$out" ./cmd/agent-compose
    echo "$out"
done
# The binary embeds no roster, so the seed ships beside it as its own asset.
tar -czf dist/agent-compose-roster.tar.gz -C seed roster
echo "dist/agent-compose-roster.tar.gz"
# housecast never runs on a user's machine, so the composed set ships too.
uv run python scripts/compose-bundles.py dist/bundles
tar -czf dist/agent-compose-bundles.tar.gz -C dist bundles
echo "dist/agent-compose-bundles.tar.gz"
(cd dist && sha256sum \
    agent-compose-roster.tar.gz \
    agent-compose-bundles.tar.gz \
    agent-compose-darwin-arm64 \
    agent-compose-linux-amd64 \
    agent-compose-linux-arm64 \
    agent-compose-windows-amd64.exe > SHA256SUMS)
echo "version: ${VERSION}"
