from __future__ import annotations

from typing import Any

from evalkit.matrix import abbreviate, derive
from evalkit.schema import TestType

ROSTER: dict[str, Any] = {
    "role_order": ["engineer", "ops", "creator"],
    "boundary_order": ["suggest-human-comms"],
    "boundaries": {
        "suggest-human-comms": {
            "owner": "creator",
            "summary": "Content Creator recommends communication, other roles keep records",
        }
    },
    "roles": {
        "engineer": {
            "display_name": "Engineer",
            "purpose": "Build and land work across the portfolio.",
            "boundaries": ["suggest-human-comms"],
            "personalities": ["curious", "meticulous"],
            "adjacents": [{"role": "ops", "reason": "deploying instead of handing back"}],
        },
        "ops": {
            "display_name": "DevOps",
            "purpose": "Operate the real hosted systems.",
            "boundaries": ["suggest-human-comms"],
            "personalities": ["grounded"],
            "adjacents": [{"role": "engineer", "reason": "implementing the fix"}],
        },
        "creator": {
            "display_name": "Content Creator",
            "purpose": "Turn work into content.",
            "personalities": ["warm"],
            "adjacents": [],
        },
    },
}


def test_the_owner_gets_a_pair_alongside_every_deferring_role() -> None:
    boundary = [slot for slot in derive(ROSTER) if slot.test_type is TestType.BOUNDARY]
    assert [slot.id for slot in boundary] == [
        "engineer-shc-in",
        "engineer-shc-out",
        "ops-shc-in",
        "ops-shc-out",
        "creator-shc-in",
        "creator-shc-out",
    ]


def test_a_role_declaring_no_boundary_still_owns_one_as_its_owner() -> None:
    creator = {
        slot.id: slot.descriptor
        for slot in derive(ROSTER)
        if slot.role == "creator" and slot.test_type is TestType.BOUNDARY
    }
    assert creator["creator-shc-in"] == 'owns "recommends communication"'
    assert creator["creator-shc-out"] == 'owns "recommends communication", claims nothing past it'


def test_a_deferring_half_names_the_owner_behaviour_and_an_owning_half_names_purpose() -> None:
    slots = {slot.id: slot.descriptor for slot in derive(ROSTER)}
    assert slots["ops-shc-out"] == 'defers "recommends communication" to creator'
    assert slots["ops-shc-in"] == "owns: operate the real hosted systems"


def test_the_owner_clause_drops_the_display_name_that_conjugates_it() -> None:
    from evalkit.matrix import owner_behaviour

    summary = "Executive Strategist reaches outside the local frame, other roles do not"
    roster = {"roles": {"exec": {"display_name": "Executive Strategist"}}}
    assert owner_behaviour(summary, "exec", roster) == "reaches outside the local frame"


def test_a_summary_without_the_expected_shape_survives_intact() -> None:
    from evalkit.matrix import owner_behaviour

    assert owner_behaviour("one clause only", "exec", {"roles": {}}) == "one clause only"


def test_adjacency_reasons_become_the_case_descriptors() -> None:
    slots = derive(ROSTER)
    fit = {s.id: s.descriptor for s in slots if s.test_type is TestType.ROLE_FIT}
    assert fit["engineer-fit-ops"] == "deploying instead of handing back"
    assert fit["engineer-fit-within"] == "engineer correctly identifies work it should own"


def test_personality_is_one_case_per_trait_with_no_composed_case() -> None:
    personality = [slot for slot in derive(ROSTER) if slot.test_type is TestType.PERSONALITY]
    engineer = [slot.id for slot in personality if slot.role == "engineer"]
    assert engineer == ["engineer-per-curious", "engineer-per-meticulous"]


def test_a_trait_case_names_the_peers_it_is_composed_alongside() -> None:
    slots = {slot.id: slot.descriptor for slot in derive(ROSTER)}
    assert slots["engineer-per-curious"] == "curious, composed alongside meticulous"
    assert slots["ops-per-grounded"] == "grounded"


def test_board_size_is_a_consequence_of_the_roster() -> None:
    slots = derive(ROSTER)
    counts = {kind: sum(1 for slot in slots if slot.test_type is kind) for kind in TestType}
    assert counts[TestType.BOUNDARY] == 6
    assert counts[TestType.ROLE_FIT] == 5
    assert counts[TestType.PERSONALITY] == 4


def test_boundary_abbreviations_come_from_the_slug() -> None:
    assert abbreviate("suggest-human-comms") == "shc"
    assert abbreviate("seek-external-validation") == "sev"
    assert abbreviate("modify-live-system") == "mls"
