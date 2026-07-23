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
(cd dist && sha256sum \
    agent-compose-darwin-arm64 \
    agent-compose-linux-amd64 \
    agent-compose-linux-arm64 \
    agent-compose-windows-amd64.exe > SHA256SUMS)
echo "version: ${VERSION}"
