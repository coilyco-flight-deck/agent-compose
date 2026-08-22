from __future__ import annotations

from pathlib import Path
from typing import Any

import yaml

from evalkit.matrix import BOUNDARY, PERSONALITY, ROLE_FIT, abbreviate, derive

ROSTER: dict[str, Any] = {
    "role_order": ["platform", "sysadmin", "devrel"],
    "boundary_order": ["suggest-external-comms"],
    "boundaries": {
        "suggest-external-comms": {
            "owner": "devrel",
            "summary": "Developer Advocate recommends communication, other roles keep records",
        }
    },
    "roles": {
        "platform": {
            "display_name": "Developer Platform Engineer",
            "purpose": "Build and land work across the portfolio.",
            "boundaries": ["suggest-external-comms"],
            "personalities": ["tenacious", "grounded"],
            "adjacents": [{"role": "sysadmin", "reason": "deploying instead of handing back"}],
        },
        "sysadmin": {
            "display_name": "Systems Administrator",
            "purpose": "Operate the real hosted systems.",
            "boundaries": ["suggest-external-comms"],
            "personalities": ["grounded"],
            "adjacents": [{"role": "platform", "reason": "implementing the fix"}],
        },
        "devrel": {
            "display_name": "Developer Advocate",
            "purpose": "Turn work into content.",
            "personalities": ["warm"],
            "adjacents": [],
        },
    },
}

SCOPED: dict[str, Any] = {
    "role_order": ["gamedev", "platform", "sysadmin"],
    "boundary_order": ["modify-live-backend"],
    "boundaries": {
        "modify-live-backend": {
            "owner": "sysadmin",
            "summary": "Systems Administrator changes running systems, others hand it over",
        }
    },
    "roles": {
        "gamedev": {
            "display_name": "Game Developer",
            "purpose": "Ship playable games.",
            "scoped_boundaries": [
                {"name": "modify-live-backend", "scope": "a local world you run yourself"}
            ],
            "personalities": ["immersed"],
            "adjacents": [],
        },
        "platform": {
            "display_name": "Developer Platform Engineer",
            "purpose": "Build and land work.",
            "boundaries": ["modify-live-backend"],
            "personalities": ["tenacious"],
            "adjacents": [],
        },
        "sysadmin": {
            "display_name": "Systems Administrator",
            "purpose": "Operate the real hosted systems.",
            "personalities": ["protective"],
            "adjacents": [],
        },
    },
}


def test_the_owner_gets_a_pair_alongside_every_deferring_role() -> None:
    boundary = [slot for slot in derive(ROSTER) if slot.test_type == BOUNDARY]
    assert [slot.id for slot in boundary] == [
        "platform-sec-in",
        "platform-sec-out",
        "sysadmin-sec-in",
        "sysadmin-sec-out",
        "devrel-sec-in",
        "devrel-sec-out",
    ]


def test_a_role_declaring_no_boundary_still_owns_one_as_its_owner() -> None:
    devrel = {
        slot.id: slot.descriptor
        for slot in derive(ROSTER)
        if slot.role == "devrel" and slot.test_type == BOUNDARY
    }
    assert devrel["devrel-sec-in"] == 'owns "recommends communication"'
    assert devrel["devrel-sec-out"] == 'owns "recommends communication", claims nothing past it'


def test_a_deferring_half_names_the_owner_behaviour_and_an_owning_half_names_purpose() -> None:
    slots = {slot.id: slot.descriptor for slot in derive(ROSTER)}
    assert slots["sysadmin-sec-out"] == 'defers "recommends communication" to devrel'
    assert slots["sysadmin-sec-in"] == "owns: operate the real hosted systems"


def test_the_owner_clause_drsysadmin_the_display_name_that_conjugates_it() -> None:
    from evalkit.matrix import owner_behaviour

    summary = "Executive Strategist reaches outside the local frame, other roles do not"
    roster = {"roles": {"exec": {"display_name": "Executive Strategist"}}}
    assert owner_behaviour(summary, "exec", roster) == "reaches outside the local frame"


def test_a_summary_without_the_expected_shape_survives_intact() -> None:
    from evalkit.matrix import owner_behaviour

    assert owner_behaviour("one clause only", "exec", {"roles": {}}) == "one clause only"


def test_adjacency_reasons_become_the_case_descriptors() -> None:
    slots = derive(ROSTER)
    fit = {s.id: s.descriptor for s in slots if s.test_type == ROLE_FIT}
    assert fit["platform-fit-sysadmin"] == "deploying instead of handing back"
    assert fit["platform-fit-within"] == "platform correctly identifies work it should own"


def test_personality_is_one_case_per_trait_with_no_composed_case() -> None:
    personality = [slot for slot in derive(ROSTER) if slot.test_type == PERSONALITY]
    platform = [slot.id for slot in personality if slot.role == "platform"]
    assert platform == ["platform-per-tenacious", "platform-per-grounded"]


def test_a_trait_case_names_the_peers_it_is_composed_alongside() -> None:
    slots = {slot.id: slot.descriptor for slot in derive(ROSTER)}
    assert slots["platform-per-tenacious"] == "tenacious, composed alongside grounded"
    assert slots["sysadmin-per-grounded"] == "grounded"


def test_board_size_is_a_consequence_of_the_roster() -> None:
    slots = derive(ROSTER)
    counts = {
        kind: sum(1 for slot in slots if slot.test_type == kind)
        for kind in (BOUNDARY, ROLE_FIT, PERSONALITY)
    }
    assert counts[BOUNDARY] == 6
    assert counts[ROLE_FIT] == 5
    assert counts[PERSONALITY] == 4


def test_boundary_abbreviations_come_from_the_slug() -> None:
    assert abbreviate("suggest-external-comms") == "sec"
    assert abbreviate("seek-external-validation") == "sev"
    assert abbreviate("modify-live-backend") == "mlb"


def test_a_scoped_grant_earns_its_own_pair_between_the_deferrers_and_the_owner() -> None:
    boundary = [slot for slot in derive(SCOPED, group="tier") if slot.test_type == BOUNDARY]
    assert [slot.id for slot in boundary] == [
        "platform-mlb-in",
        "platform-mlb-out",
        "gamedev-mlb-in",
        "gamedev-mlb-out",
        "sysadmin-mlb-in",
        "sysadmin-mlb-out",
    ]


def test_the_scoped_halves_measure_the_grant_and_its_limit() -> None:
    slots = {slot.id: slot.descriptor for slot in derive(SCOPED)}
    within = 'holds "changes running systems" within: a local world you run yourself'
    assert slots["gamedev-mlb-in"] == within
    beyond = 'defers "changes running systems" past that scope to sysadmin'
    assert slots["gamedev-mlb-out"] == beyond


def test_a_deferring_role_beside_a_scoped_one_still_reads_the_classic_way() -> None:
    slots = {slot.id: slot.descriptor for slot in derive(SCOPED)}
    assert slots["platform-mlb-out"] == 'defers "changes running systems" to sysadmin'
    assert slots["platform-mlb-in"] == "owns: build and land work"


def test_the_scope_travels_with_the_case_so_an_author_can_read_the_limit() -> None:
    payloads = {slot.id: slot.to_dict() for slot in derive(SCOPED)}
    assert payloads["gamedev-mlb-in"]["scope"] == "a local world you run yourself"
    assert "scope" not in payloads["platform-mlb-in"]


def test_a_roster_declaring_no_scoped_grant_carries_no_scope_anywhere() -> None:
    slots = derive(ROSTER)
    assert all(slot.scope is None for slot in slots)
    assert all("scope" not in slot.to_dict() for slot in slots)
    assert sum(1 for slot in slots if slot.test_type == BOUNDARY) == 6


BOARD = Path(__file__).resolve().parents[2] / "evaluations" / "reflow-v3" / "boundaries.yaml"


def _board() -> list[dict[str, Any]]:
    entries: list[dict[str, Any]] = yaml.safe_load(BOARD.read_text())["boundaries"]
    return entries


def test_every_declared_pair_is_named_the_way_the_matrix_derives_it() -> None:
    for entry in _board():
        origin = str(entry["origin"]).removeprefix("boundary-")
        assert entry["id"] == f"{entry['role']}-{abbreviate(origin)}"


def test_the_board_declares_six_non_owner_seats_for_each_of_four_boundaries() -> None:
    board = _board()
    per_origin: dict[str, set[str]] = {}
    for entry in board:
        per_origin.setdefault(str(entry["origin"]), set()).add(str(entry["role"]))
    assert len(board) == 24
    assert len(per_origin) == 4
    assert all(len(roles) == 6 for roles in per_origin.values())
