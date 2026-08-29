#!/bin/sh
# Fail the release unless the GitHub mirror actually carries the tag.
#
# GitHub is the module origin: `go.mod` names github.com, so proxy.golang.org
# fetches the tag from there and a tag that never arrives is a release no Go
# consumer can resolve. Nothing else in this pipeline looks at the mirror, so
# without this a failed sync is green everywhere and invisible.
#
# Verifies rather than pushes. Something already syncs refs to the mirror
# outside this workflow, and a second pusher would race it. A read proves the
# property either way and needs no credential, because the mirror is public.
# See docs/release.md.

set -eu

: "${TAG:?TAG is required}"
MIRROR="${MIRROR_REPO:-coilyco-flight-deck/agent-compose}"
API="${GITHUB_API:-https://api.github.com}"
# The sync is asynchronous, so a miss on the first read means nothing.
ATTEMPTS="${MIRROR_ATTEMPTS:-60}"
DELAY="${MIRROR_DELAY:-10}"

url="$API/repos/$MIRROR/git/ref/tags/$TAG"
attempt=1

while [ "$attempt" -le "$ATTEMPTS" ]; do
  code=$(curl -s -o /dev/null -w '%{http_code}' -m 20 "$url" || echo 000)
  if [ "$code" = "200" ]; then
    echo "release-verify-mirror: $MIRROR carries $TAG after ${attempt} attempt(s)."
    exit 0
  fi
  # 403 is the anonymous rate limit rather than a missing tag, so it is worth
  # saying out loud: the retry is fine, a wall of them is the real problem.
  [ "$code" = "403" ] && echo "release-verify-mirror: rate limited, retrying." >&2
  attempt=$((attempt + 1))
  sleep "$DELAY"
done

echo "::error::release-verify-mirror: $MIRROR does not carry $TAG after $((ATTEMPTS * DELAY))s." >&2
echo "GitHub is the module origin, so this release is unresolvable to Go consumers." >&2
echo "The tag exists on Forgejo. Check the push mirror, then re-run this job." >&2
exit 1
