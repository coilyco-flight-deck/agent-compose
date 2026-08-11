from __future__ import annotations

import pytest

from evalkit.schema import (
    Annotation,
    DatasetEntry,
    Fit,
    Half,
    Sample,
    TestType,
    Verdict,
    annotation_order,
    pair_results,
)


def boundary(role: str, half: Half, variant: int = 1) -> Sample:
    return Sample(
        id=f"{role}-shc-{half.value}",
        role=role,
        test_type=TestType.BOUNDARY,
        prompt="prompt",
        target="target",
        variant=variant,
        discriminator=[r"\bdraft(s|ed)?\b"],
        boundary="suggest-human-comms",
        half=half,
        pair_id=f"{role}-shc",
    )


def case(sample: Sample) -> DatasetEntry:
    return DatasetEntry(sample=sample, output="output", failure_count=2)


def test_boundary_case_requires_its_pair_identity() -> None:
    with pytest.raises(ValueError, match="boundary, half, and pair_id"):
        Sample(
            id="ops-shc-out",
            role="ops",
            test_type=TestType.BOUNDARY,
            prompt="p",
            target="e",
            discriminator=[r"d"],
        )


def test_pass_fail_case_requires_a_discriminator() -> None:
    with pytest.raises(ValueError, match="needs a discriminator"):
        Sample(
            id="ops-fit-within",
            role="ops",
            test_type=TestType.ROLE_FIT,
            prompt="p",
            target="e",
            against="within",
        )


def test_personality_case_rejects_a_discriminator() -> None:
    with pytest.raises(ValueError, match="cannot carry a discriminator"):
        Sample(
            id="ops-per-grounded",
            role="ops",
            test_type=TestType.PERSONALITY,
            prompt="p",
            target="e",
            discriminator=[r"d"],
            trait="grounded",
        )


def test_word_caps_split_by_scoring_tier() -> None:
    assert boundary("ops", Half.OUT).word_cap == 50
    personality = Sample(
        id="ops-per", role="ops", test_type=TestType.PERSONALITY, prompt="p", target="e"
    )
    assert personality.word_cap == 100


def personality(role: str) -> DatasetEntry:
    return case(
        Sample(id=f"{role}-per", role=role, test_type=TestType.PERSONALITY, prompt="p", target="e")
    )


def test_grading_order_keeps_a_role_together_so_its_charter_loads_once() -> None:
    cases = [
        personality("ops"),
        case(boundary("creator", Half.IN)),
        case(boundary("ops", Half.IN)),
        personality("creator"),
    ]
    ordered = [entry.id for entry in annotation_order(cases, ["ops", "creator"])]
    assert ordered == ["ops-shc-in", "ops-per", "creator-shc-in", "creator-per"]


def test_boundaries_precede_personality_inside_a_role() -> None:
    cases = [personality("ops"), case(boundary("ops", Half.OUT))]
    assert [entry.id for entry in annotation_order(cases, ["ops"])] == ["ops-shc-out", "ops-per"]


def test_role_order_falls_back_to_alphabetical_without_a_roster() -> None:
    cases = [personality("qa"), personality("ai")]
    assert [entry.id for entry in annotation_order(cases)] == ["ai-per", "qa-per"]


def test_a_split_pair_is_a_boundary_failure_not_a_half_pass() -> None:
    cases = [case(boundary("ops", Half.IN)), case(boundary("ops", Half.OUT))]
    grades = {
        "ops-shc-in": Annotation(id="ops-shc-in", label=Verdict.PASS),
        "ops-shc-out": Annotation(id="ops-shc-out", label=Verdict.FAIL),
    }
    results = pair_results(cases, grades)
    assert len(results) == 1
    assert results[0].complete
    assert not results[0].passed


def test_a_pair_passes_only_when_both_halves_pass() -> None:
    cases = [case(boundary("ops", Half.IN)), case(boundary("ops", Half.OUT))]
    grades = {
        "ops-shc-in": Annotation(id="ops-shc-in", label=Verdict.PASS),
        "ops-shc-out": Annotation(id="ops-shc-out", label=Verdict.PASS),
    }
    assert pair_results(cases, grades)[0].passed


def test_undecided_counts_as_a_deduction_so_it_earns_a_note() -> None:
    assert Annotation(id="x", label=Fit.UNDECIDED).is_deduction
    assert Annotation(id="x", label=Fit.NO_FIT).is_deduction
    assert not Annotation(id="x", label=Fit.FIT).is_deduction
