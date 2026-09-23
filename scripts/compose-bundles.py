"""Compose every shipped bundle from the seed roster, at build time.

The Go engine composes, against the same seed roster the release ships, so the
bundles and the roster asset cannot disagree. See docs/FEATURES.md.

A shipped bundle is addressed by what a launch knows, `<role>-<tier>-<delivery>`,
rather than by the content hash the runtime cache uses. Nothing has to publish a
pointer, and the layout survives an unzip on a filesystem without symlinks.
"""

from __future__ import annotations

import argparse
import json
import os
import pathlib
import shutil
import subprocess
import sys
import tempfile

DELIVERY_MODES = ("native-skills", "compiled")
REPO = pathlib.Path(__file__).resolve().parent.parent


def bundle_name(role: str, tier: str, delivery: str) -> str:
    return f"{role}-{tier}-{delivery}"


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("out", type=pathlib.Path, help="directory to compose into")
    args = parser.parse_args()

    out: pathlib.Path = args.out
    if out.exists():
        shutil.rmtree(out)
    out.mkdir(parents=True)

    with tempfile.TemporaryDirectory() as scratch:
        work = pathlib.Path(scratch)
        binary = work / "agent-compose"
        subprocess.run(["go", "build", "-o", str(binary), "./cmd/agent-compose"], cwd=REPO, check=True)
        env = dict(os.environ, AGENT_COMPOSE_ROSTER=str(REPO / "seed" / "roster"))

        def run(*argv: str) -> str:
            return subprocess.run([str(binary), *argv], env=env, check=True, capture_output=True, text=True).stdout

        roles = json.loads(run("catalog", "roles", "--json"))["items"]
        composed = 0
        skipped: list[str] = []
        for role in roles:
            # An archived seat has no bundle to ship: the engine refuses to compose one.
            if role.get("archived"):
                skipped.append(role["slug"])
                continue
            for tier in role.get("model_tiers") or ["frontier"]:
                for delivery in DELIVERY_MODES:
                    name = bundle_name(role["slug"], tier, delivery)
                    request = work / f"{name}.kdl"
                    request.write_text(
                        f'compose {{\n    role "{role["slug"]}"\n    model-tier "{tier}"\n    delivery "{delivery}"\n}}\n'
                    )
                    staged = work / name
                    run("compose", str(request), "--out", str(staged))
                    (built,) = [p for p in staged.iterdir() if p.is_dir()]
                    shutil.move(str(built), str(out / name))
                    composed += 1

    note = f", skipping archived {', '.join(skipped)}" if skipped else ""
    print(f"composed {composed} bundles into {out}{note}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
