"""Score preserved evaluation responses in independent reviewer sessions.

The reviewer never sees the driver session. It receives the rendered role
context, the verbatim case prompt, the rubric with its score scale, and one
preserved response, then returns one score and one evidence sentence per
criterion. It never learns which driver arm produced the answer, and it never
sees a competing answer to the same case.

Totals and verdicts are derived by the record writer from the pack review
rule, so the reviewer only supplies judgement and the passing total, zero
rejection, and criterion minimums cannot drift with reviewer wording.
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

# The reviewer judges preserved text and must never act on a real system. Both
# flags take the equals form because a variadic flag would swallow the prompt.
SEALED_TOOLS = "--tools="
SEALED_MCP = '--mcp-config={"mcpServers":{}}'

INSTRUCTIONS = """You are an independent evaluation reviewer. You did not write
the role context, the case prompt, or the response. Score the preserved
response exactly as written. Do not rewrite, complete, excuse, or improve it.
Do not reward length, and do not credit a claim the response does not make.

Score every criterion from 0 to 2 using its own scale. Record one sentence of
evidence for each score, quoting or naming what in the response earned it.

Return only a JSON object of this exact shape, with no prose around it:

{"scores": [{"criterion": "<id>", "score": <0|1|2>, "evidence": "<one sentence>"}]}

Include every criterion exactly once. Use no other keys.
"""


def role_context(pack: dict, case: dict) -> str:
    """Render the same doctrine the driver received, for judging role fit."""

    blocks = [
        f"# Role under review: {pack['role']}",
        f"Purpose: {pack['purpose']}",
        pack["briefing"],
    ]
    for meld in pack.get("melds") or []:
        blocks.append(meld["definition"])
    if case["scenario_kind"] == "personality":
        for personality in pack.get("personalities") or []:
            blocks.append(f"## Personality: {personality['name']}")
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
            role_context(pack, case),
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


def parse_result(stdout: str) -> dict | None:
    """Pull the harness result envelope out of --output-format json output."""

    decoder = json.JSONDecoder()
    for index, character in enumerate(stdout):
        if character not in "[{":
            continue
        try:
            payload, _ = decoder.raw_decode(stdout[index:])
        except json.JSONDecodeError:
            continue
        messages = payload if isinstance(payload, list) else [payload]
        for message in reversed(messages):
            if isinstance(message, dict) and message.get("type") == "result":
                return message
    return None


def parse_scores(text: str) -> list[dict] | None:
    decoder = json.JSONDecoder()
    for index, character in enumerate(text):
        if character != "{":
            continue
        try:
            payload, _ = decoder.raw_decode(text[index:])
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
        "--output-format",
        "json",
        SEALED_MCP,
        SEALED_TOOLS,
        review_prompt(pack, case, record.get("answer", "")),
    ]

    expected = {item["id"] for item in case["rubric"]}
    base = {"id": case["id"], "role": pack["role"], "arm": record.get("arm")}
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
            reason = f"attempt {attempt}: timeout after {timeout}s"
            continue
        payload = parse_result(completed.stdout) or {}
        scores = parse_scores(payload.get("result") or "")
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
            reason = f"attempt {attempt}: score outside 0..2"
            continue
        if any(not str(entry.get("evidence", "")).strip() for entry in scores):
            reason = f"attempt {attempt}: missing evidence sentence"
            continue
        return {
            **base,
            "scenario_kind": case["scenario_kind"],
            "reviewed": True,
            "reviewer_cost_usd": payload.get("total_cost_usd"),
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
    return {**base, "reviewed": False, "reason": reason, "scores": []}


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--packs", type=Path, required=True)
    parser.add_argument(
        "--runs",
        type=Path,
        nargs="+",
        required=True,
        help="one or more driver run JSON files, one per arm",
    )
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
    parser.add_argument("--jobs", type=int, default=4)
    args = parser.parse_args(argv)

    packs: dict[str, dict] = {}
    cases: dict[tuple[str, str], dict] = {}
    for pack_path in sorted(args.packs.glob("*.yaml")):
        pack = yaml.safe_load(pack_path.read_text())
        packs[pack["role"]] = pack
        for case in pack["cases"]:
            cases[(pack["role"], case["id"])] = case

    sources: list[dict] = []
    work: list[dict] = []
    for run_path in args.runs:
        run = json.loads(run_path.read_text())
        sources.append(
            {
                "arm": run["arm"],
                "model": run["model"],
                "source_revision": run["source_revision"],
            }
        )
        for record in run["cases"]:
            # A case the driver could not complete has no answer worth scoring.
            if not record.get("succeeded"):
                continue
            record = dict(record)
            record.setdefault("arm", run["arm"])
            work.append(record)

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
            f"[{reviewed['arm']}] {role}/{record['id']}"
            f" {'ok' if reviewed['reviewed'] else 'FAILED'}",
            file=sys.stderr,
            flush=True,
        )
        return reviewed

    if args.jobs > 1:
        with ThreadPoolExecutor(max_workers=args.jobs) as pool:
            reviews = list(pool.map(execute, work))
    else:
        reviews = [execute(record) for record in work]

    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(
        json.dumps(
            {
                "format": REVIEW_FORMAT,
                "reviewer_model": args.model,
                "reviewer_effort": args.effort,
                "runs": sources,
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
