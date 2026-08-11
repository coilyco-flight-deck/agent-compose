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


def test_a_candidate_every_run_passes_is_dropped_as_too_easy() -> None:
    sample = role_fit("director", 1)
    report = item_filter.run([sample], responses(sample, failures=0), pattern_matcher)
    assert report.kept == []
    assert report.dropped[0].reason == "every run passed"


def test_a_candidate_every_run_fails_is_dropped_as_broken() -> None:
    sample = role_fit("director", 1)
    report = item_filter.run([sample], responses(sample, failures=5), pattern_matcher)
    assert report.kept == []
    assert report.dropped[0].reason == "every run failed"


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


def test_a_half_without_its_partner_drops_the_whole_pair() -> None:
    lonely = Sample(
        id="ops-shc-out",
        role="ops",
        test_type=TestType.BOUNDARY,
        prompt="prompt",
        target="target",
        discriminator=[r"absorbed"],
        boundary="suggest-human-comms",
        half=Half.OUT,
        pair_id="ops-shc",
    )
    report = item_filter.run([lonely], responses(lonely, failures=2), pattern_matcher)
    assert report.kept == []
    assert any("incomplete" in drop.reason for drop in report.dropped)


def test_every_drop_is_reported_so_truncation_is_never_silent() -> None:
    sample = role_fit("director", 1)
    report = item_filter.run([sample], [], pattern_matcher)
    assert report.dropped[0].reason == "no subject runs"
    assert report.summary == "0 kept, 1 dropped"


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
