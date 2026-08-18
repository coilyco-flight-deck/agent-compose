"""Derive the case list from the roster Go exports.

The board is a consequence of the roster, not a hand-maintained list. Adding a
boundary or changing adjacency changes this output. See docs/evaluation.md.
"""

from __future__ import annotations

import argparse
import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any

import yaml
from aos_eval.schema import AGENT_COMPOSE, Half

# The taxonomy is the profile's, not this module's. Unpacking by arity means
# a fourth test type fails loudly here rather than being silently unranked.
BOUNDARY, ROLE_FIT, PERSONALITY = (spec.name for spec in AGENT_COMPOSE.test_types)


@dataclass(frozen=True)
class Slot:
    """One sample the dataset must contain, before anyone authors it."""

    id: str
    role: str
    test_type: str
    descriptor: str
    boundary: str | None = None
    half: Half | None = None
    pair_id: str | None = None
    target: str | None = None
    trait: str | None = None

    def to_dict(self) -> dict[str, Any]:
        payload: dict[str, Any] = {
            "id": self.id,
            "role": self.role,
            "test_type": self.test_type,
        }
        for key in ("boundary", "pair_id", "target", "trait"):
            value = getattr(self, key)
            if value:
                payload[key] = value
        if self.half:
            payload["half"] = self.half.value
        payload["descriptor"] = self.descriptor
        return payload


def abbreviate(slug: str) -> str:
    return "".join(part[0] for part in slug.split("-"))


def boundary_slots(roster: dict[str, Any]) -> list[Slot]:
    slots: list[Slot] = []
    order = [name for name in AGENT_COMPOSE.boundary_order if name in roster.get("boundaries", {})]
    order += [name for name in roster.get("boundary_order", []) if name not in order]

    for boundary in order:
        spec = roster["boundaries"][boundary]
        owner = str(spec.get("owner", ""))
        deferring = [
            role
            for role in roster["role_order"]
            if boundary in roster["roles"][role].get("boundaries", [])
        ]
        short = abbreviate(boundary)
        behaviour = owner_behaviour(str(spec.get("summary", "")), owner, roster)
        for role in [*deferring, owner]:
            if not role:
                continue
            purpose = _shorten(str(roster["roles"][role].get("purpose", "")))
            for half in (Half.IN, Half.OUT):
                slots.append(
                    Slot(
                        id=f"{role}-{short}-{half.value}",
                        role=role,
                        test_type=BOUNDARY,
                        boundary=boundary,
                        half=half,
                        pair_id=f"{role}-{short}",
                        descriptor=_boundary_descriptor(
                            role == owner, half, owner, behaviour, purpose
                        ),
                    )
                )
    return slots


def owner_behaviour(summary: str, owner: str, roster: dict[str, Any]) -> str:
    """The owner's clause of a summary reading '<Owner> <does X>, other roles ...'.

    Quoted verbatim downstream. The clause is conjugated for the owner's name, so
    negating it in prose would need inflection this renderer has no business doing.
    """
    head = summary.split(",", 1)[0].strip()
    display = str(roster["roles"].get(owner, {}).get("display_name", "")).strip()
    if display and head.lower().startswith(display.lower()):
        head = head[len(display) :].strip()
    return head or summary.strip()


def role_fit_slots(roster: dict[str, Any]) -> list[Slot]:
    slots: list[Slot] = []
    for role in roster["role_order"]:
        slots.append(
            Slot(
                id=f"{role}-fit-within",
                role=role,
                test_type=ROLE_FIT,
                target="within",
                descriptor=f"{role} correctly identifies work it should own",
            )
        )
        for adjacent in roster["roles"][role].get("adjacents", []):
            slots.append(
                Slot(
                    id=f"{role}-fit-{adjacent['role']}",
                    role=role,
                    test_type=ROLE_FIT,
                    target=str(adjacent["role"]),
                    descriptor=str(adjacent["reason"]),
                )
            )
    return slots


def personality_slots(roster: dict[str, Any]) -> list[Slot]:
    """One case per trait. Every case still runs against the fully composed bundle."""
    slots: list[Slot] = []
    for role in roster["role_order"]:
        traits = list(roster["roles"][role]["personalities"])
        for trait in traits:
            peers = ", ".join(other for other in traits if other != trait)
            slots.append(
                Slot(
                    id=f"{role}-per-{trait}",
                    role=role,
                    test_type=PERSONALITY,
                    trait=trait,
                    descriptor=f"{trait}, composed alongside {peers}" if peers else trait,
                )
            )
    return slots


def derive(roster: dict[str, Any], group: str = "tier") -> list[Slot]:
    slots = boundary_slots(roster) + role_fit_slots(roster) + personality_slots(roster)
    if group != "role":
        return slots
    order = list(roster["role_order"])
    return sorted(slots, key=lambda s: (order.index(s.role), AGENT_COMPOSE.rank(s.test_type)))


def render(slots: list[Slot]) -> str:
    lines: list[str] = []
    for index, slot in enumerate(slots, start=1):
        lines.append(f"{index:3d}. {slot.id:<26} {slot.role:<9} {slot.descriptor}")

    tiers: dict[str, int] = {}
    per_role: dict[str, int] = {}
    for slot in slots:
        tiers[slot.test_type] = tiers.get(slot.test_type, 0) + 1
        per_role[slot.role] = per_role.get(slot.role, 0) + 1

    lines.append("")
    lines.append(f"{len(slots)} cases: " + ", ".join(f"{v} {k}" for k, v in tiers.items()))
    lines.append("per role: " + ", ".join(f"{k} {v}" for k, v in sorted(per_role.items())))
    return "\n".join(lines)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Print the case list the roster implies.")
    parser.add_argument("--roster", type=Path, required=True, help="person.json from the roster")
    parser.add_argument("--format", choices=("text", "yaml"), default="text")
    parser.add_argument("--group", choices=("role", "tier"), default="role")
    args = parser.parse_args(argv)

    slots = derive(json.loads(args.roster.read_text()), args.group)
    if args.format == "yaml":
        print(yaml.safe_dump({"slots": [s.to_dict() for s in slots]}, sort_keys=False, width=100))
    else:
        print(render(slots))
    return 0


def _boundary_descriptor(
    is_owner: bool, half: Half, owner: str, behaviour: str, purpose: str
) -> str:
    if is_owner and half is Half.IN:
        return f'owns "{behaviour}"'
    if is_owner:
        return f'owns "{behaviour}", claims nothing past it'
    if half is Half.IN:
        return f"owns: {purpose}"
    return f'defers "{behaviour}" to {owner}'


def _shorten(purpose: str) -> str:
    text = purpose.strip().rstrip(".")
    return text[:1].lower() + text[1:] if text else text


if __name__ == "__main__":
    raise SystemExit(main())
