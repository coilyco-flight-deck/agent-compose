from __future__ import annotations

from pathlib import Path

import pytest
import yaml

from evalkit.export import ExportRefusedError, build_run, export_run_dir, scan_for_secrets
from evalkit.schema import Annotation, DatasetEntry, Half, Sample, TestType, Verdict


def boundary_entry(pair: str, half: Half, output: str = "a reply") -> DatasetEntry:
    return DatasetEntry(
        sample=Sample(
            id=f"{pair}-{half.value}",
            role="ops",
            test_type=TestType.BOUNDARY,
            prompt="a prompt",
            target="a target",
            boundary="modify-live-system",
            half=half,
            pair_id=pair,
        ),
        output=output,
    )


def graded(pair: str, half: Half, label: Verdict, **extra: str) -> Annotation:
    return Annotation(id=f"{pair}-{half.value}", label=label, **extra)


def annotated_pair(**extra: str) -> tuple[list[DatasetEntry], dict[str, Annotation]]:
    dataset = [boundary_entry("p", Half.IN), boundary_entry("p", Half.OUT)]
    annotations = {
        "p-in": graded("p", Half.IN, Verdict.FAIL, **extra),
        "p-out": graded("p", Half.OUT, Verdict.PASS),
    }
    return dataset, annotations


def test_a_default_export_withholds_the_graders_own_notes() -> None:
    dataset, annotations = annotated_pair(critique="my scratch note", evidence="a span")
    run = build_run("run", dataset, annotations)
    rendered = yaml.safe_dump(run.to_dict())
    assert "my scratch note" not in rendered
    assert "a span" not in rendered
    assert run.to_dict()["includes_private_fields"] is False


def test_withholding_the_prose_keeps_the_labels() -> None:
    # The surface exists for the pass-or-fail shape, which has to survive
    # dropping the prose around it.
    dataset, annotations = annotated_pair(critique="my scratch note")
    run = build_run("run", dataset, annotations)
    assert {c.id: c.label for c in run.cases} == {"p-in": "fail", "p-out": "pass"}


def test_opting_in_carries_the_notes_and_records_that_it_did() -> None:
    dataset, annotations = annotated_pair(critique="my scratch note")
    run = build_run("run", dataset, annotations, include_private=True)
    assert "my scratch note" in yaml.safe_dump(run.to_dict())
    assert run.to_dict()["includes_private_fields"] is True


@pytest.mark.parametrize(
    "text",
    [
        "AKIAIOSFODNN7EXAMPLE",
        "ghp_abcdefghijklmnopqrstuvwxyz0123",
        "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dBjftJeZ4CVPmB92K27uhbUJU1p1r",
        "-----BEGIN RSA PRIVATE KEY-----",
        "/sirens-echo/postgres-password",
        "318190481467244544",
        "kai-server.tailfoo.ts.net",
        "someone@example.com",
    ],
)
def test_each_unsafe_shape_is_recognized(text: str) -> None:
    assert scan_for_secrets(text)


def test_ordinary_evaluation_prose_is_not_flagged() -> None:
    # A gate that cries wolf on real records gets turned off.
    assert not scan_for_secrets(
        "Writes the factual incident record itself, keeping to before-state"
        " and after-state evidence."
    )


def test_an_unsafe_output_refuses_the_whole_run() -> None:
    dataset = [boundary_entry("p", Half.IN, output="ping 318190481467244544")]
    with pytest.raises(ExportRefusedError, match="snowflake"):
        build_run("run", dataset, {})


def test_one_pass_reports_every_reason() -> None:
    # Reporting one problem per run turns a fix into a loop.
    dataset = [boundary_entry("p", Half.IN, output="318190481467244544 at host.ts.net")]
    with pytest.raises(ExportRefusedError) as refusal:
        build_run("run", dataset, {})
    assert "snowflake" in str(refusal.value)
    assert "tailnet" in str(refusal.value)


def test_a_withheld_note_is_not_scanned() -> None:
    # Text that never leaves cannot leak, so refusing on it would block an
    # export that is actually safe.
    dataset = [boundary_entry("p", Half.IN)]
    annotations = {"p-in": graded("p", Half.IN, Verdict.FAIL, critique="see 318190481467244544")}
    build_run("run", dataset, annotations)
    with pytest.raises(ExportRefusedError):
        build_run("run", dataset, annotations, include_private=True)


def test_a_passing_pair_travels_as_one_unit() -> None:
    dataset = [boundary_entry("p", Half.IN), boundary_entry("p", Half.OUT)]
    annotations = {
        "p-in": graded("p", Half.IN, Verdict.PASS),
        "p-out": graded("p", Half.OUT, Verdict.PASS),
    }
    run = build_run("run", dataset, annotations)
    assert len(run.pairs) == 1
    assert run.pairs[0]["passed"] is True
    assert run.counts()["pairs_passed"] == 1


def test_one_failing_half_fails_the_pair() -> None:
    dataset, annotations = annotated_pair()
    run = build_run("run", dataset, annotations)
    assert run.pairs[0]["passed"] is False


def test_a_half_graded_pair_does_not_read_as_a_pass() -> None:
    dataset = [boundary_entry("p", Half.IN), boundary_entry("p", Half.OUT)]
    run = build_run("run", dataset, {"p-in": graded("p", Half.IN, Verdict.PASS)})
    assert run.pairs[0]["complete"] is False
    assert run.pairs[0]["passed"] is False


def test_a_missing_dataset_refuses(tmp_path: Path) -> None:
    with pytest.raises(ExportRefusedError, match=r"no dataset\.yaml"):
        export_run_dir(tmp_path)


def test_a_committed_run_directory_loads(tmp_path: Path) -> None:
    entry = boundary_entry("p", Half.IN)
    (tmp_path / "dataset.yaml").write_text(yaml.safe_dump({"dataset": [entry.to_dict()]}))
    (tmp_path / "annotations.yaml").write_text(
        yaml.safe_dump({"annotations": [{"id": "p-in", "label": "pass"}]})
    )
    run = export_run_dir(tmp_path)
    assert run.name == tmp_path.name
    assert run.counts() == {"cases": 1, "annotated": 1, "pairs": 1, "pairs_passed": 0}
