#!/bin/sh
# Python gates for evalkit. The pre-commit suite is managed by agentic-os, so
# repo-local Python linting runs here instead of being hand-added to that file.
set -e
uv run ruff check evalkit
uv run ruff format --check evalkit
uv run mypy
uv run pytest -q
