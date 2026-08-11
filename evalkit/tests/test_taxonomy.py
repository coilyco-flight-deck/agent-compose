from __future__ import annotations

from evalkit.schema import Annotation, DatasetEntry, Fit, Half, Sample, TestType, Verdict
from evalkit.taxonomy import axis_of, build, salient_terms


def boundary_entry(role: str, half: Half) -> DatasetEntry:
    return DatasetEntry(
        sample=Sample(
            id=f"{role}-shc-{half.value}",
            role=role,
            test_type=TestType.BOUNDARY,
            prompt="p",
            target="t",
            discriminator=[r"d"],
            boundary="suggest-human-comms",
            half=half,
            pair_id=f"{role}-shc",
        ),
        output="drafted the notice",
        failure_count=2,
    )


def personality_entry(role: str, trait: str) -> DatasetEntry:
    return DatasetEntry(
        sample=Sample(
            id=f"{role}-per-{trait}",
            role=role,
            test_type=TestType.PERSONALITY,
            prompt="p",
            target="t",
            trait=trait,
        ),
        output="flat",
        failure_count=0,
    )


def test_only_deductions_enter_the_taxonomy() -> None:
    dataset = [boundary_entry("ops", Half.OUT), boundary_entry("design", Half.OUT)]
    annotations = {
        "ops-shc-out": Annotation(id="ops-shc-out", label=Verdict.FAIL, critique="drafted wording"),
        "design-shc-out": Annotation(id="design-shc-out", label=Verdict.PASS),
    }
    modes = build(dataset, annotations)
    assert sum(m.count for m in modes) == 1
    assert modes[0].sample_ids == ["ops-shc-out"]


def test_shared_critique_terms_on_one_axis_collapse_into_one_mode() -> None:
    dataset = [boundary_entry("ops", Half.OUT), boundary_entry("design", Half.OUT)]
    annotations = {
        "ops-shc-out": Annotation(id="ops-shc-out", label=Verdict.FAIL, critique="drafted wording"),
        "design-shc-out": Annotation(
            id="design-shc-out", label=Verdict.FAIL, critique="drafted wording"
        ),
    }
    modes = build(dataset, annotations)
    assert len(modes) == 1
    assert modes[0].count == 2
    assert dict(modes[0].roles) == {"ops": 1, "design": 1}


def test_the_taxonomy_ranks_by_frequency() -> None:
    dataset = [
        boundary_entry("ops", Half.OUT),
        boundary_entry("design", Half.OUT),
        personality_entry("qa", "candid"),
    ]
    annotations = {
        "ops-shc-out": Annotation(id="ops-shc-out", label=Verdict.FAIL, critique="drafted wording"),
        "design-shc-out": Annotation(
            id="design-shc-out", label=Verdict.FAIL, critique="drafted wording"
        ),
        "qa-per-candid": Annotation(id="qa-per-candid", label=Fit.NO_FIT, critique="hedged"),
    }
    modes = build(dataset, annotations)
    assert modes[0].count == 2
    assert modes[1].count == 1


def test_undecided_is_a_deduction_so_bad_cases_surface_as_a_mode() -> None:
    dataset = [personality_entry("qa", "candid")]
    annotations = {
        "qa-per-candid": Annotation(
            id="qa-per-candid", label=Fit.UNDECIDED, critique="prompt was ambiguous"
        )
    }
    modes = build(dataset, annotations)
    assert modes[0].count == 1
    assert "personality:candid" in modes[0].key


def test_the_axis_comes_from_structure_before_prose() -> None:
    assert axis_of(boundary_entry("ops", Half.IN)) == "suggest-human-comms:in"
    assert axis_of(personality_entry("qa", "candid")) == "personality:candid"


def test_stopwords_do_not_become_failure_mode_names() -> None:
    assert salient_terms("it was not the role that did this") == []
    assert "deployed" in salient_terms("deployed the change deployed again")


def test_evidence_rides_along_with_its_mode() -> None:
    dataset = [boundary_entry("ops", Half.OUT)]
    annotations = {
        "ops-shc-out": Annotation(
            id="ops-shc-out",
            label=Verdict.FAIL,
            critique="drafted wording",
            evidence="drafted the notice",
        )
    }
    assert build(dataset, annotations)[0].evidence == ["drafted the notice"]
