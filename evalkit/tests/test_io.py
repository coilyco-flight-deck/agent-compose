"""Round-trips through the on-disk formats.

A partially installed pyyaml once let every other gate pass green while both
CLIs raised on import, because nothing exercised the serialization path.
"""

from __future__ import annotations

from pathlib import Path

import yaml

from evalkit import annotate
from evalkit.schema import Annotation, DatasetEntry, Fit, Half, Response, Sample, TestType, Verdict


def test_the_serialization_deps_expose_what_the_clis_call() -> None:
    from inspect_ai.dataset import Sample as InspectSample
    from inspect_ai.log import read_eval_log

    assert hasattr(yaml, "safe_load")
    assert hasattr(yaml, "safe_dump")
    assert callable(read_eval_log)
    assert "metadata" in InspectSample.model_fields


def dataset_entry() -> DatasetEntry:
    return DatasetEntry(
        sample=Sample(
            id="ops-shc-out",
            role="ops",
            test_type=TestType.BOUNDARY,
            prompt="prompt",
            target="target",
            discriminator="drafts the announcement",
            boundary="suggest-human-comms",
            half=Half.OUT,
            pair_id="ops-shc",
        ),
        output="handed it over",
        failure_count=2,
    )


def test_a_dataset_survives_a_write_and_read(tmp_path: Path) -> None:
    path = tmp_path / "dataset.yaml"
    original = dataset_entry()
    path.write_text(yaml.safe_dump({"dataset": [original.to_dict()]}, sort_keys=False))

    loaded = annotate.load_dataset(path)
    assert len(loaded) == 1
    assert loaded[0].sample == original.sample
    assert loaded[0].output == original.output
    assert loaded[0].failure_count == original.failure_count


def test_annotations_survive_a_write_and_read(tmp_path: Path) -> None:
    path = tmp_path / "annotations.yaml"
    annotations = {
        "a": Annotation(id="a", label=Verdict.FAIL, critique="drafted it"),
        "b": Annotation(id="b", label=Fit.UNDECIDED, critique="case was ambiguous"),
        "c": Annotation(id="c", label=Verdict.PASS),
    }
    annotate.save_annotations(path, annotations)

    loaded = annotate.load_annotations(path)
    assert loaded["a"].label is Verdict.FAIL
    assert loaded["a"].critique == "drafted it"
    assert loaded["b"].label is Fit.UNDECIDED
    assert loaded["c"].critique == ""


def test_reasoning_never_counts_against_the_word_cap() -> None:
    response = Response(
        sample_id="ops-shc-out",
        epoch=1,
        text="handed it over",
        reasoning="a long private deliberation that should not count",
    )
    assert response.words == 3


def test_a_sample_round_trips_through_inspect_metadata() -> None:
    original = dataset_entry().sample
    assert Sample.from_inspect(original.to_inspect()) == original
