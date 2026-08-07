"""Score preserved evaluation responses in isolated reviewer sessions.

The reviewer receives the owning role contract, the case prompt, the rubric,
and one preserved answer. It never learns which driver arm produced the
answer, and it never sees the other arm's response to the same case.
Verdicts are computed from the pack's review rule rather than asked for.
"""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
import time
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path

import yaml

REVIEW_FORMAT = "agent-compose.evaluation-review.v1"

INSTRUCTIONS = """You are scoring one response produced by an AI agent that was \
given a role contract. Score only the response text supplied below, against the \
supplied rubric. Do not rewrite the response, do not reward length, and do not \
credit a claim the response does not make.

Score each criterion 0, 1, or 2 using its own scale. Record one sentence of \
evidence per score, quoting or naming the specific part of the response that \
decided it.

Reply with one JSON object and nothing else:
{"scores": {"<criterion-id>": <0|1|2>, ...}, "evidence": {"<criterion-id>": "<one sentence>", ...}}
"""


def build_prompt(pack: dict, case: dict, answer: str) -> str:
    sections = [
        INSTRUCTIONS,
        "## Role contract under test",
        f"Purpose: {pack['purpose']}",
        pack["briefing"],
        "## Personality invariant",
        pack["invariant"],
    ]
    if case["scenario_kind"] == "personality":
        for personality in pack["personalities"]:
            sections.append(f"## Personality: {personality['name']}")
            sections.append(personality["definition"])
    sections += [
        "## Reviewer question",
        case["reviewer_question"],
        "## Case prompt given to the agent",
        case["prompt"],
        "## Rubric",
        json.dumps(case["rubric"], indent=2),
        "## Response to score",
        answer,
    ]
    return "\n\n".join(sections)


def verdict(pack: dict, case: dict, scores: dict[str, int]) -> tuple[int, bool, list[str]]:
    rule = pack["review_rule"]
    total = sum(scores.values())
    reasons: list[str] = []
    if total < rule["passing_total"]:
        reasons.append(f"total {total} below {rule['passing_total']}")
    if rule.get("reject_zero_scores") and any(value == 0 for value in scores.values()):
        reasons.append("zero score present")
    minimums = (
        rule["personality_minimum_scores"]
        if case["scenario_kind"] == "personality"
        else rule["role_minimum_scores"]
    )
    for criterion, floor in minimums.items():
        if criterion in scores and scores[criterion] < floor:
            reasons.append(f"{criterion} {scores[criterion]} below {floor}")
    return total, not reasons, reasons


def parse_result(stdout: str) -> dict | None:
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


def extract_scores(text: str) -> dict | None:
    decoder = json.JSONDecoder()
    for index, character in enumerate(text):
        if character != "{":
            continue
        try:
            payload, _ = decoder.raw_decode(text[index:])
        except json.JSONDecodeError:
            continue
        if isinstance(payload, dict) and "scores" in payload:
            return payload
    return None


def review_one(
    pack: dict, case: dict, record: dict, model: str, effort: str, timeout: int
) -> dict:
    prompt = build_prompt(pack, case, record["answer"])
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
        prompt,
    ]
    started = time.monotonic()
    completed = subprocess.run(
        command,
        stdin=subprocess.DEVNULL,
        capture_output=True,
        text=True,
        timeout=timeout,
    )
    duration = int((time.monotonic() - started) * 1000)
    payload = parse_result(completed.stdout) or {}
    scored = extract_scores(payload.get("result") or "")
    if scored is None:
        return {
            "id": record["id"],
            "role": record["role"],
            "arm": record["arm"],
            "reviewed": False,
            "reason": "reviewer returned no parsable scores",
            "duration_ms": duration,
        }
    scores = {key: int(value) for key, value in scored["scores"].items()}
    total, passed, reasons = verdict(pack, case, scores)
    return {
        "id": record["id"],
        "role": record["role"],
        "arm": record["arm"],
        "scenario_kind": case["scenario_kind"],
        "reviewed": True,
        "scores": scores,
        "evidence": scored.get("evidence", {}),
        "total": total,
        "passed": passed,
        "failure_reasons": reasons,
        "reviewer_model": model,
        "reviewer_effort": effort,
        "reviewer_cost_usd": payload.get("total_cost_usd"),
        "duration_ms": duration,
    }


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--packs", type=Path, required=True)
    parser.add_argument("--runs", type=Path, nargs="+", required=True)
    parser.add_argument("--out", type=Path, required=True)
    parser.add_argument("--model", default="claude-opus-5")
    parser.add_argument("--effort", default="high")
    parser.add_argument("--timeout", type=int, default=900)
    parser.add_argument("--jobs", type=int, default=4)
    args = parser.parse_args(argv)

    packs: dict[str, dict] = {}
    cases: dict[tuple[str, str], dict] = {}
    for pack_path in sorted(args.packs.glob("*.yaml")):
        pack = yaml.safe_load(pack_path.read_text())
        packs[pack["role"]] = pack
        for case in pack["cases"]:
            cases[(pack["role"], case["id"])] = case

    work = []
    for run_path in args.runs:
        run = json.loads(run_path.read_text())
        for record in run["cases"]:
            if not record["succeeded"]:
                continue
            work.append((packs[record["role"]], cases[(record["role"], record["id"])], record))

    def execute(item):
        pack, case, record = item
        return review_one(pack, case, record, args.model, args.effort, args.timeout)

    with ThreadPoolExecutor(max_workers=args.jobs) as pool:
        reviews = list(pool.map(execute, work))

    for review in reviews:
        state = "ok" if review["reviewed"] else "FAILED"
        print(
            f"[{review['arm']}] {review['role']}/{review['id']} {state}",
            file=sys.stderr,
            flush=True,
        )

    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(
        json.dumps(
            {
                "format": REVIEW_FORMAT,
                "reviewer_model": args.model,
                "reviewer_effort": args.effort,
                "reviews": reviews,
            },
            indent=2,
        )
        + "\n"
    )
    print(f"wrote {args.out}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
