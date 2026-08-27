"""The differential oracle must not skip silently.

Every parity test skips when the Go toolchain is absent, and a skipped suite
reads exactly like a passing one in a CI summary. #339 deletes the Go engine on
the strength of this suite, so a run where it quietly stopped executing is the
worst way to arrive at that decision.

These assert the oracle is actually live wherever it is supposed to be.
"""

from __future__ import annotations

import os
import pathlib
import shutil

import pytest

REPO = pathlib.Path(__file__).resolve().parents[2]


def test_the_go_engine_is_still_here() -> None:
    """When this fails, #339 has landed and this whole package should be gone."""
    assert (REPO / "internal" / "person").is_dir()
    assert (REPO / "cmd" / "agent-compose").is_dir()


@pytest.mark.skipif(not os.environ.get("CI"), reason="local runs may legitimately lack Go")
def test_go_is_available_in_ci() -> None:
    """In CI the oracle has no excuse: a skip there is a suite that stopped running."""
    assert shutil.which("go") is not None, "CI has no Go, so every parity test skipped"


def test_housecast_is_resolved_from_the_pinned_dependency() -> None:
    """Parity has to test the published artifact, not a stray sibling checkout."""
    import housecast

    location = pathlib.Path(housecast.__file__).resolve()
    assert REPO not in location.parents or ".venv" in location.parts, (
        f"housecast resolved to {location}, which is inside the working tree "
        "rather than the pinned dependency"
    )
