#!/bin/sh
# Full repository validation: engine tests first, then the hook catalog.
set -e
go test ./...
pre-commit run --all-files
