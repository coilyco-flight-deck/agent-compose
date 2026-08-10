"""Round-trips through the on-disk formats.

A partially installed pyyaml once let every other gate pass green while both
CLIs raised on import, because nothing exercised the serialization path.
"""

from __future__ import annotations

import json
from pathlib import Path

import yaml

from evalkit import filter as item_filter
from evalkit import grade
from evalkit.schema import BoardCase, Candidate, Fit, Grade, Half, Kind, Response, Verdict


def test_pyyaml_and_httpx_expose_what_the_clis_call() -> None:
    import httpx

    assert hasattr(yaml, "safe_load")
    assert hasattr(yaml, "safe_dump")
    assert hasattr(httpx, "AsyncClient")


def board_case() -> BoardCase:
    return BoardCase(
        candidate=Candidate(
            id="ops-shc-out",
            role="ops",
            kind=Kind.BOUNDARY,
            prompt="prompt",
            expected="expected",
            discriminator="drafts the announcement",
            boundary="suggest-human-comms",
            half=Half.OUT,
            pair_id="ops-shc",
        ),
        response="handed it over",
        failure_count=2,
    )


def test_a_board_survives_a_write_and_read(tmp_path: Path) -> None:
    path = tmp_path / "board.yaml"
    original = board_case()
    path.write_text(yaml.safe_dump({"board": [original.to_dict()]}, sort_keys=False))

    loaded = grade.load_board(path)
    assert len(loaded) == 1
    assert loaded[0].candidate == original.candidate
    assert loaded[0].response == original.response
    assert loaded[0].failure_count == original.failure_count


def test_grades_survive_a_write_and_read(tmp_path: Path) -> None:
    path = tmp_path / "grades.yaml"
    grades = {
        "a": Grade(id="a", verdict=Verdict.FAIL, note="drafted it"),
        "b": Grade(id="b", verdict=Fit.UNDECIDED, note="case was ambiguous"),
        "c": Grade(id="c", verdict=Verdict.PASS),
    }
    grade.save_grades(path, grades)

    loaded = grade.load_grades(path)
    assert loaded["a"].verdict is Verdict.FAIL
    assert loaded["a"].note == "drafted it"
    assert loaded["b"].verdict is Fit.UNDECIDED
    assert loaded["c"].note == ""


def test_responses_survive_a_jsonl_round_trip(tmp_path: Path) -> None:
    path = tmp_path / "responses.jsonl"
    original = Response(
        candidate_id="ops-shc-out",
        variant=1,
        run=1,
        text="handed it over",
        reasoning="a long private deliberation",
    )
    path.write_text(json.dumps(original.to_dict()) + "\n")

    loaded = item_filter.load_responses(path)
    assert loaded == [original]
    assert loaded[0].words == 3, "reasoning must not count against the word cap"


def test_a_response_without_reasoning_omits_the_key() -> None:
    plain = Response(candidate_id="x", variant=1, run=1, text="hi")
    assert "reasoning" not in plain.to_dict()
