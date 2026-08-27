#!/usr/bin/env python3
"""Throwaway YAML-to-bundle compositor for the #332 tracer.

Reads testdata/tracer/roster.yaml and emits a bundle `agent-compose verify`
accepts, with no Go in the loop. #333 replaces this with the real compositor,
so nothing here is built to last and nothing else should import it.

What it cheats, so #333 knows what it owes:

* favorite_color is read from the YAML. The Go engine derives it as the OKLab
  centroid of the meld with chroma restored and lightness clamped. Not ported.
* Only the `native-skills` delivery mode. `compiled` is not emitted.
* One role, one person package. No local library merge, no conflict detection,
  no cross-package collision handling.
* No validation of any kind: no copy contract, no role-skill frontmatter, no
  three-paragraph rule, no 400-word body cap, no role-by-tier compatibility
  matrix, and no rejection of a role that declares its own boundary.
* Content digests cover the emitted bytes, which happens to be right, but
  nothing checks them and the real engine digests logical content instead.
* Decision reasons are close paraphrases of the Go strings rather than the
  same strings. Byte-identity is #333's acceptance bar, not this script's.
* model_tier is passed through. The tier compatibility check is absent.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys

import yaml

SOURCE_SAFE = set("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_.")


def source_segment(value: str) -> str:
    out = []
    for octet in value.encode():
        char = chr(octet)
        out.append(char if char in SOURCE_SAFE else f"%{octet:02X}")
    return "".join(out)


def digest(text: str) -> str:
    return "sha256:" + hashlib.sha256(text.encode()).hexdigest()


def allocate_boundaries(role: dict, slug: str, boundaries: dict) -> list[tuple[str, str]]:
    """Deferred and scoped in source order, then the one this role owns."""
    allocated = [(name, "defer") for name in role.get("defers", [])]
    allocated += [(name, "scope") for name in role.get("scoped", {})]
    owned = [name for name, spec in boundaries.items() if spec["owner"] == slug]
    if any(name in dict(allocated) for name in owned):
        raise SystemExit(f"role {slug} declares the boundary it owns")
    return allocated + [(name, "own") for name in owned]


def render_instructions(roster: dict, slug: str, allocated: list[tuple[str, str]]) -> str:
    role = roster["roles"][slug]
    boundaries = roster["boundaries"]
    melds = [roster["personalities"][name] for name in role["personalities"]]
    skills = [(role["skill"], role["body"])]
    skills += [(boundaries[name]["skill"], boundaries[name]["body"]) for name, _ in allocated]
    skills += [(meld["skill"], meld["body"]) for meld in melds]
    total = sum(len(body.encode()) for _, body in skills)

    seats = " // ".join(agent["key"] for agent in role["agents"])
    boundary_marks = " // ".join(f"`{boundaries[name]['skill']}`" for name, _ in allocated)
    lines = [
        "# Role instructions",
        "",
        f"Agent-compose assigned you the `{slug}` role from the caller's compose request.",
        "Treat it as authoritative and fixed for this session.",
        "",
        f"# {role['display_name']}",
        "",
        role["purpose"],
        "",
        f"**Role skill // `{role['skill']}`**",
        f"**Boundaries // {boundary_marks}**",
        f"**Favorite color // `{role['favorite_color']}`**",
        f"**Agent // {role['identity']['name']} ({role['identity']['pronouns']})**",
        f"**Seats // {seats}**",
        "",
        "## Personality meld",
        "",
    ]
    for name, meld in zip(role["personalities"], melds):
        emblem = " / ".join(meld["emblem"]["names"])
        lines += [
            f"### {meld['emblem']['emoji']} {name.capitalize()}",
            "",
            f"**{meld['color']} // {emblem} // {meld['motif']}**",
            "",
        ]
    lines += ["## Boundaries", ""]
    for name, disposition in allocated:
        spec = boundaries[name]
        verb = {"defer": "you defer this", "own": "you own this",
                "scope": "you hold this within a scope"}[disposition]
        entry = f"* `{spec['skill']}` - {verb}. {spec['summary']}"
        if disposition == "scope":
            entry += f". Your scope: {role['scoped'][name]}"
        lines.append(entry)
    lines += [
        "",
        "## Active doctrine",
        "",
        f"Everything above summarizes {total:,} bytes of doctrine across these "
        f"{len(skills)} skills, and a summary is not the operative text. "
        "Before acting, load each one:",
        "",
    ]
    for skill, body in skills:
        lines.append(f"* `{skill}` - {len(body.encode()):,} bytes")
    lines += ["", "", roster["invariant"].rstrip("\n"), ""]
    return "\n".join(lines)


def compose(roster: dict, slug: str, out_dir: pathlib.Path) -> pathlib.Path:
    role = roster["roles"][slug]
    boundaries = roster["boundaries"]
    source = roster["source"]
    allocated = allocate_boundaries(role, slug, boundaries)

    selected = [(role["skill"], role["body"], f'role "{slug}" composes this skill')]
    selected += [
        (boundaries[name]["skill"], boundaries[name]["body"], f'role "{slug}" composes this skill')
        for name, _ in allocated
    ]
    selected += [
        (roster["personalities"][name]["skill"], roster["personalities"][name]["body"],
         f'active personality "{name}" binds this skill')
        for name in role["personalities"]
    ]

    out_dir.mkdir(parents=True, exist_ok=True)
    skills_root = out_dir / "content" / "skills" / source_segment(source)
    context_bytes = 0
    for skill, body, _ in selected:
        target = skills_root / skill / "SKILL.md"
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(body)
        context_bytes += len(body.encode())

    instructions = render_instructions(roster, slug, allocated)
    (out_dir / "content" / "instructions.md").write_text(instructions)

    manifest = {
        "format": "agent-compose.bundle",
        "role": slug,
        "role_skill": role["skill"],
        "role_skill_source": f"{source}:role:{slug}",
        "role_skill_digest": digest(role["body"]),
        "model_tier": role["model_tier"][0],
        "personalities": list(role["personalities"]),
        "boundaries": [name for name, _ in allocated],
        "color": role["favorite_color"],
        "identity": {
            "person": roster["person"],
            "purpose": role["purpose"],
            "seats": [
                {"key": a["key"], "harness": a["key"], "name": role["identity"]["name"],
                 "pronouns": role["identity"]["pronouns"],
                 **({"tier": a["tier"]} if "tier" in a else {})}
                for a in role["agents"]
            ],
            "personalities": [
                {"name": name, "color": roster["personalities"][name]["color"],
                 "emblem": {"names": roster["personalities"][name]["emblem"]["names"],
                            "emoji": roster["personalities"][name]["emblem"]["emoji"]}}
                for name in role["personalities"]
            ],
        },
        "sources": [source],
        "content": [{"id": f"{source}:instructions", "digest": digest(instructions)}]
        + [{"id": f"{source}:skill:{skill}", "digest": digest(body)} for skill, body, _ in selected],
        "delivery": {
            "mode": "native-skills",
            "instructions": "content/instructions.md",
            "skills_root": "content/skills",
            "body_bytes": len(instructions.encode()),
        },
    }

    decisions = [{"subject": f"source:{source}", "kind": "source", "source": source,
                  "outcome": "selected", "reason": "the roster file admits this source"}]
    decisions.append({"subject": f"role:{slug}", "kind": "profile", "source": source,
                      "outcome": "selected", "reason": f"{source} defines this role: {role['purpose']}"})
    meld_list = ", ".join(role["personalities"])
    for name in role["personalities"]:
        decisions.append({"subject": f"personality:{name}", "kind": "profile", "source": source,
                          "outcome": "selected",
                          "reason": f'role "{slug}" activates its full personality set: {meld_list}'})
    for name, disposition in allocated:
        decisions.append({"subject": f"boundary:{name}", "kind": "profile", "source": source,
                          "outcome": "selected",
                          "reason": f'role "{slug}" {disposition}s boundary "{name}"'})
    decisions.append({"subject": "instruction:personality-invariant", "kind": "instruction",
                      "source": source, "outcome": "selected",
                      "reason": "instructions from admitted sources are always selected"})
    for skill, _, reason in selected:
        decisions.append({"subject": f"skill:{skill}", "kind": "skill", "source": source,
                          "outcome": "selected", "reason": reason})
    decisions.append({"subject": "content/instructions.md", "kind": "delivery",
                      "outcome": "delivered", "reason": "canonical selected instructions"})
    decisions.append({"subject": "content/skills", "kind": "delivery",
                      "outcome": "delivered", "reason": "canonical selected skill trees"})

    trace = {
        "format": "agent-compose.trace",
        "decisions": decisions,
        "providers": [{
            "source": source, "category": "person-package", "scope": "person",
            "outcome": "selected", "reason": "the roster file admits this source",
            "skills": len(selected), "context_bytes": context_bytes,
            "approximate_tokens": (context_bytes + 3) // 4,
        }],
    }

    (out_dir / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    (out_dir / "trace.json").write_text(json.dumps(trace, indent=2) + "\n")
    return out_dir


def main() -> int:
    parser = argparse.ArgumentParser(prog="tracer-compose")
    parser.add_argument("--roster", default="testdata/tracer/roster.yaml")
    parser.add_argument("--role", default="tpm")
    parser.add_argument("--out", required=True)
    args = parser.parse_args()

    roster = yaml.safe_load(pathlib.Path(args.roster).read_text())
    if args.role not in roster["roles"]:
        raise SystemExit(f"roster defines no role {args.role!r}")
    out = compose(roster, args.role, pathlib.Path(args.out))
    print(f"composed {args.role} into {out}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
