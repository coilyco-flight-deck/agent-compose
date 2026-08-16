"""Project a committed run into a display payload. One way, never back.

Phoenix is a read-only display surface here: nothing is authored in it, so
nothing returns from it. Committed records stay canonical. See
coilyco-bridge/deploy#572 and docs/eval-export.md.
"""

from __future__ import annotations

import argparse
import json
import re
import sys
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

import yaml

from evalkit.annotate import load_annotations, load_dataset
from evalkit.schema import Annotation, DatasetEntry, pair_results

EXPORT_FORMAT = "agent-compose.eval-export.v1"

# Free text a human wrote for herself while grading. Withheld unless asked for,
# because the display target is a permanent public recording. See the doc.
PRIVATE_FIELDS = ("critique", "evidence")

# A refusal list, not a redaction list: the exporter stops rather than scrubs,
# so a record cannot reach a public surface by having a pattern missed.
SECRET_PATTERNS: tuple[tuple[str, re.Pattern[str]], ...] = (
    ("an AWS key id", re.compile(r"\bAKIA[0-9A-Z]{16}\b")),
    (
        "a bearer or API token",
        re.compile(r"\b(?:ghp|gho|github_pat|xox[baprs])[-_][A-Za-z0-9_]{16,}"),
    ),
    ("a JWT", re.compile(r"\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\b")),
    ("a private key block", re.compile(r"-----BEGIN [A-Z ]*PRIVATE KEY-----")),
    (
        "an SSM parameter path",
        re.compile(r"/[a-z0-9][a-z0-9-]*/[a-z0-9-]*(?:token|password|secret|key)\b", re.I),
    ),
    ("a Discord snowflake", re.compile(r"\b\d{17,20}\b")),
    ("a tailnet host", re.compile(r"\b[a-z0-9-]+\.ts\.net\b", re.I)),
    ("an email address", re.compile(r"\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b")),
)


class ExportRefusedError(Exception):
    """Raised instead of exporting. The caller fixes the record or opts out."""


@dataclass(frozen=True)
class ExportCase:
    """One case as a display surface needs it."""

    id: str
    role: str
    test_type: str
    prompt: str
    target: str
    output: str
    label: str | None = None
    boundary: str | None = None
    half: str | None = None
    pair_id: str | None = None
    against: str | None = None
    trait: str | None = None
    critique: str = ""
    evidence: str = ""

    def to_dict(self) -> dict[str, Any]:
        payload = {
            key: value
            for key, value in self.__dict__.items()
            if value not in (None, "") or key in ("id", "role", "test_type")
        }
        return payload


@dataclass
class ExportRun:
    """A whole run, plus the pair structure that makes a board legible."""

    name: str
    cases: list[ExportCase] = field(default_factory=list)
    pairs: list[dict[str, Any]] = field(default_factory=list)
    includes_private: bool = False

    def to_dict(self) -> dict[str, Any]:
        return {
            "format": EXPORT_FORMAT,
            "run": self.name,
            "includes_private_fields": self.includes_private,
            "counts": self.counts(),
            "pairs": self.pairs,
            "cases": [case.to_dict() for case in self.cases],
        }

    def counts(self) -> dict[str, int]:
        annotated = [case for case in self.cases if case.label]
        return {
            "cases": len(self.cases),
            "annotated": len(annotated),
            "pairs": len(self.pairs),
            "pairs_passed": sum(1 for pair in self.pairs if pair["passed"]),
        }


def scan_for_secrets(text: str) -> list[str]:
    """Every reason a string is unsafe, so one pass reports all of them."""
    return [reason for reason, pattern in SECRET_PATTERNS if pattern.search(text)]


def _label_of(annotation: Annotation | None) -> str | None:
    return annotation.label.value if annotation else None


def build_run(
    name: str,
    dataset: list[DatasetEntry],
    annotations: dict[str, Annotation],
    include_private: bool = False,
) -> ExportRun:
    """Join a dataset to its annotations, refusing anything unsafe to display."""
    run = ExportRun(name=name, includes_private=include_private)
    problems: list[str] = []

    for entry in dataset:
        sample = entry.sample
        annotation = annotations.get(entry.id)
        scanned = {"prompt": sample.prompt, "target": sample.target, "output": entry.output}
        if include_private and annotation:
            scanned["critique"] = annotation.critique
            scanned["evidence"] = annotation.evidence
        for field_name, text in scanned.items():
            for reason in scan_for_secrets(text or ""):
                problems.append(f"{entry.id}.{field_name} contains what looks like {reason}")

        run.cases.append(
            ExportCase(
                id=entry.id,
                role=sample.role,
                test_type=sample.test_type.value,
                prompt=sample.prompt,
                target=sample.target,
                output=entry.output,
                label=_label_of(annotation),
                boundary=sample.boundary,
                half=sample.half.value if sample.half else None,
                pair_id=sample.pair_id,
                against=sample.against,
                trait=sample.trait,
                critique=annotation.critique if (include_private and annotation) else "",
                evidence=annotation.evidence if (include_private and annotation) else "",
            )
        )

    if problems:
        raise ExportRefusedError(
            "refusing to export, because the display target is public:\n  " + "\n  ".join(problems)
        )

    # The pair is the scoring unit, so it travels as its own structure rather
    # than being reconstructed from case rows by whatever renders this.
    for pair in pair_results(dataset, annotations):
        run.pairs.append(
            {
                "pair_id": pair.pair_id,
                "role": pair.role,
                "boundary": pair.boundary,
                "halves": {half: verdict.value for half, verdict in sorted(pair.halves.items())},
                "complete": pair.complete,
                "passed": pair.passed,
            }
        )
    run.pairs.sort(key=lambda item: item["pair_id"])
    return run


def export_run_dir(run_dir: Path, include_private: bool = False) -> ExportRun:
    """Read a committed run directory: dataset.yaml beside annotations.yaml."""
    dataset_path = run_dir / "dataset.yaml"
    if not dataset_path.exists():
        raise ExportRefusedError(f"{run_dir} has no dataset.yaml, so there is nothing to display")
    dataset = load_dataset(dataset_path)
    annotations = load_annotations(run_dir / "annotations.yaml")
    return build_run(run_dir.name, dataset, annotations, include_private)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("run_dir", type=Path, help="a committed run directory")
    parser.add_argument("--out", type=Path, help="write here instead of stdout")
    parser.add_argument(
        "--include-private",
        action="store_true",
        help="also export critique and evidence, which are written for the grader, not an audience",
    )
    parser.add_argument("--format", choices=("json", "yaml"), default="json")
    args = parser.parse_args(argv)

    try:
        run = export_run_dir(args.run_dir, args.include_private)
    except ExportRefusedError as refusal:
        print(f"evalkit.export: {refusal}", file=sys.stderr)
        return 1

    payload = run.to_dict()
    text = (
        json.dumps(payload, indent=2, sort_keys=False)
        if args.format == "json"
        else yaml.safe_dump(payload, sort_keys=False, width=100)
    )
    if args.out:
        args.out.write_text(text + "\n")
        print(f"wrote {args.out} ({run.counts()['cases']} cases, {run.counts()['pairs']} pairs)")
    else:
        print(text)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
