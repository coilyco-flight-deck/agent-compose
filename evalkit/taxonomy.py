"""Axial coding. Groups critiques into a ranked failure taxonomy.

Both practitioner references describe error analysis as open coding, then axial
coding, then a taxonomy. Critiques are the open codes. This is the axial step,
and the taxonomy is what you act on. See docs/eval-grading.md.
"""

from __future__ import annotations

import argparse
import re
from collections import Counter
from dataclasses import dataclass, field
from pathlib import Path

import yaml

from evalkit.annotate import load_annotations, load_dataset
from evalkit.schema import Annotation, DatasetEntry

# Open codes are prose. Grouping them needs a shared vocabulary, so the
# taxonomy keys off the axes the dataset already carries.
STOPWORDS = frozenset(
    [
        "a",
        "an",
        "the",
        "it",
        "its",
        "is",
        "was",
        "be",
        "been",
        "are",
        "and",
        "or",
        "but",
        "of",
        "to",
        "in",
        "on",
        "for",
        "with",
        "without",
        "that",
        "this",
        "these",
        "those",
        "then",
        "than",
        "as",
        "at",
        "by",
        "from",
        "into",
        "over",
        "under",
        "not",
        "no",
        "non",
        "do",
        "does",
        "did",
        "done",
        "have",
        "has",
        "had",
        "you",
        "your",
        "they",
        "their",
        "them",
        "role",
        "instead",
        "rather",
    ]
)


@dataclass
class FailureMode:
    """One axial category: a recurring failure with the samples that show it."""

    key: str
    sample_ids: list[str] = field(default_factory=list)
    roles: Counter[str] = field(default_factory=Counter)
    test_types: Counter[str] = field(default_factory=Counter)
    evidence: list[str] = field(default_factory=list)

    @property
    def count(self) -> int:
        return len(self.sample_ids)

    def to_dict(self) -> dict[str, object]:
        payload: dict[str, object] = {
            "failure_mode": self.key,
            "count": self.count,
            "roles": dict(self.roles.most_common()),
            "test_types": dict(self.test_types.most_common()),
            "samples": sorted(self.sample_ids),
        }
        if self.evidence:
            payload["evidence"] = self.evidence[:3]
        return payload


def axis_of(entry: DatasetEntry) -> str:
    """The structural axis a failure sits on, before its prose is considered."""
    sample = entry.sample
    if sample.boundary and sample.half:
        return f"{sample.boundary}:{sample.half.value}"
    if sample.against:
        return f"role-fit:{sample.against}"
    if sample.trait:
        return f"personality:{sample.trait}"
    return sample.test_type.value


def salient_terms(critique: str, limit: int = 3) -> list[str]:
    words = [w for w in re.findall(r"[a-z][a-z-]{2,}", critique.lower()) if w not in STOPWORDS]
    return [word for word, _ in Counter(words).most_common(limit)]


def build(dataset: list[DatasetEntry], annotations: dict[str, Annotation]) -> list[FailureMode]:
    """Group every deduction by structural axis, then by shared critique terms."""
    by_id = {entry.id: entry for entry in dataset}
    modes: dict[str, FailureMode] = {}

    for annotation in annotations.values():
        if not annotation.is_deduction:
            continue
        entry = by_id.get(annotation.id)
        if entry is None:
            continue
        terms = salient_terms(annotation.critique)
        key = axis_of(entry)
        if terms:
            key = f"{key} / {' '.join(sorted(terms))}"
        mode = modes.setdefault(key, FailureMode(key=key))
        mode.sample_ids.append(annotation.id)
        mode.roles[entry.sample.role] += 1
        mode.test_types[entry.sample.test_type.value] += 1
        if annotation.evidence:
            mode.evidence.append(annotation.evidence)

    return sorted(modes.values(), key=lambda m: (-m.count, m.key))


def render(modes: list[FailureMode], total: int) -> str:
    if not modes:
        return "no deductions recorded, so there is no taxonomy to build"
    lines = [f"{sum(m.count for m in modes)} deductions across {total} samples", ""]
    for index, mode in enumerate(modes, start=1):
        roles = ", ".join(f"{role} x{n}" for role, n in mode.roles.most_common())
        lines.append(f"{index:2d}. [{mode.count}] {mode.key}")
        lines.append(f"      roles: {roles}")
        lines.append(f"      samples: {', '.join(sorted(mode.sample_ids))}")
    return "\n".join(lines)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Cluster critiques into a failure taxonomy.")
    parser.add_argument("--dataset", type=Path, required=True)
    parser.add_argument("--annotations", type=Path, required=True)
    parser.add_argument("--format", choices=("text", "yaml"), default="text")
    parser.add_argument("--out", type=Path, help="write instead of printing")
    args = parser.parse_args(argv)

    dataset = load_dataset(args.dataset)
    modes = build(dataset, load_annotations(args.annotations))

    if args.format == "yaml":
        rendered = yaml.safe_dump(
            {"failure_taxonomy": [m.to_dict() for m in modes]}, sort_keys=False, width=100
        )
    else:
        rendered = render(modes, len(dataset))

    if args.out:
        args.out.write_text(rendered if rendered.endswith("\n") else rendered + "\n")
    else:
        print(rendered)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
