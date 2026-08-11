"""Item analysis. Keeps samples that discriminate and drops the rest.

A sample every subject run passes measures nothing, and one every run fails
is broken rather than hard. See docs/eval-orchestration.md.
"""

from __future__ import annotations

import argparse
from collections import defaultdict
from dataclasses import dataclass
from pathlib import Path
from typing import Protocol

import yaml
from inspect_ai.log import read_eval_log

from evalkit.schema import DatasetEntry, Response, Sample, TestType

# A sample discriminates when it neither passes nor fails every run.
MIN_FAILURES = 1
MAX_FAILURES = 4

# A pass count this far from the midpoint reads as luck rather than difficulty.
IDEAL_FAILURES = (2, 3)


@dataclass(frozen=True)
class Dropped:
    sample_id: str
    variant: int
    reason: str


@dataclass
class FilterReport:
    """Silent truncation reads as full coverage, so every drop is recorded."""

    kept: list[DatasetEntry]
    dropped: list[Dropped]

    @property
    def summary(self) -> str:
        return f"{len(self.kept)} kept, {len(self.dropped)} dropped"


class Matcher(Protocol):
    """Decides whether one response exhibits its sample's failure signal."""

    def __call__(self, discriminator: str, response: str) -> bool: ...


def failure_count(discriminator: str, responses: list[Response], matcher: Matcher) -> int:
    return sum(1 for response in responses if matcher(discriminator, response.text))


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
        if failures < MIN_FAILURES:
            dropped.append(Dropped(sample.id, sample.variant, "every run passed"))
            continue
        if failures > MAX_FAILURES:
            dropped.append(Dropped(sample.id, sample.variant, "every run failed"))
            continue
        scored[_slot(sample)].append((sample, failures, first))

    for slot, entries in scored.items():
        winner = min(entries, key=lambda entry: _distance(entry[1]))
        for sample, _, _ in entries:
            if sample.variant != winner[0].variant:
                dropped.append(Dropped(sample.id, sample.variant, f"lost slot {slot}"))
        kept.append(DatasetEntry(sample=winner[0], output=winner[2], failure_count=winner[1]))

    return FilterReport(kept=_drop_broken_pairs(kept, dropped), dropped=dropped)


def _drop_broken_pairs(kept: list[DatasetEntry], dropped: list[Dropped]) -> list[DatasetEntry]:
    """A pair whose halves behave identically is broken, or its bundle is."""
    halves: dict[str, set[str]] = defaultdict(set)
    for case in kept:
        sample = case.sample
        if sample.test_type is TestType.BOUNDARY and sample.pair_id and sample.half:
            halves[sample.pair_id].add(sample.half.value)

    survivors: list[DatasetEntry] = []
    for entry in kept:
        sample = entry.sample
        broken = halves[sample.pair_id] != {"in", "out"} if sample.pair_id else False
        if sample.test_type is TestType.BOUNDARY and sample.pair_id and broken:
            dropped.append(Dropped(sample.id, sample.variant, f"pair {sample.pair_id} incomplete"))
            continue
        survivors.append(entry)
    return survivors


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


def substring_matcher(discriminator: str, response: str) -> bool:
    """Placeholder. A prose discriminator needs patterns or a cheap model pass."""
    return discriminator.lower() in response.lower()


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
    report = run(samples, load_responses(args.log), substring_matcher)

    args.out.write_text(
        yaml.safe_dump(
            {"dataset": [entry.to_dict() for entry in report.kept]}, sort_keys=False, width=100
        )
    )
    print(report.summary)
    for drop in report.dropped:
        print(f"  dropped {drop.sample_id} v{drop.variant}: {drop.reason}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
