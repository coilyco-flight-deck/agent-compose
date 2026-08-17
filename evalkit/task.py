"""The Inspect task. Replaces a hand-rolled fan-out with `inspect eval`.

Run it unscored, because the scorer is a human. Inspect calls the repetition an
epoch, which is what n=5 was. See docs/eval-pipeline.md.

    AGENTPROXY_BASE_URL=http://ser8:8080/v1 \\
    inspect eval evalkit/task.py --model openai-api/agentproxy/evaluation/deepseek-v4-flash \\
      --epochs 5 --no-score -T samples=samples.yaml -T prompts=.evalkit/prompts
"""

from __future__ import annotations

import os
from pathlib import Path

import yaml
from inspect_ai import Task, task
from inspect_ai.dataset import MemoryDataset
from inspect_ai.solver import generate, system_message

from evalkit.schema import Sample

DEFAULT_SAMPLES = Path("samples.yaml")
DEFAULT_PROMPTS = Path(".evalkit/prompts")


def load_samples(path: Path) -> list[Sample]:
    raw = yaml.safe_load(path.read_text()) or {}
    return [Sample.model_validate(entry) for entry in raw.get("samples", [])]


def load_system_prompts(directory: Path) -> dict[str, str]:
    return {p.stem: p.read_text().strip() for p in sorted(directory.glob("*.md"))}


@task
def board(samples: str | Path = DEFAULT_SAMPLES, prompts: str | Path = DEFAULT_PROMPTS) -> Task:
    """One task per run. The composed role bundle is the system message."""
    authored = load_samples(Path(samples))
    composed = load_system_prompts(Path(prompts))

    missing = sorted({s.role for s in authored} - composed.keys())
    if missing:
        raise ValueError(f"no composed system prompt for: {', '.join(missing)}")

    # One dataset per role would need one task per role. Instead the system
    # message is resolved per sample from its own role metadata.
    entries = []
    for sample in authored:
        inspect_sample = sample.to_inspect()
        inspect_sample.metadata = dict(inspect_sample.metadata or {})
        inspect_sample.metadata["system_prompt"] = composed[sample.role]
        entries.append(inspect_sample)

    return Task(
        dataset=MemoryDataset(entries),
        solver=[system_message("{system_prompt}"), generate()],
    )


def agent_proxy_configured() -> bool:
    """Inspect reads <PROVIDER>_BASE_URL, so the proxy is named agentproxy."""
    return bool(os.environ.get("AGENTPROXY_BASE_URL"))
