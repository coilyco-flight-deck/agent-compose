from __future__ import annotations

from evalkit import filter as item_filter
from evalkit.schema import Candidate, Half, Kind, Response


def substring_matcher(discriminator: str, response: str) -> bool:
    return discriminator.lower() in response.lower()


def role_fit(target: str, variant: int, discriminator: str = "absorbed") -> Candidate:
    return Candidate(
        id=f"ops-fit-{target}-v{variant}",
        role="ops",
        kind=Kind.ROLE_FIT,
        prompt="prompt",
        expected="expected",
        variant=variant,
        discriminator=discriminator,
        target=target,
    )


def responses(candidate: Candidate, failures: int, total: int = 5) -> list[Response]:
    return [
        Response(
            candidate_id=candidate.id,
            variant=candidate.variant,
            run=index,
            text="absorbed the work" if index <= failures else "handed it back",
        )
        for index in range(1, total + 1)
    ]


def test_a_candidate_every_run_passes_is_dropped_as_too_easy() -> None:
    candidate = role_fit("director", 1)
    report = item_filter.run([candidate], responses(candidate, failures=0), substring_matcher)
    assert report.kept == []
    assert report.dropped[0].reason == "every run passed"


def test_a_candidate_every_run_fails_is_dropped_as_broken() -> None:
    candidate = role_fit("director", 1)
    report = item_filter.run([candidate], responses(candidate, failures=5), substring_matcher)
    assert report.kept == []
    assert report.dropped[0].reason == "every run failed"


def test_the_candidate_closest_to_the_midpoint_wins_its_slot() -> None:
    weak = role_fit("director", 1)
    strong = role_fit("director", 2)
    runs = responses(weak, failures=1) + responses(strong, failures=3)
    report = item_filter.run([weak, strong], runs, substring_matcher)
    assert [entry.candidate.variant for entry in report.kept] == [2]
    assert any("lost slot" in drop.reason for drop in report.dropped)


def test_personality_candidates_bypass_the_filter() -> None:
    candidate = Candidate(
        id="ops-per-grounded",
        role="ops",
        kind=Kind.PERSONALITY,
        prompt="prompt",
        expected="expected",
        trait="grounded",
    )
    runs = [
        Response(candidate_id=candidate.id, variant=1, run=index, text="steady")
        for index in range(1, 6)
    ]
    report = item_filter.run([candidate], runs, substring_matcher)
    assert len(report.kept) == 1
    assert report.kept[0].failure_count == 0


def test_a_half_without_its_partner_drops_the_whole_pair() -> None:
    lonely = Candidate(
        id="ops-shc-out",
        role="ops",
        kind=Kind.BOUNDARY,
        prompt="prompt",
        expected="expected",
        discriminator="absorbed",
        boundary="suggest-human-comms",
        half=Half.OUT,
        pair_id="ops-shc",
    )
    report = item_filter.run([lonely], responses(lonely, failures=2), substring_matcher)
    assert report.kept == []
    assert any("incomplete" in drop.reason for drop in report.dropped)


def test_every_drop_is_reported_so_truncation_is_never_silent() -> None:
    candidate = role_fit("director", 1)
    report = item_filter.run([candidate], [], substring_matcher)
    assert report.dropped[0].reason == "no subject runs"
    assert report.summary == "0 kept, 1 dropped"
