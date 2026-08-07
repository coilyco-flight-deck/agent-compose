"""Score preserved evaluation responses in independent reviewer sessions.

The reviewer never sees the driver session. It receives the rendered role
context, the verbatim case prompt, the rubric with its score scale, and the
preserved response, then returns one score and one evidence sentence per
criterion. Totals and verdicts are derived by the record writer from the
pack review rule, so the reviewer only supplies judgement.
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
import time
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path

import yaml

REVIEW_FORMAT = "agent-compose.evaluation-review.v1"

INSTRUCTIONS = """You are an independent evaluation reviewer. You did not write
the role context, the case prompt, or the response. Score the preserved
response exactly as written. Do not rewrite, complete, excuse, or improve it.

Score every criterion from 0 to 2 using its own scale. Record one sentence of
evidence for each score, quoting or naming what in the response earned it.

Return only a JSON object of this exact shape, with no prose around it:

{"scores": [{"criterion": "<id>", "score": <0|1|2>, "evidence": "<one sentence>"}]}

Include every criterion exactly once. Use no other keys.
"""


def load_pack(packs: Path, role: str) -> dict:
    for path in sorted(packs.glob("*.yaml")):
        pack = yaml.safe_load(path.read_text())
        if pack["role"] == role:
            return pack
    raise SystemExit(f"no pack for role {role}")


def role_context(pack: dict) -> str:
    """Render the same doctrine the driver received, for judging role fit."""

    blocks = [f"# Role under review: {pack['role']}", pack["briefing"]]
    for meld in pack.get("melds") or []:
        blocks.append(meld["definition"])
    for personality in pack.get("personalities") or []:
        blocks.append(personality["definition"])
    blocks.append(pack["invariant"])
    return "\n\n".join(block.strip() for block in blocks)


def review_prompt(pack: dict, case: dict, answer: str) -> str:
    rubric = [
        {"criterion": item["id"], "question": item["question"], "scale": item["scale"]}
        for item in case["rubric"]
    ]
    return "\n\n".join(
        [
            INSTRUCTIONS.strip(),
            "## Role context given to the responder",
            role_context(pack),
            "## Reviewer question",
            case["reviewer_question"],
            "## Case prompt given to the responder",
            case["prompt"],
            "## Preserved response",
            answer if answer.strip() else "(empty response)",
            "## Rubric",
            json.dumps(rubric, indent=2),
        ]
    )


def parse_scores(stdout: str) -> list[dict] | None:
    decoder = json.JSONDecoder()
    for index, character in enumerate(stdout):
        if character != "{":
            continue
        try:
            payload, _ = decoder.raw_decode(stdout[index:])
        except json.JSONDecodeError:
            continue
        if isinstance(payload, dict) and isinstance(payload.get("scores"), list):
            return payload["scores"]
    return None


def review_case(
    pack: dict,
    case: dict,
    record: dict,
    model: str,
    effort: str,
    home: Path,
    timeout: int,
    retries: int,
) -> dict:
    env = dict(os.environ)
    env.pop("AGENT_COMPOSE_LAUNCH", None)
    env.pop("AGENT_COMPOSE_MODEL_TIER", None)
    env["HOME"] = str(home)

    command = [
        "claude",
        "--print",
        "--model",
        model,
        "--effort",
        effort,
        "--no-session-persistence",
        # The reviewer judges preserved text and must not act on the host.
        "--tools",
        "",
        review_prompt(pack, case, record.get("answer", "")),
    ]

    expected = {item["id"] for item in case["rubric"]}
    reason = "no attempt"
    for attempt in range(1, retries + 2):
        started = time.monotonic()
        try:
            completed = subprocess.run(
                command,
                cwd=str(home),
                env=env,
                stdin=subprocess.DEVNULL,
                capture_output=True,
                text=True,
                timeout=timeout,
            )
        except subprocess.TimeoutExpired:
            reason = "timeout"
            continue
        scores = parse_scores(completed.stdout)
        if scores is None:
            reason = (
                f"attempt {attempt}: unparsable reviewer output"
                f" (exit {completed.returncode})"
            )
            continue
        seen = sorted(str(entry.get("criterion")) for entry in scores)
        if set(seen) != expected:
            reason = f"attempt {attempt}: criteria {seen} do not match {sorted(expected)}"
            continue
        if any(entry.get("score") not in (0, 1, 2) for entry in scores):
            reason = "score outside 0..2"
            continue
        if any(not str(entry.get("evidence", "")).strip() for entry in scores):
            reason = "missing evidence sentence"
            continue
        return {
            "id": case["id"],
            "role": pack["role"],
            "reviewed": True,
            "duration_ms": int((time.monotonic() - started) * 1000),
            "scores": [
                {
                    "criterion": entry["criterion"],
                    "score": entry["score"],
                    "evidence": str(entry["evidence"]).strip(),
                }
                for entry in scores
            ],
        }
    return {
        "id": case["id"],
        "role": pack["role"],
        "reviewed": False,
        "reason": reason,
        "scores": [],
    }


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--packs", type=Path, required=True)
    parser.add_argument("--run", type=Path, required=True, help="driver run JSON")
    parser.add_argument("--out", type=Path, required=True)
    parser.add_argument("--model", default="claude-opus-5")
    parser.add_argument("--effort", default="high")
    parser.add_argument(
        "--home",
        type=Path,
        required=True,
        help="isolated reviewer home, separate from the driver context",
    )
    parser.add_argument("--timeout", type=int, default=900)
    parser.add_argument("--retries", type=int, default=1)
    parser.add_argument("--jobs", type=int, default=1)
    args = parser.parse_args(argv)

    run = json.loads(args.run.read_text())
    packs: dict[str, dict] = {}
    cases: dict[tuple[str, str], dict] = {}
    for record in run["cases"]:
        role = record["role"]
        if role not in packs:
            packs[role] = load_pack(args.packs, role)
            for case in packs[role]["cases"]:
                cases[(role, case["id"])] = case

    def execute(record: dict) -> dict:
        role = record["role"]
        reviewed = review_case(
            packs[role],
            cases[(role, record["id"])],
            record,
            args.model,
            args.effort,
            args.home,
            args.timeout,
            args.retries,
        )
        print(
            f"reviewed {role}/{record['id']}"
            f" {'ok' if reviewed['reviewed'] else 'FAILED'}",
            file=sys.stderr,
            flush=True,
        )
        return reviewed

    if args.jobs > 1:
        with ThreadPoolExecutor(max_workers=args.jobs) as pool:
            reviews = list(pool.map(execute, run["cases"]))
    else:
        reviews = [execute(record) for record in run["cases"]]

    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(
        json.dumps(
            {
                "format": REVIEW_FORMAT,
                "reviewer_model": args.model,
                "reviewer_effort": args.effort,
                "driver_arm": run["arm"],
                "driver_model": run["model"],
                "source_revision": run["source_revision"],
                "reviews": reviews,
            },
            indent=2,
        )
        + "\n"
    )
    failed = [review for review in reviews if not review["reviewed"]]
    print(f"wrote {args.out} ({len(failed)} unreviewed)", file=sys.stderr)
    return 1 if failed else 0


if __name__ == "__main__":
    raise SystemExit(main())
