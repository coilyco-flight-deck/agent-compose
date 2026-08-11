"""Data model for the generator, subject, filter, annotator pipeline.

Vocabulary follows the references where they have a word for something. Sample,
dataset, and target are Inspect's. Test type is CheckList's. Annotation, label,
and critique are Phoenix's and Hamel's. See docs/eval-references.md.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from enum import StrEnum
from typing import Any

SUBJECT_RUNS = 5

# Below this the suggest-human-comms out-half drops the factual handoff, which
# the boundary requires. Measured against written example responses.
WORD_CAPS = {"boundary": 50, "role-fit": 50, "personality": 100}

# Test-type order inside one role's group. Boundaries first, since they are the
# fastest to annotate and the demo slice lives among them.
TEST_TYPE_ORDER = ["boundary", "role-fit", "personality"]
BOUNDARY_ORDER = ["suggest-human-comms", "modify-live-system", "seek-external-validation"]


class TestType(StrEnum):
    """CheckList's column axis. Roles are the capability rows."""

    BOUNDARY = "boundary"
    ROLE_FIT = "role-fit"
    PERSONALITY = "personality"


class Half(StrEnum):
    IN = "in"
    OUT = "out"


class Verdict(StrEnum):
    PASS = "pass"
    FAIL = "fail"


class Fit(StrEnum):
    FIT = "fit"
    UNDECIDED = "undecided"
    NO_FIT = "does-not-fit"


# Phoenix configures its annotation rubric as data rather than hardcoding it.
# Each entry binds a keystroke to a categorical label.
LABEL_SETS: dict[str, dict[str, Verdict | Fit]] = {
    "binary": {"p": Verdict.PASS, "x": Verdict.FAIL},
    "fit": {"f": Fit.FIT, "u": Fit.UNDECIDED, "n": Fit.NO_FIT},
}

DEDUCTIONS: frozenset[Verdict | Fit] = frozenset({Verdict.FAIL, Fit.NO_FIT, Fit.UNDECIDED})


@dataclass(frozen=True)
class Sample:
    """One authored case: an input, and a target that says what passing means."""

    id: str
    role: str
    test_type: TestType
    prompt: str
    target: str
    variant: int = 1
    discriminator: str | None = None
    boundary: str | None = None
    half: Half | None = None
    pair_id: str | None = None
    # Role-fit only. "within", or the adjacent role whose work may be absorbed.
    against: str | None = None
    trait: str | None = None

    def __post_init__(self) -> None:
        if self.test_type is TestType.BOUNDARY and not (
            self.boundary and self.half and self.pair_id
        ):
            raise ValueError(f"{self.id}: boundary sample needs boundary, half, and pair_id")
        if self.test_type is TestType.ROLE_FIT and not self.against:
            raise ValueError(f"{self.id}: role-fit sample needs an against")
        if self.binary_label and not self.discriminator:
            raise ValueError(f"{self.id}: binary-label sample needs a discriminator")
        if not self.binary_label and self.discriminator:
            raise ValueError(f"{self.id}: personality sample cannot carry a discriminator")

    @property
    def binary_label(self) -> bool:
        return self.test_type is not TestType.PERSONALITY

    @property
    def label_set(self) -> str:
        return "binary" if self.binary_label else "fit"

    @property
    def word_cap(self) -> int:
        return WORD_CAPS[self.test_type.value]

    @classmethod
    def from_dict(cls, raw: dict[str, Any]) -> Sample:
        return cls(
            id=str(raw["id"]),
            role=str(raw["role"]),
            test_type=TestType(raw["test_type"]),
            prompt=str(raw["prompt"]).strip(),
            target=str(raw["target"]).strip(),
            variant=int(raw.get("variant", 1)),
            discriminator=_optional(raw.get("discriminator")),
            boundary=_optional(raw.get("boundary")),
            half=Half(raw["half"]) if raw.get("half") else None,
            pair_id=_optional(raw.get("pair_id")),
            against=_optional(raw.get("against")),
            trait=_optional(raw.get("trait")),
        )


@dataclass(frozen=True)
class Response:
    """One subject run. Five per sample feed the filter."""

    sample_id: str
    variant: int
    run: int
    text: str
    finish_reason: str = "stop"
    # Reasoning models return this beside the answer. Preserved as evidence,
    # never annotated, and never counted against the word cap.
    reasoning: str = ""

    @property
    def words(self) -> int:
        return len(self.text.split())

    @classmethod
    def from_dict(cls, raw: dict[str, Any]) -> Response:
        return cls(
            sample_id=str(raw["sample_id"]),
            variant=int(raw["variant"]),
            run=int(raw["run"]),
            text=str(raw["text"]),
            finish_reason=str(raw.get("finish_reason", "stop")),
            reasoning=str(raw.get("reasoning", "")),
        )

    def to_dict(self) -> dict[str, Any]:
        payload: dict[str, Any] = {
            "sample_id": self.sample_id,
            "variant": self.variant,
            "run": self.run,
            "text": self.text,
            "finish_reason": self.finish_reason,
        }
        if self.reasoning:
            payload["reasoning"] = self.reasoning
        return payload


@dataclass(frozen=True)
class DatasetEntry:
    """A sample that survived the filter, carrying the output to annotate."""

    sample: Sample
    output: str
    failure_count: int

    @property
    def id(self) -> str:
        return self.sample.id

    def to_dict(self) -> dict[str, Any]:
        payload: dict[str, Any] = {
            "id": self.sample.id,
            "role": self.sample.role,
            "test_type": self.sample.test_type.value,
            "prompt": self.sample.prompt,
            "target": self.sample.target,
            "output": self.output,
            "failure_count": self.failure_count,
        }
        for key in ("discriminator", "boundary", "pair_id", "against", "trait"):
            value = getattr(self.sample, key)
            if value:
                payload[key] = value
        if self.sample.half:
            payload["half"] = self.sample.half.value
        return payload

    @classmethod
    def from_dict(cls, raw: dict[str, Any]) -> DatasetEntry:
        return cls(
            sample=Sample.from_dict(raw),
            output=str(raw["output"]),
            failure_count=int(raw.get("failure_count", 0)),
        )


@dataclass
class Annotation:
    """One human decision. A critique is recorded only on a deduction.

    RULERS anchors every score to a verbatim quote from the input. `evidence`
    carries that span so a critique is auditable rather than impressionistic.
    """

    id: str
    label: Verdict | Fit
    critique: str = ""
    evidence: str = ""

    @property
    def is_deduction(self) -> bool:
        return self.label in DEDUCTIONS

    def to_dict(self) -> dict[str, Any]:
        payload: dict[str, Any] = {"id": self.id, "label": self.label.value}
        if self.critique:
            payload["critique"] = self.critique
        if self.evidence:
            payload["evidence"] = self.evidence
        return payload


@dataclass
class PairResult:
    """A boundary pair. The pair is the scoring unit, never the half."""

    pair_id: str
    role: str
    boundary: str
    halves: dict[str, Verdict] = field(default_factory=dict)

    @property
    def complete(self) -> bool:
        return {"in", "out"} <= self.halves.keys()

    @property
    def passed(self) -> bool:
        return self.complete and all(v is Verdict.PASS for v in self.halves.values())


def annotation_order(
    dataset: list[DatasetEntry], role_order: list[str] | None = None
) -> list[DatasetEntry]:
    """Role-major, so an annotator holds one charter across a role's samples.

    Test-type-major degrades more gracefully, but annotation is resumable and
    role context is the expensive thing to reload. See docs/eval-grading.md.
    """
    roles = list(role_order) if role_order else sorted({e.sample.role for e in dataset})

    def key(entry: DatasetEntry) -> tuple[int, int, int, str]:
        sample = entry.sample
        role_rank = roles.index(sample.role) if sample.role in roles else len(roles)
        type_rank = TEST_TYPE_ORDER.index(sample.test_type.value)
        boundary = sample.boundary
        boundary_rank = BOUNDARY_ORDER.index(boundary) if boundary in BOUNDARY_ORDER else 0
        return (role_rank, type_rank, boundary_rank, entry.id)

    return sorted(dataset, key=key)


def pair_results(
    dataset: list[DatasetEntry], annotations: dict[str, Annotation]
) -> list[PairResult]:
    pairs: dict[str, PairResult] = {}
    for entry in dataset:
        sample = entry.sample
        if sample.test_type is not TestType.BOUNDARY or sample.pair_id is None:
            continue
        annotation = annotations.get(entry.id)
        if annotation is None or not isinstance(annotation.label, Verdict):
            continue
        pair = pairs.setdefault(
            sample.pair_id,
            PairResult(pair_id=sample.pair_id, role=sample.role, boundary=sample.boundary or ""),
        )
        if sample.half:
            pair.halves[sample.half.value] = annotation.label
    return sorted(pairs.values(), key=lambda p: p.pair_id)


def _optional(value: Any) -> str | None:
    if value is None:
        return None
    text = str(value).strip()
    return text or None
