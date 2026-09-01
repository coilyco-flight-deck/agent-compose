"""Compose every shipped bundle from the roster, at build time.

housecast never runs on a user's machine, so the release carries the composed
set rather than the data it was composed from. See docs/FEATURES.md.

A shipped bundle is addressed by what a launch knows, `<role>-<tier>-<delivery>`,
rather than by the content hash the runtime cache uses. Nothing has to publish a
pointer, and the layout survives an unzip on a filesystem without symlinks.
"""

from __future__ import annotations

import argparse
import pathlib
import shutil
import sys

from housecast import compose, roster

DELIVERY_MODES = ("native-skills", "compiled")


def bundle_name(role: str, tier: str, delivery: str) -> str:
    return f"{role}-{tier}-{delivery}"


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("out", type=pathlib.Path, help="directory to compose into")
    args = parser.parse_args()

    loaded = roster.load()
    out: pathlib.Path = args.out
    if out.exists():
        shutil.rmtree(out)
    out.mkdir(parents=True)

    composed = 0
    for role_name in loaded.role_order:
        role = loaded.roles[role_name]
        for tier in role.supported_model_tiers:
            for delivery in DELIVERY_MODES:
                name = bundle_name(role_name, tier, delivery)
                compose.compose(loaded, role_name, tier, out / name, delivery=delivery)
                composed += 1

    print(f"composed {composed} bundles into {out}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
