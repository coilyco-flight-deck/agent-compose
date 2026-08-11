"""Human annotation loop. One sample per screen, one keystroke per decision.

Annotations are appended after every decision, so an interrupted session keeps
everything already annotated. See docs/eval-grading.md.
"""

from __future__ import annotations

import argparse
import json
import sys
import termios
import time
import tty
from pathlib import Path
from typing import Any

import yaml
from rich.console import Console
from rich.panel import Panel
from rich.table import Table

from evalkit.schema import (
    LABEL_SETS,
    Annotation,
    DatasetEntry,
    Fit,
    TestType,
    Verdict,
    annotation_order,
    pair_results,
)

TYPE_STYLE = {
    TestType.BOUNDARY: "bright_cyan",
    TestType.ROLE_FIT: "bright_magenta",
    TestType.PERSONALITY: "bright_yellow",
}

LABEL_HELP = {
    Verdict.PASS: "pass",
    Verdict.FAIL: "fail",
    Fit.FIT: "fit",
    Fit.UNDECIDED: "undecided",
    Fit.NO_FIT: "does not fit",
}


def read_key() -> str:
    """Single keypress without a newline, so annotation stays one stroke."""
    if not sys.stdin.isatty():
        return sys.stdin.read(1) or "q"
    fd = sys.stdin.fileno()
    saved = termios.tcgetattr(fd)
    try:
        tty.setraw(fd)
        return sys.stdin.read(1)
    finally:
        termios.tcsetattr(fd, termios.TCSADRAIN, saved)


def load_dataset(path: Path) -> list[DatasetEntry]:
    raw = yaml.safe_load(path.read_text()) or {}
    return [DatasetEntry.from_dict(entry) for entry in raw.get("dataset", [])]


def load_annotations(path: Path) -> dict[str, Annotation]:
    if not path.exists():
        return {}
    raw = yaml.safe_load(path.read_text()) or {}
    annotations: dict[str, Annotation] = {}
    for entry in raw.get("annotations", []):
        annotations[str(entry["id"])] = Annotation(
            id=str(entry["id"]),
            label=_decode_label(str(entry["label"])),
            critique=str(entry.get("critique", "")),
            evidence=str(entry.get("evidence", "")),
        )
    return annotations


def save_annotations(path: Path, annotations: dict[str, Annotation]) -> None:
    payload = {"annotations": [annotations[key].to_dict() for key in sorted(annotations)]}
    path.write_text(yaml.safe_dump(payload, sort_keys=False, width=100))


def role_header(console: Console, roster: dict[str, Any], role: str) -> None:
    """Printed once per role group, so the charter is loaded rather than implied."""
    spec = roster.get("roles", {}).get(role)
    if not spec:
        return
    lines = [f"[bold]{spec.get('display_name', role)}[/bold]  {spec.get('purpose', '')}"]
    owned = [
        name
        for name, boundary in roster.get("boundaries", {}).items()
        if boundary.get("owner") == role
    ]
    if owned:
        lines.append("owns: " + ", ".join(owned))
    if spec.get("boundaries"):
        lines.append("defers: " + ", ".join(spec["boundaries"]))
    if spec.get("personalities"):
        lines.append("personalities: " + ", ".join(spec["personalities"]))
    for adjacent in spec.get("adjacents", []):
        lines.append(f"adjacent {adjacent['role']}: {adjacent['reason']}")
    console.print(Panel("\n".join(lines), border_style="bright_white", title="role"))


def render(
    console: Console,
    entry: DatasetEntry,
    position: int,
    total: int,
    started: float,
    roster: dict[str, Any] | None = None,
) -> None:
    sample = entry.sample
    console.clear()
    if roster:
        role_header(console, roster, sample.role)

    header = Table.grid(padding=(0, 2))
    header.add_column(style="bold")
    header.add_column()
    header.add_row("sample", f"[bold]{sample.id}[/bold]")
    header.add_row("role", sample.role)
    header.add_row("test type", f"[{TYPE_STYLE[sample.test_type]}]{sample.test_type.value}[/]")
    if sample.boundary:
        half = sample.half.value if sample.half else ""
        header.add_row("boundary", f"{sample.boundary} ({half})")
    if sample.against:
        header.add_row("against", sample.against)
    if sample.trait:
        header.add_row("trait", sample.trait)
    header.add_row("progress", _progress(position, total, started))
    console.print(header)
    console.print()

    console.print(Panel(sample.prompt, title="prompt", border_style="dim"))
    words = len(entry.output.split())
    console.print(
        Panel(
            entry.output,
            title=f"output ({words} words, cap {sample.word_cap})",
            border_style=TYPE_STYLE[sample.test_type],
        )
    )
    console.print(Panel(sample.target, title="target", border_style="green"))

    keys = LABEL_SETS[sample.label_set]
    hints = "    ".join(f"[bold]{key}[/bold] {LABEL_HELP[label]}" for key, label in keys.items())
    console.print(f"{hints}    [bold]s[/bold] skip    [bold]q[/bold] quit")


def collect_evidence(console: Console, output: str) -> tuple[str, str]:
    """RULERS anchors a deduction to a verbatim span, verified against the output."""
    critique = console.input("[bold]critique[/bold] (what it did wrong): ").strip()
    while True:
        evidence = console.input("[bold]evidence[/bold] (verbatim quote, blank to skip): ").strip()
        if not evidence or evidence.lower() in output.lower():
            return critique, evidence
        console.print("[red]not found verbatim in the output[/red]")


def annotate_session(
    dataset: list[DatasetEntry],
    annotations: dict[str, Annotation],
    out: Path,
    roster: dict[str, Any] | None = None,
) -> bool:
    """Returns False when the annotator quit before finishing."""
    console = Console()
    role_order = list(roster.get("role_order", [])) if roster else None
    pending = [e for e in annotation_order(dataset, role_order) if e.id not in annotations]
    started = time.monotonic()

    for index, entry in enumerate(pending, start=1):
        keys = LABEL_SETS[entry.sample.label_set]
        while True:
            render(console, entry, index, len(pending), started, roster)
            key = read_key().lower()
            if key == "q":
                return False
            if key == "s":
                break
            if key in keys:
                label = keys[key]
                critique, evidence = "", ""
                if label in {Verdict.FAIL, Fit.NO_FIT, Fit.UNDECIDED}:
                    console.print()
                    critique, evidence = collect_evidence(console, entry.output)
                annotations[entry.id] = Annotation(
                    id=entry.id, label=label, critique=critique, evidence=evidence
                )
                save_annotations(out, annotations)
                break
    return True


def summarize(
    console: Console, dataset: list[DatasetEntry], annotations: dict[str, Annotation]
) -> None:
    pairs = pair_results(dataset, annotations)
    if pairs:
        table = Table(title="boundary pairs, the scoring unit")
        for column in ("pair", "role", "boundary", "result"):
            table.add_column(column)
        for pair in pairs:
            if not pair.complete:
                result = "[yellow]incomplete[/yellow]"
            elif pair.passed:
                result = "[green]pass[/green]"
            else:
                result = "[red]fail[/red]"
            table.add_row(pair.pair_id, pair.role, pair.boundary, result)
        console.print(table)

    counts: dict[str, int] = {}
    for annotation in annotations.values():
        counts[annotation.label.value] = counts.get(annotation.label.value, 0) + 1
    console.print(
        f"\nannotated {len(annotations)} of {len(dataset)}: "
        + ", ".join(f"{value} {key}" for key, value in sorted(counts.items()))
    )


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Annotate the eval dataset by hand.")
    parser.add_argument("--dataset", type=Path, required=True)
    parser.add_argument("--out", type=Path, required=True)
    parser.add_argument("--roster", type=Path, help="person.json, for the per-role header")
    parser.add_argument("--role", action="append", help="annotate only these roles, repeatable")
    parser.add_argument("--summary", action="store_true", help="print results and exit")
    args = parser.parse_args(argv)

    dataset = load_dataset(args.dataset)
    if args.role:
        dataset = [e for e in dataset if e.sample.role in set(args.role)]
    annotations = load_annotations(args.out)
    roster = json.loads(args.roster.read_text()) if args.roster else None
    console = Console()

    if not args.summary and not annotate_session(dataset, annotations, args.out, roster):
        console.print("\n[yellow]stopped early, annotations saved[/yellow]")

    summarize(console, dataset, annotations)
    return 0


def _progress(position: int, total: int, started: float) -> str:
    elapsed = time.monotonic() - started
    if position <= 1:
        return f"{position}/{total}"
    rate = elapsed / (position - 1)
    remaining = rate * (total - position + 1)
    return f"{position}/{total}  {elapsed / 60:.0f}m elapsed  ~{remaining / 60:.0f}m left"


def _decode_label(value: str) -> Verdict | Fit:
    try:
        return Verdict(value)
    except ValueError:
        return Fit(value)


if __name__ == "__main__":
    raise SystemExit(main())
