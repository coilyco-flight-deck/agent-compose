"""Map a retired role slug to its current one, from the table the Go launcher embeds.

challenges.yaml keeps the slugs it was authored and graded under, so its case ids
still line up with committed boards. Loading maps them to the roster's current
slugs. The table is internal/roleslug/retired.json, one file for both languages.
"""

from __future__ import annotations

import json
from functools import cache
from pathlib import Path
from typing import Any

TABLE = Path(__file__).resolve().parents[1] / "internal" / "roleslug" / "retired.json"
ROLE_FIELDS = ("entity", "attribute")


@cache
def _retired() -> dict[str, str]:
    loaded: dict[str, str] = json.loads(TABLE.read_text(encoding="utf-8"))
    return loaded


def canonical(slug: str) -> str:
    slug = slug.strip()
    return _retired().get(slug, slug)


def canonical_case(entry: dict[str, Any]) -> dict[str, Any]:
    """A raw challenge with its role-bearing fields on current slugs."""
    mapped = dict(entry)
    for field in ROLE_FIELDS:
        if isinstance(mapped.get(field), str):
            mapped[field] = canonical(mapped[field])
    return mapped
