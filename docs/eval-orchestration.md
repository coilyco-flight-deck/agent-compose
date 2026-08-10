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
generator  ->  candidates.yaml
evalkit.run     ->  responses.jsonl   (n=5 per candidate, transport recorded)
evalkit.filter  ->  board.yaml        (survivors, with run 1 attached)
evalkit.grade   ->  grades.yaml       (one human decision per case)
```

Go turns grades into the canonical record, so totals and verdicts come from the
pack rule rather than from the grader.

## Scoring tiers

* Boundary and role fit - pass or fail, 50-word cap.
* Personality - fit, undecided, or does not fit, 100-word cap.

`undecided` is a signal rather than a hedge. A case returning it is usually a bad
case, so a cluster is item analysis for the tier with no mechanical filter.

## The pair is the scoring unit

A boundary case comes in halves. One inside the boundary where the role must
own the work, one outside where it must defer. A role passing one half and
failing the other is a boundary failure, not fifty percent, because the pair is
what catches a degenerate always-defer policy.

`grade.py` reports pair results, never half results.

## Item analysis

Every candidate runs five times. One that passes all five measures nothing, and
one that fails all five is broken rather than hard, so both are dropped. Two
candidates compete per slot and the one closest to the midpoint wins. Every
drop is reported, because silent truncation reads as full coverage.

The grader sees run 1. The other four supply a failure-spread estimate at no
human cost.

## Transport

`evalkit.run` records the transport on every response. Agent Proxy is the only
transport that produces a measured result. `--direct` exists for incident
isolation, marks its output, and warns, because Agent Proxy sits inside the
transport path and a direct call is a different configuration.

## Open decision

`substring_matcher` is a placeholder. A prose discriminator cannot be matched
reliably by substring. The two real options are machine-checkable patterns
emitted alongside the prose, or a cheap model pass. A model here is acceptable
where it would not be for grading, since a filter error costs a slightly worse
case rather than a wrong score.

## Commands

Ward owns the verbs: `evalkit-sync`, `evalkit-run`, `evalkit-filter`,
`evalkit-grade`, and `evalkit-check`.

`evalkit-check` runs ruff, format, mypy strict, and pytest from
`scripts/evalkit-check.sh` rather than pre-commit, because that file is managed
by agentic-os and a hand-added hook there is overwritten on the next sync.

## See also

* [Evaluation](evaluation.md) - packs, records, and review policy.
* [Role adjacency](role-adjacency.md) - the axis role-fit cases read.
