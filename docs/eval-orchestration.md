# Eval orchestration

`evalkit` is the Python half of the evaluation system. Go owns what a pack is
and what a valid record is. Python owns running the subject, filtering
candidates, and putting a case in front of a human.

## Why the seam sits here

Orchestration changes often and is throwaway. Contracts change rarely and are
load-bearing. The rule that keeps the seam honest: **Python consumes what Go
emits and never restates it.** No second pack schema, no second coverage rule,
no second record writer. Two parsers is the failure this split avoids.

## Pipeline

```text
generator       ->  candidates.yaml
evalkit.run     ->  responses.jsonl   (n=5 per candidate, transport recorded)
evalkit.filter  ->  board.yaml        (survivors, with run 1 attached)
evalkit.grade   ->  grades.yaml       (one human decision per case)
```

Go turns grades into the canonical record, so totals and verdicts come from the
pack rule rather than from the grader.

## The board is derived

`evalkit.board` reads the roster Go exports and prints the cases it implies.
Boundaries and their owners produce the pairs, adjacency produces the role-fit
targets, and each role's meld produces the personality cases.

Adding a boundary, flipping an adjacency edge, or swapping a personality moves
the case list on its own, so the board cannot drift from the roster. Adjacency
reasons become the role-fit descriptors directly, which is the same text a
generator needs to construct the right confusion.

## Execution is one case per session

Batching a role's cases into one request would save background machine time and
no human time, while manufacturing the reflexive deferral the in-out pair
exists to catch, breaking the per-case independence n=5 assumes, and adding
order effects.

Whether doctrine survives accumulated context is a real question, but it is a
second arm rather than a cheaper version of this one.

## Transport

`evalkit.run` records the transport on every response. Agent Proxy is the only
transport that produces a measured result. `--direct` exists for incident
isolation, marks its output, and warns, because Agent Proxy sits inside the
transport path and a direct call is a different configuration.

## Open decision

`substring_matcher` is a placeholder, since a prose discriminator cannot be
matched reliably by substring. The options are machine-checkable patterns
emitted alongside the prose, or a cheap model pass. A model is acceptable here
where it would not be for grading, because a filter error costs a slightly
worse case rather than a wrong score.

## Commands

Ward owns `evalkit-sync`, `evalkit-run`, `evalkit-filter`, `evalkit-grade`,
`evalkit-board`, and `evalkit-check`. The last runs ruff, format, mypy strict,
and pytest from `scripts/evalkit-check.sh` rather than pre-commit, because that
file is managed by agentic-os and a hand-added hook is overwritten on sync.

## See also

* [Eval grading](eval-grading.md) - scoring tiers, ordering, and item analysis.
* [Evaluation](evaluation.md) - packs, records, and review policy.
* [Role adjacency](role-adjacency.md) - the axis role-fit cases read.
