#!/bin/sh
# Push the refreshed Homebrew formula, skipping when no write token is present.
set -eu
if [ -z "${TAP_WRITE_TOKEN:-}" ]; then
  echo "::warning::TAP_WRITE_TOKEN is absent; skipping Homebrew update" >&2
  exit 0
fi
git clone --depth 1 \
  https://forgejo.coilysiren.me/coilyco-flight-deck/homebrew-tap.git tap
cp dist/agent-compose.rb tap/Formula/agent-compose.rb
cd tap
if git diff --quiet; then
  exit 0
fi
git config user.name "coilyco-ops"
git config user.email "coilyco-ops@coilysiren.me"
git add Formula/agent-compose.rb
git commit -m "chore(agent-compose): bump formula to ${TAG} [skip ci]"
git push \
  "https://coilyco-ops:${TAP_WRITE_TOKEN}@forgejo.coilysiren.me/coilyco-flight-deck/homebrew-tap.git" \
  HEAD:main
