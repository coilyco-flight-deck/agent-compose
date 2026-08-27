"""Compose one bundle from the roster: python -m housecast --role tpm --out DIR."""

from __future__ import annotations

import argparse
import pathlib
import sys

from housecast import compose, roster


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(prog="housecast")
    parser.add_argument("--roster", default=str(roster.DATA))
    parser.add_argument("--role", required=True)
    parser.add_argument("--model-tier", default="frontier")
    parser.add_argument("--out", required=True)
    args = parser.parse_args(argv)

    loaded = roster.load(args.roster)
    if args.role not in loaded.roles:
        raise SystemExit(f"roster defines no role {args.role!r}")
    out = compose.compose(loaded, args.role, args.model_tier, pathlib.Path(args.out))
    print(f"composed {args.role} at {args.model_tier} into {out}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
