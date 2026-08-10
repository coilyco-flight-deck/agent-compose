"""Human grading loop. One case per screen, one keystroke per decision.

Grades are appended after every decision, so an interrupted session keeps
everything already graded. See docs/eval-orchestration.md.
"""

from __future__ import annotations

import argparse
import sys
import termios
import time
import tty
from pathlib import Path

import yaml
from rich.console import Console
from rich.panel import Panel
from rich.table import Table

from evalkit.schema import BoardCase, Fit, Grade, Kind, Verdict, grading_order, pair_results

PASS_FAIL_KEYS = {"p": Verdict.PASS, "x": Verdict.FAIL}
PERSONALITY_KEYS = {"f": Fit.FIT, "u": Fit.UNDECIDED, "n": Fit.NO_FIT}

KIND_STYLE = {
    Kind.BOUNDARY: "bright_cyan",
    Kind.ROLE_FIT: "bright_magenta",
    Kind.PERSONALITY: "bright_yellow",
}


def read_key() -> str:
    """Single keypress without a newline, so grading stays one stroke per case."""
    if not sys.stdin.isatty():
        return sys.stdin.read(1) or "q"
    fd = sys.stdin.fileno()
    saved = termios.tcgetattr(fd)
    try:
        tty.setraw(fd)
        return sys.stdin.read(1)
    finally:
        termios.tcsetattr(fd, termios.TCSADRAIN, saved)


def load_board(path: Path) -> list[BoardCase]:
    raw = yaml.safe_load(path.read_text()) or {}
    return [BoardCase.from_dict(entry) for entry in raw.get("board", [])]


def load_grades(path: Path) -> dict[str, Grade]:
    if not path.exists():
        return {}
    raw = yaml.safe_load(path.read_text()) or {}
    grades: dict[str, Grade] = {}
    for entry in raw.get("grades", []):
        verdict = _decode_verdict(str(entry["verdict"]))
        grades[str(entry["id"])] = Grade(
            id=str(entry["id"]), verdict=verdict, note=str(entry.get("note", ""))
        )
    return grades


def save_grades(path: Path, grades: dict[str, Grade]) -> None:
    payload = {"grades": [grades[key].to_dict() for key in sorted(grades)]}
    path.write_text(yaml.safe_dump(payload, sort_keys=False, width=100))


def render(console: Console, case: BoardCase, position: int, total: int, started: float) -> None:
    candidate = case.candidate
    console.clear()

    header = Table.grid(padding=(0, 2))
    header.add_column(style="bold")
    header.add_column()
    header.add_row("case", f"[bold]{candidate.id}[/bold]")
    header.add_row("role", candidate.role)
    header.add_row("kind", f"[{KIND_STYLE[candidate.kind]}]{candidate.kind.value}[/]")
    if candidate.boundary:
        half = candidate.half.value if candidate.half else ""
        header.add_row("boundary", f"{candidate.boundary} ({half})")
    if candidate.target:
        header.add_row("target", candidate.target)
    if candidate.trait:
        header.add_row("trait", candidate.trait)
    header.add_row("progress", _progress(position, total, started))
    console.print(header)
    console.print()

    console.print(Panel(candidate.prompt, title="prompt", border_style="dim"))
    words = len(case.response.split())
    console.print(
        Panel(
            case.response,
            title=f"response ({words} words, cap {candidate.word_cap})",
            border_style=KIND_STYLE[candidate.kind],
        )
    )
    console.print(Panel(candidate.expected, title="passing looks like", border_style="green"))

    if candidate.scored_pass_fail:
        console.print(
            "[bold]p[/bold] pass    [bold]x[/bold] fail    "
            "[bold]s[/bold] skip    [bold]q[/bold] quit"
        )
    else:
        console.print(
            "[bold]f[/bold] fit    [bold]u[/bold] undecided    "
            "[bold]n[/bold] does not fit    [bold]s[/bold] skip    [bold]q[/bold] quit"
        )


def grade_session(board: list[BoardCase], grades: dict[str, Grade], out: Path) -> bool:
    """Returns False when the grader quit before finishing."""
    console = Console()
    pending = [case for case in grading_order(board) if case.id not in grades]
    started = time.monotonic()

    for index, case in enumerate(pending, start=1):
        keys = PASS_FAIL_KEYS if case.candidate.scored_pass_fail else PERSONALITY_KEYS
        while True:
            render(console, case, index, len(pending), started)
            key = read_key().lower()
            if key == "q":
                return False
            if key == "s":
                break
            if key in keys:
                verdict = keys[key]
                note = ""
                grade = Grade(id=case.id, verdict=verdict)
                if grade.is_deduction:
                    console.print()
                    note = console.input("[bold]note[/bold] (why it missed): ").strip()
                grades[case.id] = Grade(id=case.id, verdict=verdict, note=note)
                save_grades(out, grades)
                break
    return True


def summarize(console: Console, board: list[BoardCase], grades: dict[str, Grade]) -> None:
    pairs = pair_results(board, grades)
    if pairs:
        table = Table(title="boundary pairs, the scoring unit")
        table.add_column("pair")
        table.add_column("role")
        table.add_column("boundary")
        table.add_column("result")
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
    for grade in grades.values():
        counts[grade.verdict.value] = counts.get(grade.verdict.value, 0) + 1
    console.print(
        f"\ngraded {len(grades)} of {len(board)}: "
        + ", ".join(f"{value} {key}" for key, value in sorted(counts.items()))
    )


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Grade the behavior board by hand.")
    parser.add_argument("--board", type=Path, required=True)
    parser.add_argument("--out", type=Path, required=True)
    parser.add_argument("--summary", action="store_true", help="print results and exit")
    args = parser.parse_args(argv)

    board = load_board(args.board)
    grades = load_grades(args.out)
    console = Console()

    if not args.summary:
        finished = grade_session(board, grades, args.out)
        if not finished:
            console.print("\n[yellow]stopped early, grades saved[/yellow]")

    summarize(console, board, grades)
    return 0


def _progress(position: int, total: int, started: float) -> str:
    elapsed = time.monotonic() - started
    if position <= 1:
        return f"{position}/{total}"
    rate = elapsed / (position - 1)
    remaining = rate * (total - position + 1)
    return f"{position}/{total}  {elapsed / 60:.0f}m elapsed  ~{remaining / 60:.0f}m left"


def _decode_verdict(value: str) -> Verdict | Fit:
    try:
        return Verdict(value)
    except ValueError:
        return Fit(value)


if __name__ == "__main__":
    raise SystemExit(main())
