"""Derive the case list from the roster Go exports.

The board is a consequence of the roster, not a hand-maintained list. Adding a
boundary or changing adjacency changes this output. See docs/eval-orchestration.md.
"""

from __future__ import annotations

import argparse
import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any

import yaml

from evalkit.schema import BOUNDARY_ORDER, Half, Kind


@dataclass(frozen=True)
class Slot:
    """One case the board must contain, before anyone authors its prompt."""

    id: str
    role: str
    kind: Kind
    descriptor: str
    boundary: str | None = None
    half: Half | None = None
    pair_id: str | None = None
    target: str | None = None
    trait: str | None = None

    def to_dict(self) -> dict[str, Any]:
        payload: dict[str, Any] = {"id": self.id, "role": self.role, "kind": self.kind.value}
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
    order = [name for name in BOUNDARY_ORDER if name in roster.get("boundaries", {})]
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
                        kind=Kind.BOUNDARY,
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
                kind=Kind.ROLE_FIT,
                target="within",
                descriptor=f"{role}'s own work",
            )
        )
        for adjacent in roster["roles"][role].get("adjacents", []):
            slots.append(
                Slot(
                    id=f"{role}-fit-{adjacent['role']}",
                    role=role,
                    kind=Kind.ROLE_FIT,
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
                    kind=Kind.PERSONALITY,
                    trait=trait,
                    descriptor=f"{trait}, composed alongside {peers}" if peers else trait,
                )
            )
    return slots


def derive(roster: dict[str, Any]) -> list[Slot]:
    return boundary_slots(roster) + role_fit_slots(roster) + personality_slots(roster)


def render(slots: list[Slot]) -> str:
    lines: list[str] = []
    for index, slot in enumerate(slots, start=1):
        lines.append(f"{index:3d}. {slot.id:<26} {slot.role:<9} {slot.descriptor}")

    tiers: dict[str, int] = {}
    per_role: dict[str, int] = {}
    for slot in slots:
        tiers[slot.kind.value] = tiers.get(slot.kind.value, 0) + 1
        per_role[slot.role] = per_role.get(slot.role, 0) + 1

    lines.append("")
    lines.append(f"{len(slots)} cases: " + ", ".join(f"{v} {k}" for k, v in tiers.items()))
    lines.append("per role: " + ", ".join(f"{k} {v}" for k, v in sorted(per_role.items())))
    return "\n".join(lines)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Print the case list the roster implies.")
    parser.add_argument("--roster", type=Path, required=True, help="person.json from the roster")
    parser.add_argument("--format", choices=("text", "yaml"), default="text")
    args = parser.parse_args(argv)

    slots = derive(json.loads(args.roster.read_text()))
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
