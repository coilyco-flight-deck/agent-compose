# Eval orchestration

`evalkit` is the Python half of the evaluation system. Go owns what a pack is
and what a valid record is. Python owns running the subject, filtering
samples, and putting a case in front of a human.

## Why the seam sits here

Orchestration changes often and is throwaway. Contracts change rarely and are
load-bearing. The rule that keeps the seam honest: **Python consumes what Go
emits and never restates it.** No second pack schema, no second coverage rule,
no second record writer. Two parsers is the failure this split avoids.

## Pipeline

```text
generator         ->  samples.yaml
inspect eval      ->  .eval log      (five epochs per sample, unscored)
evalkit.filter    ->  dataset.yaml   (epoch 1 attached to every sample)
evalkit.annotate  ->  annotations.yaml
evalkit.taxonomy  ->  failure modes, ranked
```

Go turns annotations into the canonical record, so totals and verdicts come from the
pack rule rather than from the annotator.

## The dataset is derived

`evalkit.matrix` reads the roster Go exports and prints the cases it implies.
Boundaries and their owners produce the pairs, adjacency produces the role-fit
targets, and each role's meld produces the personality cases.

Adding a boundary, flipping an adjacency edge, or swapping a personality moves
the sample list on its own, so the dataset cannot drift from the roster.
Adjacency reasons become the role-fit descriptors directly, which is the same
text a generator needs to construct the right confusion.

## Execution is one case per session

Batching a role's cases into one request would save background machine time and
no human time, while manufacturing the reflexive deferral the in-out pair exists
to catch, breaking per-case independence, and adding order effects. Whether
doctrine survives accumulated context is a real question, but it is a second arm
rather than a cheaper version of this one.

## Transport

`evalkit.run` records the transport on every response. Agent Proxy is the only
transport that produces a measured result. `--direct` exists for incident
isolation, marks its output, and warns, because Agent Proxy sits inside the
transport path and a direct call is a different configuration.

## There is no mechanical scorer

Samples carried a `discriminator`, a list of regexes for the failing behaviour,
and item analysis used the match count to pick which samples reached the
annotator. The first human-graded board removed it. Across nine pass-or-fail
cases the patterns and the grader agreed on nothing that mattered: one false
positive, two false negatives, and six agreements on cases where nothing
happened. A follow-up probe repeated it, three detectors disagreeing with each
other and with a reading of the same twenty responses.

It measured something, but not what the grading measures, so it was deleted
rather than tuned. One authored case per slot now, since the match count was
also the only rule for choosing between candidates. Epochs stay at five, with
the annotator seeing epoch 1 and the rest left in the log as evidence.

## Commands

Ward owns `evalkit-sync`, `evalkit-run`, `evalkit-filter`, `evalkit-annotate`,
`evalkit-matrix`, and `evalkit-check`. The last runs ruff, format, mypy strict,
and pytest from `scripts/evalkit-check.sh` rather than pre-commit, because that
file is managed by agentic-os and a hand-added hook is overwritten on sync.

## See also

* [Eval annotation](eval-annotation.md) - scoring tiers, ordering, and item analysis.
* [Evaluation](evaluation.md) - packs, records, and review policy.
* [Role adjacency](role-adjacency.md) - the axis role-fit cases read.
