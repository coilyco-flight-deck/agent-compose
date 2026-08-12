from __future__ import annotations

import re

from evalkit import filter as item_filter
from evalkit.schema import Half, Response, Sample, TestType


def pattern_matcher(patterns: list[str], response: str) -> bool:
    return any(re.search(p, response, re.IGNORECASE) for p in patterns)


def role_fit(against: str, variant: int, discriminator: list[str] | None = None) -> Sample:
    return Sample(
        id=f"ops-fit-{against}-v{variant}",
        role="ops",
        test_type=TestType.ROLE_FIT,
        prompt="prompt",
        target="hands it back",
        variant=variant,
        discriminator=discriminator or [r"absorbed"],
        against=against,
    )


def responses(sample: Sample, failures: int, total: int = 5) -> list[Response]:
    return [
        Response(
            sample_id=sample.id,
            variant=sample.variant,
            epoch=index,
            text="absorbed the work" if index <= failures else "handed it back",
        )
        for index in range(1, total + 1)
    ]


def test_a_candidate_every_run_passes_is_kept_and_noted() -> None:
    sample = role_fit("director", 1)
    report = item_filter.run([sample], responses(sample, failures=0), pattern_matcher)
    assert [entry.id for entry in report.kept] == [sample.id]
    assert "never discriminated" in report.notes[0].reason


def test_a_candidate_every_run_fails_is_kept_and_noted() -> None:
    sample = role_fit("director", 1)
    report = item_filter.run([sample], responses(sample, failures=5), pattern_matcher)
    assert [entry.id for entry in report.kept] == [sample.id]
    assert "failed every run" in report.notes[0].reason


def test_a_negative_control_that_passes_is_not_noted() -> None:
    within = role_fit("within", 1)
    report = item_filter.run([within], responses(within, failures=0), pattern_matcher)
    assert [entry.id for entry in report.kept] == [within.id]
    assert report.notes == []


def test_the_candidate_closest_to_the_midpoint_wins_its_slot() -> None:
    weak = role_fit("director", 1)
    strong = role_fit("director", 2)
    runs = responses(weak, failures=1) + responses(strong, failures=3)
    report = item_filter.run([weak, strong], runs, pattern_matcher)
    assert [entry.sample.variant for entry in report.kept] == [2]
    assert any("lost slot" in drop.reason for drop in report.dropped)


def test_personality_candidates_bypass_the_filter() -> None:
    sample = Sample(
        id="ops-per-grounded",
        role="ops",
        test_type=TestType.PERSONALITY,
        prompt="prompt",
        target="target",
        trait="grounded",
    )
    runs = [
        Response(sample_id=sample.id, variant=1, epoch=index, text="steady")
        for index in range(1, 6)
    ]
    report = item_filter.run([sample], runs, pattern_matcher)
    assert len(report.kept) == 1
    assert report.kept[0].failure_count == 0


def boundary_half(half: Half, discriminator: list[str] | None = None) -> Sample:
    return Sample(
        id=f"ops-shc-{half.value}",
        role="ops",
        test_type=TestType.BOUNDARY,
        prompt="prompt",
        target=f"target {half.value}",
        discriminator=discriminator or [r"absorbed"],
        boundary="suggest-human-comms",
        half=half,
        pair_id="ops-shc",
    )


def test_a_half_without_its_partner_is_kept_and_noted() -> None:
    lonely = boundary_half(Half.OUT)
    report = item_filter.run([lonely], responses(lonely, failures=2), pattern_matcher)
    assert [entry.id for entry in report.kept] == [lonely.id]
    assert any("missing a half" in note.reason for note in report.notes)


# The control holding while the far half fails is the most informative shape a
# boundary pair produces. The old rule deleted it. See docs/eval-annotation.md.
def test_a_holding_control_and_a_failing_far_half_both_survive() -> None:
    control = boundary_half(Half.IN)
    far = boundary_half(Half.OUT)
    runs = responses(control, failures=0) + responses(far, failures=4)
    report = item_filter.run([control, far], runs, pattern_matcher)
    assert sorted(entry.id for entry in report.kept) == sorted([control.id, far.id])
    assert {entry.id: entry.failure_count for entry in report.kept}[far.id] == 4
    assert report.notes == []


def test_identical_halves_are_noted_rather_than_dropped() -> None:
    control = boundary_half(Half.IN)
    far = boundary_half(Half.OUT)
    runs = responses(control, failures=2) + responses(far, failures=2)
    report = item_filter.run([control, far], runs, pattern_matcher)
    assert len(report.kept) == 2
    assert any("behaved identically" in note.reason for note in report.notes)


def test_every_drop_is_reported_so_truncation_is_never_silent() -> None:
    sample = role_fit("director", 1)
    report = item_filter.run([sample], [], pattern_matcher)
    assert report.dropped[0].reason == "no subject runs"
    assert report.summary == "0 kept, 1 dropped, 0 noted"


def test_regex_matcher_fires_on_any_pattern() -> None:
    patterns = [r"\bdraft(s|ed)?\b", r"^Dear \w+"]
    assert item_filter.regex_matcher(patterns, "I drafted the notice")
    assert item_filter.regex_matcher(patterns, "Dear players, ...")
    assert not item_filter.regex_matcher(patterns, "handed the facts over")


def test_regex_matcher_ignores_case_and_spans_lines() -> None:
    assert item_filter.regex_matcher([r"^announcement"], "context\nAnnouncement: outage")


def test_a_discriminator_that_is_not_a_regex_fails_at_load() -> None:
    import pytest

    with pytest.raises(ValueError, match="is not a regex"):
        Sample(
            id="ops-fit-within",
            role="ops",
            test_type=TestType.ROLE_FIT,
            prompt="p",
            target="t",
            against="within",
            discriminator=[r"unclosed ("],
        )
