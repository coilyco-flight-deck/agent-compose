#!/bin/sh
# Build the release binaries and their package metadata.
set -eu
ward exec release-build
ward exec release-package
