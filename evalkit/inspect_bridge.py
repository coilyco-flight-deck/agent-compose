"""Translate the shared schema to and from Inspect's Sample.

This is the seam the shared layer deliberately does not cross. `aos_eval`
carries no runner and no model client, so the mapping onto Inspect's carrier
lives here, in the repo whose runner is Inspect.
"""

from __future__ import annotations

from enum import StrEnum

from aos_eval.schema import Sample
from inspect_ai.dataset import Sample as InspectSample

# Inspect's Sample carries only input, target, id, and metadata, so every
# domain field rides in metadata.
METADATA_FIELDS = (
    "role",
    "test_type",
    "boundary",
    "half",
    "pair_id",
    "against",
    "trait",
)


def to_inspect(sample: Sample) -> InspectSample:
    metadata = {
        key: (value.value if isinstance(value, StrEnum) else value)
        for key, value in ((name, getattr(sample, name)) for name in METADATA_FIELDS)
        if value is not None
    }
    return InspectSample(
        id=sample.id,
        input=sample.prompt,
        target=sample.target,
        metadata=metadata,
    )


def from_inspect(sample: InspectSample) -> Sample:
    metadata = dict(sample.metadata or {})
    return Sample(
        id=str(sample.id),
        prompt=str(sample.input),
        target=str(sample.target),
        **metadata,
    )
