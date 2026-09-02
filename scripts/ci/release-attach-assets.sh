#!/bin/sh
# Attach every built artifact to the created Forgejo release.
#
# The set is read off dist/ rather than restated here. A hardcoded list drifted
# once: the roster and bundles tarballs joined release-build.sh and the rendered
# formula in v2.90.0 but never this script, so v2.90.0 through v2.94.0 each
# published a formula naming two assets no release carried.
set -eu
for asset in dist/*
do
  [ -f "$asset" ] || continue
  name=$(basename "$asset")
  curl -fsSL -X POST \
    -H "Authorization: token ${FORGEJO_TOKEN}" \
    -F "attachment=@${asset}" \
    "${FORGEJO_API}/releases/${RELEASE_ID}/assets?name=${name}"
  echo "uploaded ${name}"
done
