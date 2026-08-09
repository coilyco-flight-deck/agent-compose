#!/bin/sh
# Attach every built artifact to the created Forgejo release.
set -eu
for asset in \
  dist/agent-compose-darwin-arm64 \
  dist/agent-compose-linux-amd64 \
  dist/agent-compose-linux-arm64 \
  dist/agent-compose-windows-amd64.exe \
  dist/SHA256SUMS \
  dist/agent-compose.rb \
  dist/agent-compose.json
do
  name=$(basename "$asset")
  curl -fsSL -X POST \
    -H "Authorization: token ${FORGEJO_TOKEN}" \
    -F "attachment=@${asset}" \
    "${FORGEJO_API}/releases/${RELEASE_ID}/assets?name=${name}"
  echo "uploaded ${name}"
done
