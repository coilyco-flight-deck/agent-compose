#!/bin/sh
# Classify product release impact and publish the verdict as a step output.
set -eu
publish="$(ward exec release-impact)"
echo "publish=${publish}" >> "${GITHUB_OUTPUT}"
