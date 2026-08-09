#!/bin/sh
# Push the refreshed Scoop manifest, skipping when no write token is present.
set -eu
if [ -z "${SCOOP_WRITE_TOKEN:-}" ]; then
  echo "::warning::SCOOP_WRITE_TOKEN is absent; skipping Scoop update" >&2
  exit 0
fi
git clone --depth 1 \
  https://forgejo.coilysiren.me/coilyco-flight-deck/scoop-bucket.git bucket
cp dist/agent-compose.json bucket/bucket/agent-compose.json
cd bucket
if git diff --quiet; then
  exit 0
fi
git config user.name "coilyco-ops"
git config user.email "coilyco-ops@coilysiren.me"
git add bucket/agent-compose.json
git commit -m "chore(agent-compose): bump manifest to ${TAG} [skip ci]"
git push \
  "https://coilyco-ops:${SCOOP_WRITE_TOKEN}@forgejo.coilysiren.me/coilyco-flight-deck/scoop-bucket.git" \
  HEAD:main
