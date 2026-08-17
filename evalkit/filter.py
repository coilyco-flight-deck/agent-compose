"""Turn an Inspect eval log into the dataset a human annotates.

There is no mechanical scorer here. A regex tier used to select which samples
reached the annotator, and on the first graded board it disagreed with the
grader on every case where either of them deviated from a pass. It measured
something, but not what the grading measures, so it was removed rather than
tuned. See docs/eval-pipeline.md.

The annotator sees epoch 1. The other epochs stay in the Inspect log as
evidence a reader can open.
"""

from __future__ import annotations

import argparse
from collections import defaultdict
from dataclasses import dataclass
from pathlib import Path

import yaml
from inspect_ai.log import read_eval_log

from evalkit.schema import DatasetEntry, Response, Sample


@dataclass(frozen=True)
class Dropped:
    sample_id: str
    reason: str


@dataclass
class DatasetReport:
    """Silent truncation reads as full coverage, so every drop is recorded."""

    kept: list[DatasetEntry]
    dropped: list[Dropped]

    @property
    def summary(self) -> str:
        return f"{len(self.kept)} kept, {len(self.dropped)} dropped"


def run(samples: list[Sample], responses: list[Response]) -> DatasetReport:
    by_sample: dict[str, list[Response]] = defaultdict(list)
    for response in responses:
        by_sample[response.sample_id].append(response)

    kept: list[DatasetEntry] = []
    dropped: list[Dropped] = []
    for sample in samples:
        runs = sorted(by_sample[sample.id], key=lambda r: r.epoch)
        if not runs:
            dropped.append(Dropped(sample.id, "no subject runs"))
            continue
        kept.append(DatasetEntry(sample=sample, output=runs[0].text))
    return DatasetReport(kept=kept, dropped=dropped)


def load_responses(path: Path) -> list[Response]:
    """Read subject outputs from an Inspect eval log, one Response per epoch."""
    log = read_eval_log(str(path))
    if log.status != "success":
        raise SystemExit(f"eval log status is {log.status}, refusing to read it")
    responses: list[Response] = []
    for sample in log.samples or []:
        message = sample.output.choices[0].message if sample.output.choices else None
        responses.append(
            Response(
                sample_id=str(sample.id),
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
    parser = argparse.ArgumentParser(description="Build the annotation dataset from a run.")
    parser.add_argument("--samples", type=Path, required=True)
    parser.add_argument("--log", type=Path, required=True, help=".eval log from inspect eval")
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args(argv)

    samples = load_samples(args.samples)
    report = run(samples, load_responses(args.log))

    args.out.write_text(
        yaml.safe_dump(
            {"dataset": [entry.to_dict() for entry in report.kept]}, sort_keys=False, width=100
        )
    )
    print(report.summary)
    for drop in report.dropped:
        print(f"  dropped {drop.sample_id}: {drop.reason}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
