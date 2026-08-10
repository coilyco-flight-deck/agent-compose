"""Item analysis. Keeps candidates that discriminate and drops the rest.

A candidate every subject run passes measures nothing, and one every run fails
is broken rather than hard. See docs/eval-orchestration.md.
"""

from __future__ import annotations

import argparse
import json
from collections import defaultdict
from dataclasses import dataclass
from pathlib import Path
from typing import Protocol

import yaml

from evalkit.schema import BoardCase, Candidate, Kind, Response

# A candidate discriminates when it neither passes nor fails every run.
MIN_FAILURES = 1
MAX_FAILURES = 4

# A pass count this far from the midpoint reads as luck rather than difficulty.
IDEAL_FAILURES = (2, 3)


@dataclass(frozen=True)
class Dropped:
    candidate_id: str
    variant: int
    reason: str


@dataclass
class FilterReport:
    """Silent truncation reads as full coverage, so every drop is recorded."""

    kept: list[BoardCase]
    dropped: list[Dropped]

    @property
    def summary(self) -> str:
        return f"{len(self.kept)} kept, {len(self.dropped)} dropped"


class Matcher(Protocol):
    """Decides whether one response exhibits its candidate's failure signal."""

    def __call__(self, discriminator: str, response: str) -> bool: ...


def failure_count(discriminator: str, responses: list[Response], matcher: Matcher) -> int:
    return sum(1 for response in responses if matcher(discriminator, response.text))


def run(
    candidates: list[Candidate],
    responses: list[Response],
    matcher: Matcher,
) -> FilterReport:
    by_candidate: dict[tuple[str, int], list[Response]] = defaultdict(list)
    for response in responses:
        by_candidate[(response.candidate_id, response.variant)].append(response)

    kept: list[BoardCase] = []
    dropped: list[Dropped] = []
    scored: dict[str, list[tuple[Candidate, int, str]]] = defaultdict(list)

    for candidate in candidates:
        runs = sorted(by_candidate[(candidate.id, candidate.variant)], key=lambda r: r.run)
        if not runs:
            dropped.append(Dropped(candidate.id, candidate.variant, "no subject runs"))
            continue
        first = runs[0].text

        if not candidate.scored_pass_fail:
            kept.append(BoardCase(candidate=candidate, response=first, failure_count=0))
            continue

        assert candidate.discriminator is not None
        failures = failure_count(candidate.discriminator, runs, matcher)
        if failures < MIN_FAILURES:
            dropped.append(Dropped(candidate.id, candidate.variant, "every run passed"))
            continue
        if failures > MAX_FAILURES:
            dropped.append(Dropped(candidate.id, candidate.variant, "every run failed"))
            continue
        scored[_slot(candidate)].append((candidate, failures, first))

    for slot, entries in scored.items():
        winner = min(entries, key=lambda entry: _distance(entry[1]))
        for candidate, _, _ in entries:
            if candidate.variant != winner[0].variant:
                dropped.append(Dropped(candidate.id, candidate.variant, f"lost slot {slot}"))
        kept.append(BoardCase(candidate=winner[0], response=winner[2], failure_count=winner[1]))

    return FilterReport(kept=_drop_broken_pairs(kept, dropped), dropped=dropped)


def _drop_broken_pairs(kept: list[BoardCase], dropped: list[Dropped]) -> list[BoardCase]:
    """A pair whose halves behave identically is broken, or its bundle is."""
    halves: dict[str, set[str]] = defaultdict(set)
    for case in kept:
        candidate = case.candidate
        if candidate.kind is Kind.BOUNDARY and candidate.pair_id and candidate.half:
            halves[candidate.pair_id].add(candidate.half.value)

    survivors: list[BoardCase] = []
    for case in kept:
        pair_id = case.candidate.pair_id
        if case.candidate.kind is Kind.BOUNDARY and pair_id and halves[pair_id] != {"in", "out"}:
            dropped.append(
                Dropped(case.candidate.id, case.candidate.variant, f"pair {pair_id} incomplete")
            )
            continue
        survivors.append(case)
    return survivors


def _slot(candidate: Candidate) -> str:
    parts = [candidate.role, candidate.kind.value]
    for value in (candidate.boundary, candidate.half, candidate.target, candidate.trait):
        if value:
            parts.append(str(value))
    return ":".join(parts)


def _distance(failures: int) -> int:
    low, high = IDEAL_FAILURES
    if low <= failures <= high:
        return 0
    return min(abs(failures - low), abs(failures - high))


def substring_matcher(discriminator: str, response: str) -> bool:
    """Placeholder. A prose discriminator needs patterns or a cheap model pass."""
    return discriminator.lower() in response.lower()


def load_responses(path: Path) -> list[Response]:
    responses: list[Response] = []
    for line in path.read_text().splitlines():
        if line.strip():
            responses.append(Response.from_dict(json.loads(line)))
    return responses


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Pick the board from scored candidates.")
    parser.add_argument("--candidates", type=Path, required=True)
    parser.add_argument("--responses", type=Path, required=True, help="jsonl from evalkit.run")
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args(argv)

    raw = yaml.safe_load(args.candidates.read_text()) or {}
    candidates = [Candidate.from_dict(entry) for entry in raw.get("candidates", [])]
    report = run(candidates, load_responses(args.responses), substring_matcher)

    args.out.write_text(
        yaml.safe_dump(
            {"board": [case.to_dict() for case in report.kept]}, sort_keys=False, width=100
        )
    )
    print(report.summary)
    for drop in report.dropped:
        print(f"  dropped {drop.candidate_id} v{drop.variant}: {drop.reason}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
