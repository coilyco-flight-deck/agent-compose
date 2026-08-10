"""Subject fan-out. Runs every candidate n times and preserves raw responses.

Go composes the system prompt. This module only sends it. Transport is recorded
per response, so a fallback run can never be mistaken for a measured one.
"""

from __future__ import annotations

import argparse
import asyncio
import json
from dataclasses import dataclass
from pathlib import Path

import httpx
import yaml

from evalkit.schema import SUBJECT_RUNS, Candidate, Response

DEFAULT_CONCURRENCY = 10

# Agent Proxy is the only transport that produces a measured result. Direct
# provider calls are incident isolation, and the caller names the exception.
TRANSPORT_PROXY = "agent-proxy"
TRANSPORT_DIRECT = "direct"


@dataclass(frozen=True)
class Subject:
    base_url: str
    model: str
    api_key: str
    transport: str = TRANSPORT_PROXY


def load_candidates(path: Path) -> list[Candidate]:
    raw = yaml.safe_load(path.read_text()) or {}
    return [Candidate.from_dict(entry) for entry in raw.get("candidates", [])]


def load_system_prompts(directory: Path) -> dict[str, str]:
    return {path.stem: path.read_text().strip() for path in sorted(directory.glob("*.md"))}


async def ask(
    client: httpx.AsyncClient,
    subject: Subject,
    system_prompt: str,
    candidate: Candidate,
    run_index: int,
) -> Response:
    payload = {
        "model": subject.model,
        "messages": [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": candidate.prompt},
        ],
    }
    reply = await client.post(
        f"{subject.base_url.rstrip('/')}/chat/completions",
        json=payload,
        headers={"Authorization": f"Bearer {subject.api_key}"},
        timeout=120.0,
    )
    reply.raise_for_status()
    body = reply.json()
    choice = body["choices"][0]
    return Response(
        candidate_id=candidate.id,
        variant=candidate.variant,
        run=run_index,
        text=str(choice["message"]["content"]).strip(),
        finish_reason=str(choice.get("finish_reason") or "stop"),
    )


async def fan_out(
    candidates: list[Candidate],
    prompts: dict[str, str],
    subject: Subject,
    out: Path,
    runs: int = SUBJECT_RUNS,
    concurrency: int = DEFAULT_CONCURRENCY,
) -> int:
    missing = sorted({c.role for c in candidates} - prompts.keys())
    if missing:
        raise SystemExit(f"no composed system prompt for: {', '.join(missing)}")

    gate = asyncio.Semaphore(concurrency)
    written = 0

    with out.open("a") as sink:
        lock = asyncio.Lock()

        async def one(client: httpx.AsyncClient, candidate: Candidate, run_index: int) -> None:
            nonlocal written
            async with gate:
                try:
                    response = await ask(
                        client, subject, prompts[candidate.role], candidate, run_index
                    )
                except (httpx.HTTPError, KeyError, IndexError) as error:
                    response = Response(
                        candidate_id=candidate.id,
                        variant=candidate.variant,
                        run=run_index,
                        text="",
                        finish_reason=f"error: {type(error).__name__}",
                    )
            record = response.to_dict() | {"transport": subject.transport, "model": subject.model}
            async with lock:
                sink.write(json.dumps(record) + "\n")
                sink.flush()
                written += 1

        async with httpx.AsyncClient() as client:
            await asyncio.gather(
                *(
                    one(client, candidate, index)
                    for candidate in candidates
                    for index in range(1, runs + 1)
                )
            )
    return written


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Run the subject across every candidate.")
    parser.add_argument("--candidates", type=Path, required=True)
    parser.add_argument("--prompts", type=Path, required=True, help="dir of composed <role>.md")
    parser.add_argument("--out", type=Path, required=True, help="jsonl sink, appended")
    parser.add_argument("--base-url", required=True)
    parser.add_argument("--model", required=True)
    parser.add_argument("--api-key", required=True)
    parser.add_argument("--runs", type=int, default=SUBJECT_RUNS)
    parser.add_argument("--concurrency", type=int, default=DEFAULT_CONCURRENCY)
    parser.add_argument(
        "--direct",
        action="store_true",
        help="incident isolation only. Marks every response as an unmeasured fallback.",
    )
    args = parser.parse_args(argv)

    subject = Subject(
        base_url=args.base_url,
        model=args.model,
        api_key=args.api_key,
        transport=TRANSPORT_DIRECT if args.direct else TRANSPORT_PROXY,
    )
    if args.direct:
        print("WARNING: direct transport. These responses are a demonstration, not a result.")

    candidates = load_candidates(args.candidates)
    prompts = load_system_prompts(args.prompts)
    written = asyncio.run(
        fan_out(candidates, prompts, subject, args.out, args.runs, args.concurrency)
    )
    print(f"wrote {written} responses to {args.out}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
