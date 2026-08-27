#!/bin/sh
# Python gates for the checks package. The pre-commit suite is managed by
# agentic-os, so repo-local Python linting runs here instead.
set -e
uv run ruff check checks
uv run ruff format --check checks
uv run mypy
uv run pytest -q
