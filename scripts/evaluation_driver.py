"""Execute evaluation-pack cases against a composed role bundle.

The driver launches one fresh harness session per case through the real
`acompose <role> <harness>` path, so the bundle under test is the same one a
native launch would receive. It preserves the raw response, the resolved
model, the finish reason, and every retry before any review happens.
"""

from __future__ import annotations

import argparse
import json
import os
import shutil
import subprocess
import sys
import time
from concurrent.futures import ThreadPoolExecutor
from dataclasses import dataclass, field
from pathlib import Path

import yaml

RESULT_FORMAT = "agent-compose.evaluation-run.v1"


@dataclass
class Arm:
    """One frozen driver configuration under comparison."""

    name: str
    model: str
    effort: str

    @classmethod
    def parse(cls, spec: str) -> "Arm":
        parts = spec.split(":")
        if len(parts) != 3 or not all(parts):
            raise argparse.ArgumentTypeError(
                "arm must be name:model:effort, for example opus-high:claude-opus-5:high"
            )
        return cls(*parts)


@dataclass
class Case:
    role: str
    pack_path: Path
    pack_digest: str
    case_id: str
    tier: str
    dimension: str
    scenario_kind: str
    prompt: str


@dataclass
class Attempt:
    attempt: int
    outcome: str
    reason: str = ""
    duration_ms: int = 0
    result: dict = field(default_factory=dict)


def load_cases(packs: Path, tier: str, roles: list[str] | None) -> list[Case]:
    index = json.loads((packs / "index.json").read_text())
    digests = {entry["role"]: entry["pack_digest"] for entry in index["packs"]}
    cases: list[Case] = []
    for pack_path in sorted(packs.glob("*.yaml")):
        pack = yaml.safe_load(pack_path.read_text())
        role = pack["role"]
        if roles and role not in roles:
            continue
        for raw in pack["cases"]:
            if raw["model_tier"] != tier:
                continue
            cases.append(
                Case(
                    role=role,
                    pack_path=pack_path,
                    pack_digest=digests[role],
                    case_id=raw["id"],
                    tier=raw["model_tier"],
                    dimension=raw["dimension"],
                    scenario_kind=raw["scenario_kind"],
                    prompt=raw["prompt"],
                )
            )
    return cases


def stage_session(root: Path, arm: Arm, case: Case) -> Path:
    """Create the per-case working directory.

    Every case gets a fresh projection target and a fresh harness process.
    The isolated home stays stable across cases because the harness keys its
    credential to that directory, and a per-case home would start logged out.
    """

    session = root / arm.name / case.role / case.case_id
    if session.exists():
        shutil.rmtree(session)
    cwd = session / "cwd"
    cwd.mkdir(parents=True)
    return cwd


def run_case(
    case: Case,
    arm: Arm,
    root: Path,
    home: Path | None,
    harness: str,
    timeout: int,
    retries: int,
) -> dict:
    cwd = stage_session(root, arm, case)
    env = dict(os.environ)
    env.pop("AGENT_COMPOSE_LAUNCH", None)
    env.pop("AGENT_COMPOSE_MODEL_TIER", None)
    if home is not None:
        env["AGENT_COMPOSE_RUNTIME_HOME"] = str(home)

    command = [
        "acompose",
        case.role,
        harness,
        "--print",
        "--model",
        arm.model,
        "--effort",
        arm.effort,
        "--no-session-persistence",
        "--output-format",
        "json",
        case.prompt,
    ]

    attempts: list[Attempt] = []
    for attempt in range(1, retries + 2):
        started = time.monotonic()
        try:
            completed = subprocess.run(
                command,
                cwd=cwd,
                env=env,
                stdin=subprocess.DEVNULL,
                capture_output=True,
                text=True,
                timeout=timeout,
            )
        except subprocess.TimeoutExpired:
            attempts.append(
                Attempt(attempt=attempt, outcome="failed", reason="timeout")
            )
            continue
        duration = int((time.monotonic() - started) * 1000)
        payload = parse_result(completed.stdout)
        if payload is None:
            attempts.append(
                Attempt(
                    attempt=attempt,
                    outcome="failed",
                    reason=f"unparsable harness output (exit {completed.returncode})",
                    duration_ms=duration,
                )
            )
            (cwd.parent / f"stderr-{attempt}.txt").write_text(completed.stderr)
            (cwd.parent / f"stdout-{attempt}.txt").write_text(completed.stdout)
            continue
        answer = payload.get("result") or ""
        if payload.get("is_error") or not answer.strip():
            attempts.append(
                Attempt(
                    attempt=attempt,
                    outcome="failed",
                    reason=payload.get("terminal_reason") or "empty response",
                    duration_ms=duration,
                    result=payload,
                )
            )
            continue
        attempts.append(
            Attempt(
                attempt=attempt,
                outcome="succeeded",
                duration_ms=duration,
                result=payload,
            )
        )
        break

    final = attempts[-1]
    payload = final.result
    return {
        "id": case.case_id,
        "role": case.role,
        "model_tier": case.tier,
        "dimension": case.dimension,
        "scenario_kind": case.scenario_kind,
        "pack_digest": case.pack_digest,
        "arm": arm.name,
        "requested_model": arm.model,
        "requested_effort": arm.effort,
        "resolved_models": sorted((payload.get("modelUsage") or {}).keys()),
        "question": case.prompt,
        "answer": payload.get("result", ""),
        "finish_reason": payload.get("stop_reason") or payload.get("terminal_reason"),
        "succeeded": final.outcome == "succeeded",
        "duration_ms": final.duration_ms,
        "cost_usd": payload.get("total_cost_usd"),
        "usage": payload.get("usage"),
        "session_dir": str(cwd.parent),
        "isolated_home": str(home) if home is not None else None,
        "retries": [
            {"attempt": a.attempt, "outcome": a.outcome, "reason": a.reason}
            for a in attempts
            if a.outcome != "succeeded" or a.attempt > 1
        ],
    }


def parse_result(stdout: str) -> dict | None:
    """Return the harness result object from mixed launcher and harness output.

    The launcher prints composition status before the harness output, and the
    harness emits either a single result object or a message array, so the
    payload is located by decoding rather than by line position.
    """

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


def source_revision() -> str:
    result = subprocess.run(
        ["git", "rev-parse", "HEAD"], capture_output=True, text=True, check=False
    )
    return result.stdout.strip() or "unknown"


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--packs", type=Path, required=True)
    parser.add_argument("--out", type=Path, required=True)
    parser.add_argument(
        "--arm",
        type=Arm.parse,
        action="append",
        required=True,
        help="name:model:effort, repeatable",
    )
    parser.add_argument("--tier", default="frontier")
    parser.add_argument("--role", action="append", dest="roles")
    parser.add_argument("--harness", default="claude")
    parser.add_argument(
        "--home",
        type=Path,
        help="stable isolated home whose config directory holds the run credential",
    )
    parser.add_argument(
        "--allow-host-home",
        action="store_true",
        help="run against the host home, which contaminates context and cannot be evidence",
    )
    parser.add_argument("--sessions", type=Path, required=True)
    parser.add_argument("--timeout", type=int, default=900)
    parser.add_argument("--retries", type=int, default=1)
    parser.add_argument("--limit", type=int, default=0)
    parser.add_argument("--jobs", type=int, default=1)
    args = parser.parse_args(argv)

    if args.home is None and not args.allow_host_home:
        print(
            "--home is required; pass --allow-host-home only for a plumbing smoke run",
            file=sys.stderr,
        )
        return 2

    cases = load_cases(args.packs, args.tier, args.roles)
    if args.limit:
        cases = cases[: args.limit]
    if not cases:
        print("no cases selected", file=sys.stderr)
        return 2

    args.out.mkdir(parents=True, exist_ok=True)
    revision = source_revision()

    for arm in args.arm:
        completed = 0

        def execute(case: Case, arm: Arm = arm) -> dict:
            nonlocal completed
            record = run_case(
                case,
                arm,
                args.sessions,
                args.home,
                args.harness,
                args.timeout,
                args.retries,
            )
            completed += 1
            print(
                f"[{arm.name}] {completed}/{len(cases)} {case.role}/{case.case_id}"
                f" {'ok' if record['succeeded'] else 'FAILED'}",
                file=sys.stderr,
                flush=True,
            )
            return record

        if args.jobs > 1:
            with ThreadPoolExecutor(max_workers=args.jobs) as pool:
                records = list(pool.map(execute, cases))
        else:
            records = [execute(case) for case in cases]
        run = {
            "format": RESULT_FORMAT,
            "arm": arm.name,
            "model": arm.model,
            "effort": arm.effort,
            "harness": args.harness,
            "model_tier": args.tier,
            "source_revision": revision,
            "isolated_home": str(args.home) if args.home else None,
            "context_isolation": "isolated" if args.home else "host-contaminated",
            "cases": records,
        }
        target = args.out / f"{arm.name}.json"
        target.write_text(json.dumps(run, indent=2, sort_keys=False) + "\n")
        print(f"wrote {target}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
