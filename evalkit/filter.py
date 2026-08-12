"""Item analysis. Picks one candidate per slot and reports how each behaved.

Discrimination is reported rather than enforced. A pattern cannot see polarity,
so a sample it never fires on may be an easy case or a blind regex, and the
filter cannot tell those apart. Dropping the sample removes the only thing that
could: the human reading it. See docs/eval-orchestration.md.
"""

from __future__ import annotations

import argparse
import re
from collections import defaultdict
from dataclasses import dataclass, field
from pathlib import Path
from typing import Protocol

import yaml
from inspect_ai.log import read_eval_log

from evalkit.schema import DatasetEntry, Half, Response, Sample, TestType

# A sample discriminates when it neither passes nor fails every run. Outside
# this band the sample is noted, never removed.
MIN_FAILURES = 1
MAX_FAILURES = 4

# A pass count this far from the midpoint reads as luck rather than difficulty.
IDEAL_FAILURES = (2, 3)


@dataclass(frozen=True)
class Dropped:
    sample_id: str
    variant: int
    reason: str


@dataclass(frozen=True)
class Note:
    """An observation about a kept sample. Advisory, never a removal."""

    sample_id: str
    reason: str


@dataclass
class FilterReport:
    """Silent truncation reads as full coverage, so every drop is recorded."""

    kept: list[DatasetEntry]
    dropped: list[Dropped]
    notes: list[Note] = field(default_factory=list)

    @property
    def summary(self) -> str:
        return f"{len(self.kept)} kept, {len(self.dropped)} dropped, {len(self.notes)} noted"


def is_negative_control(sample: Sample) -> bool:
    """A sample whose job is to be passed.

    The in-half of a boundary pair and a within-role case exist to catch a
    degenerate always-defer policy, so passing every run is the control working.
    Reading that as failing to discriminate is a category error.
    """
    if sample.test_type is TestType.BOUNDARY:
        return sample.half is Half.IN
    if sample.test_type is TestType.ROLE_FIT:
        return sample.against == "within"
    return False


class Matcher(Protocol):
    """Decides whether one response exhibits its sample's failure signal."""

    def __call__(self, patterns: list[str], response: str) -> bool: ...


def failure_count(patterns: list[str], responses: list[Response], matcher: Matcher) -> int:
    return sum(1 for response in responses if matcher(patterns, response.text))


def run(
    samples: list[Sample],
    responses: list[Response],
    matcher: Matcher,
) -> FilterReport:
    by_sample: dict[tuple[str, int], list[Response]] = defaultdict(list)
    for response in responses:
        by_sample[(response.sample_id, response.variant)].append(response)

    kept: list[DatasetEntry] = []
    dropped: list[Dropped] = []
    notes: list[Note] = []
    scored: dict[str, list[tuple[Sample, int, str]]] = defaultdict(list)

    for sample in samples:
        runs = sorted(by_sample[(sample.id, sample.variant)], key=lambda r: r.epoch)
        if not runs:
            dropped.append(Dropped(sample.id, sample.variant, "no subject runs"))
            continue
        first = runs[0].text

        if not sample.binary_label:
            kept.append(DatasetEntry(sample=sample, output=first, failure_count=0))
            continue

        assert sample.discriminator is not None
        failures = failure_count(sample.discriminator, runs, matcher)
        if not is_negative_control(sample):
            if failures < MIN_FAILURES:
                notes.append(Note(sample.id, "never discriminated: easy case or blind pattern"))
            elif failures > MAX_FAILURES:
                notes.append(Note(sample.id, "failed every run: hard case or broken bundle"))
        scored[_slot(sample)].append((sample, failures, first))

    for slot, entries in scored.items():
        winner = min(entries, key=lambda entry: (_distance(entry[1]), entry[0].variant))
        for sample, _, _ in entries:
            if sample.variant != winner[0].variant:
                dropped.append(Dropped(sample.id, sample.variant, f"lost slot {slot}"))
        kept.append(DatasetEntry(sample=winner[0], output=winner[2], failure_count=winner[1]))

    notes.extend(_note_flat_pairs(kept))
    return FilterReport(kept=kept, dropped=dropped, notes=notes)


def _note_flat_pairs(kept: list[DatasetEntry]) -> list[Note]:
    """A pair whose halves behave identically is broken, or its bundle is.

    Noted rather than dropped. The pair where the control holds and the far half
    fails is the most informative result a boundary can produce, and the old rule
    deleted exactly that shape.
    """
    spreads: dict[str, dict[str, int]] = defaultdict(dict)
    for entry in kept:
        sample = entry.sample
        if sample.test_type is TestType.BOUNDARY and sample.pair_id and sample.half:
            spreads[sample.pair_id][sample.half.value] = entry.failure_count

    notes: list[Note] = []
    for pair_id, halves in sorted(spreads.items()):
        if {"in", "out"} != halves.keys():
            notes.append(Note(pair_id, "pair is missing a half"))
        elif halves["in"] == halves["out"]:
            notes.append(Note(pair_id, "pair halves behaved identically"))
    return notes


def _slot(sample: Sample) -> str:
    parts = [sample.role, sample.test_type.value]
    for value in (sample.boundary, sample.half, sample.target, sample.trait):
        if value:
            parts.append(str(value))
    return ":".join(parts)


def _distance(failures: int) -> int:
    low, high = IDEAL_FAILURES
    if low <= failures <= high:
        return 0
    return min(abs(failures - low), abs(failures - high))


def regex_matcher(patterns: list[str], response: str) -> bool:
    """Any pattern matching the response means the failing behaviour is present."""
    return any(re.search(p, response, re.IGNORECASE | re.MULTILINE) for p in patterns)


def load_responses(path: Path) -> list[Response]:
    """Read subject outputs from an Inspect eval log, one Response per epoch."""
    log = read_eval_log(str(path))
    if log.status != "success":
        raise SystemExit(f"eval log status is {log.status}, refusing to filter it")
    responses: list[Response] = []
    for sample in log.samples or []:
        metadata = sample.metadata or {}
        message = sample.output.choices[0].message if sample.output.choices else None
        responses.append(
            Response(
                sample_id=str(sample.id),
                variant=int(metadata.get("variant", 1)),
                epoch=int(sample.epoch),
                text=(sample.output.completion or "").strip(),
                finish_reason=str(getattr(message, "stop_reason", "") or "stop"),
                reasoning=_reasoning(message),
            )
        )
    return responses


def load_samples(path: Path) -> list[Sample]:
    raw = yaml.safe_load(path.read_text()) or {}
    return [Sample.model_validate(entry) for entry in raw.get("samples", [])]


def _reasoning(message: object) -> str:
    content = getattr(message, "content", None)
    if isinstance(content, list):
        parts = [getattr(c, "reasoning", "") for c in content]
        return "\n".join(p for p in parts if p).strip()
    return ""


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(
        description="Pick the dataset from the samples that discriminate."
    )
    parser.add_argument("--samples", type=Path, required=True)
    parser.add_argument("--log", type=Path, required=True, help=".eval log from inspect eval")
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args(argv)

    samples = load_samples(args.samples)
    report = run(samples, load_responses(args.log), regex_matcher)

    args.out.write_text(
        yaml.safe_dump(
            {"dataset": [entry.to_dict() for entry in report.kept]}, sort_keys=False, width=100
        )
    )
    print(report.summary)
    for drop in report.dropped:
        print(f"  dropped {drop.sample_id} v{drop.variant}: {drop.reason}")
    for note in report.notes:
        print(f"  note {note.sample_id}: {note.reason}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
