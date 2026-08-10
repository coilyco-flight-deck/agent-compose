"""Data model for the generator, subject, filter, grader pipeline.

Go owns what a pack is and what a valid record is. This module owns only the
intermediate artifacts that cross between them. See docs/eval-orchestration.md.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from enum import StrEnum
from typing import Any

SUBJECT_RUNS = 5

# Below this the suggest-human-comms out-half drops the factual handoff, which
# the boundary requires. Measured against written example responses.
WORD_CAPS = {"boundary": 50, "role-fit": 50, "personality": 100}

# Ordered so partial grading still leaves every role scored on the same kinds.
KIND_ORDER = ["boundary", "role-fit", "personality"]
BOUNDARY_ORDER = ["suggest-human-comms", "modify-live-system", "seek-external-validation"]


class Kind(StrEnum):
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


@dataclass(frozen=True)
class Candidate:
    """One authored case. Two candidates compete for each pass-or-fail slot."""

    id: str
    role: str
    kind: Kind
    prompt: str
    expected: str
    variant: int = 1
    discriminator: str | None = None
    boundary: str | None = None
    half: Half | None = None
    pair_id: str | None = None
    target: str | None = None
    trait: str | None = None

    def __post_init__(self) -> None:
        if self.kind is Kind.BOUNDARY and not (self.boundary and self.half and self.pair_id):
            raise ValueError(f"{self.id}: boundary case needs boundary, half, and pair_id")
        if self.kind is Kind.ROLE_FIT and not self.target:
            raise ValueError(f"{self.id}: role-fit case needs a target")
        if self.scored_pass_fail and not self.discriminator:
            raise ValueError(f"{self.id}: pass-or-fail case needs a discriminator")
        if not self.scored_pass_fail and self.discriminator:
            raise ValueError(f"{self.id}: personality case cannot carry a discriminator")

    @property
    def scored_pass_fail(self) -> bool:
        return self.kind is not Kind.PERSONALITY

    @property
    def word_cap(self) -> int:
        return WORD_CAPS[self.kind.value]

    @classmethod
    def from_dict(cls, raw: dict[str, Any]) -> Candidate:
        return cls(
            id=str(raw["id"]),
            role=str(raw["role"]),
            kind=Kind(raw["kind"]),
            prompt=str(raw["prompt"]).strip(),
            expected=str(raw["expected"]).strip(),
            variant=int(raw.get("variant", 1)),
            discriminator=_optional(raw.get("discriminator")),
            boundary=_optional(raw.get("boundary")),
            half=Half(raw["half"]) if raw.get("half") else None,
            pair_id=_optional(raw.get("pair_id")),
            target=_optional(raw.get("target")),
            trait=_optional(raw.get("trait")),
        )


@dataclass(frozen=True)
class Response:
    """One subject run. Five per candidate feed the filter."""

    candidate_id: str
    variant: int
    run: int
    text: str
    finish_reason: str = "stop"

    @property
    def words(self) -> int:
        return len(self.text.split())

    @classmethod
    def from_dict(cls, raw: dict[str, Any]) -> Response:
        return cls(
            candidate_id=str(raw["candidate_id"]),
            variant=int(raw["variant"]),
            run=int(raw["run"]),
            text=str(raw["text"]),
            finish_reason=str(raw.get("finish_reason", "stop")),
        )

    def to_dict(self) -> dict[str, Any]:
        return {
            "candidate_id": self.candidate_id,
            "variant": self.variant,
            "run": self.run,
            "text": self.text,
            "finish_reason": self.finish_reason,
        }


@dataclass(frozen=True)
class BoardCase:
    """A candidate that survived the filter, carrying the response to grade."""

    candidate: Candidate
    response: str
    failure_count: int

    @property
    def id(self) -> str:
        return self.candidate.id

    def to_dict(self) -> dict[str, Any]:
        payload: dict[str, Any] = {
            "id": self.candidate.id,
            "role": self.candidate.role,
            "kind": self.candidate.kind.value,
            "prompt": self.candidate.prompt,
            "expected": self.candidate.expected,
            "response": self.response,
            "failure_count": self.failure_count,
        }
        for key in ("discriminator", "boundary", "pair_id", "target", "trait"):
            value = getattr(self.candidate, key)
            if value:
                payload[key] = value
        if self.candidate.half:
            payload["half"] = self.candidate.half.value
        return payload

    @classmethod
    def from_dict(cls, raw: dict[str, Any]) -> BoardCase:
        return cls(
            candidate=Candidate.from_dict(raw),
            response=str(raw["response"]),
            failure_count=int(raw.get("failure_count", 0)),
        )


@dataclass
class Grade:
    """One human decision. Notes are recorded only on a deduction."""

    id: str
    verdict: Verdict | Fit
    note: str = ""

    @property
    def is_deduction(self) -> bool:
        return self.verdict in (Verdict.FAIL, Fit.NO_FIT, Fit.UNDECIDED)

    def to_dict(self) -> dict[str, Any]:
        payload: dict[str, Any] = {"id": self.id, "verdict": self.verdict.value}
        if self.note:
            payload["note"] = self.note
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


def grading_order(cases: list[BoardCase]) -> list[BoardCase]:
    """Kind-major, demo slice first, Content Creator last inside each kind."""

    def key(case: BoardCase) -> tuple[int, int, int, str]:
        kind_rank = KIND_ORDER.index(case.candidate.kind.value)
        boundary = case.candidate.boundary
        boundary_rank = BOUNDARY_ORDER.index(boundary) if boundary in BOUNDARY_ORDER else 0
        creator_last = 1 if case.candidate.role == "creator" else 0
        return (kind_rank, boundary_rank, creator_last, case.id)

    return sorted(cases, key=key)


def pair_results(cases: list[BoardCase], grades: dict[str, Grade]) -> list[PairResult]:
    pairs: dict[str, PairResult] = {}
    for case in cases:
        candidate = case.candidate
        if candidate.kind is not Kind.BOUNDARY or candidate.pair_id is None:
            continue
        grade = grades.get(case.id)
        if grade is None or not isinstance(grade.verdict, Verdict):
            continue
        pair = pairs.setdefault(
            candidate.pair_id,
            PairResult(
                pair_id=candidate.pair_id,
                role=candidate.role,
                boundary=candidate.boundary or "",
            ),
        )
        if candidate.half:
            pair.halves[candidate.half.value] = grade.verdict
    return sorted(pairs.values(), key=lambda p: p.pair_id)


def _optional(value: Any) -> str | None:
    if value is None:
        return None
    text = str(value).strip()
    return text or None
