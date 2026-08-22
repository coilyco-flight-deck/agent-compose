from __future__ import annotations

from pathlib import Path

import pytest
from aos_eval.schema import Half, Response, Sample

from evalkit import filter as dataset_builder


def role_fit(against: str) -> Sample:
    return Sample(
        id=f"sysadmin-fit-{against}",
        role="sysadmin",
        test_type="role-fit",
        prompt="prompt",
        target="hands it back",
        against=against,
    )


def responses(sample: Sample, total: int = 5) -> list[Response]:
    return [
        Response(sample_id=sample.id, epoch=index, text=f"answer {index}")
        for index in range(1, total + 1)
    ]


def test_the_annotator_sees_epoch_one() -> None:
    sample = role_fit("director")
    report = dataset_builder.run([sample], responses(sample))
    assert [entry.id for entry in report.kept] == [sample.id]
    assert report.kept[0].output == "answer 1"


def test_epoch_order_does_not_depend_on_log_order() -> None:
    sample = role_fit("director")
    shuffled = list(reversed(responses(sample)))
    report = dataset_builder.run([sample], shuffled)
    assert report.kept[0].output == "answer 1"


# Nothing is scored mechanically, so every authored case reaches the annotator.
def test_every_authored_sample_survives() -> None:
    samples = [
        role_fit("within"),
        role_fit("platform"),
        Sample(
            id="sysadmin-sec-out",
            role="sysadmin",
            test_type="boundary",
            prompt="prompt",
            target="defers wording",
            boundary="suggest-external-comms",
            half=Half.OUT,
            pair_id="sysadmin-sec",
        ),
        Sample(
            id="sysadmin-per-grounded",
            role="sysadmin",
            test_type="personality",
            prompt="prompt",
            target="stays plain",
            trait="grounded",
        ),
    ]
    runs = [r for sample in samples for r in responses(sample)]
    report = dataset_builder.run(samples, runs)
    assert len(report.kept) == len(samples)
    assert report.dropped == []


def test_a_sample_with_no_runs_is_reported_so_truncation_is_never_silent() -> None:
    sample = role_fit("director")
    report = dataset_builder.run([sample], [])
    assert report.dropped[0].reason == "no subject runs"
    assert report.summary == "0 kept, 1 dropped"


def test_a_boundary_sample_still_needs_its_pair_identity(tmp_path: Path) -> None:
    path = tmp_path / "samples.yaml"
    path.write_text(
        "samples:\n"
        "  - id: sysadmin-sec-out\n"
        "    role: sysadmin\n"
        "    test_type: boundary\n"
        "    prompt: p\n"
        "    target: t\n"
    )
    with pytest.raises(ValueError, match="boundary sample needs"):
        dataset_builder.load_samples(path)


def test_a_role_fit_sample_still_needs_an_against(tmp_path: Path) -> None:
    path = tmp_path / "samples.yaml"
    path.write_text(
        "samples:\n"
        "  - id: sysadmin-fit-within\n"
        "    role: sysadmin\n"
        "    test_type: role-fit\n"
        "    prompt: p\n"
        "    target: t\n"
    )
    with pytest.raises(ValueError, match="needs against"):
        dataset_builder.load_samples(path)


def test_a_well_formed_board_loads(tmp_path: Path) -> None:
    path = tmp_path / "samples.yaml"
    path.write_text(
        "samples:\n"
        "  - id: sysadmin-mlb-in\n"
        "    role: sysadmin\n"
        "    test_type: boundary\n"
        "    prompt: p\n"
        "    target: t\n"
        "    boundary: modify-live-backend\n"
        "    half: in\n"
        "    pair_id: sysadmin-mlb\n"
    )
    assert [s.id for s in dataset_builder.load_samples(path)] == ["sysadmin-mlb-in"]
