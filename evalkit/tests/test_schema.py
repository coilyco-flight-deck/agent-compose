from __future__ import annotations

import pytest

from evalkit.schema import (
    BoardCase,
    Candidate,
    Fit,
    Grade,
    Half,
    Kind,
    Verdict,
    grading_order,
    pair_results,
)


def boundary(role: str, half: Half, variant: int = 1) -> Candidate:
    return Candidate(
        id=f"{role}-shc-{half.value}",
        role=role,
        kind=Kind.BOUNDARY,
        prompt="prompt",
        expected="expected",
        variant=variant,
        discriminator="drafts the message instead of handing off",
        boundary="suggest-human-comms",
        half=half,
        pair_id=f"{role}-shc",
    )


def case(candidate: Candidate) -> BoardCase:
    return BoardCase(candidate=candidate, response="response", failure_count=2)


def test_boundary_case_requires_its_pair_identity() -> None:
    with pytest.raises(ValueError, match="boundary, half, and pair_id"):
        Candidate(
            id="ops-shc-out",
            role="ops",
            kind=Kind.BOUNDARY,
            prompt="p",
            expected="e",
            discriminator="d",
        )


def test_pass_fail_case_requires_a_discriminator() -> None:
    with pytest.raises(ValueError, match="needs a discriminator"):
        Candidate(
            id="ops-fit-within",
            role="ops",
            kind=Kind.ROLE_FIT,
            prompt="p",
            expected="e",
            target="within",
        )


def test_personality_case_rejects_a_discriminator() -> None:
    with pytest.raises(ValueError, match="cannot carry a discriminator"):
        Candidate(
            id="ops-per-grounded",
            role="ops",
            kind=Kind.PERSONALITY,
            prompt="p",
            expected="e",
            discriminator="d",
            trait="grounded",
        )


def test_word_caps_split_by_scoring_tier() -> None:
    assert boundary("ops", Half.OUT).word_cap == 50
    personality = Candidate(
        id="ops-per", role="ops", kind=Kind.PERSONALITY, prompt="p", expected="e"
    )
    assert personality.word_cap == 100


def personality(role: str) -> BoardCase:
    return case(
        Candidate(id=f"{role}-per", role=role, kind=Kind.PERSONALITY, prompt="p", expected="e")
    )


def test_grading_order_keeps_a_role_together_so_its_charter_loads_once() -> None:
    cases = [
        personality("ops"),
        case(boundary("creator", Half.IN)),
        case(boundary("ops", Half.IN)),
        personality("creator"),
    ]
    ordered = [entry.id for entry in grading_order(cases, ["ops", "creator"])]
    assert ordered == ["ops-shc-in", "ops-per", "creator-shc-in", "creator-per"]


def test_boundaries_precede_personality_inside_a_role() -> None:
    cases = [personality("ops"), case(boundary("ops", Half.OUT))]
    assert [entry.id for entry in grading_order(cases, ["ops"])] == ["ops-shc-out", "ops-per"]


def test_role_order_falls_back_to_alphabetical_without_a_roster() -> None:
    cases = [personality("qa"), personality("ai")]
    assert [entry.id for entry in grading_order(cases)] == ["ai-per", "qa-per"]


def test_a_split_pair_is_a_boundary_failure_not_a_half_pass() -> None:
    cases = [case(boundary("ops", Half.IN)), case(boundary("ops", Half.OUT))]
    grades = {
        "ops-shc-in": Grade(id="ops-shc-in", verdict=Verdict.PASS),
        "ops-shc-out": Grade(id="ops-shc-out", verdict=Verdict.FAIL),
    }
    results = pair_results(cases, grades)
    assert len(results) == 1
    assert results[0].complete
    assert not results[0].passed


def test_a_pair_passes_only_when_both_halves_pass() -> None:
    cases = [case(boundary("ops", Half.IN)), case(boundary("ops", Half.OUT))]
    grades = {
        "ops-shc-in": Grade(id="ops-shc-in", verdict=Verdict.PASS),
        "ops-shc-out": Grade(id="ops-shc-out", verdict=Verdict.PASS),
    }
    assert pair_results(cases, grades)[0].passed


def test_undecided_counts_as_a_deduction_so_it_earns_a_note() -> None:
    assert Grade(id="x", verdict=Fit.UNDECIDED).is_deduction
    assert Grade(id="x", verdict=Fit.NO_FIT).is_deduction
    assert not Grade(id="x", verdict=Fit.FIT).is_deduction
