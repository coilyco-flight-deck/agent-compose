# Evaluation

How role and personality behavior is measured here, and the two tools that do it. The board
derives from the roster, runs against one subject, and is graded by a human.

## Two tools, one seam

**`evalkit`** is the Python runner in this repo. It derives the case list from the roster Go
exports, runs the subject, and writes the dataset.

**`aos-eval`** is the grading half, shipped from `coilyco-flight-deck/agentic-os` so the pairing
rule has one home. It holds no runner and no model client, so grading never spends a token and
never touches a deployed system. Run `aos-eval help` for the exhaustive reference. sirens-echo is
the second intended consumer and shares the boundary declaration shape today, not the dataset.

Go owns what a pack is and what a valid record is, and the rule keeping that seam honest is
that **Python consumes what Go emits and never restates it**. No second pack schema, coverage
rule, or record writer, because two parsers is the failure the split avoids.

## The triple

Three parties, none holding two seats. The **generator** authors candidate cases, an agent
working with Kai. The **subject** produces responses, `evaluation/deepseek-v4-flash` through
Agent Proxy at the `commodity` tier. The **grader** scores them, Kai, by hand.

A role's `model-tier` is a deployment compatibility claim rather than what the board tests, and
tier does not change selected context, so one subject measures every role's composed text. See
[model tiers](harness-vendoring.md).

## The pipeline

```text
generator          ->  samples.yaml
inspect eval       ->  .eval log      (five epochs per sample, unscored)
evalkit.filter     ->  dataset.yaml   (epoch 1 attached to every sample)
aos-eval annotate  ->  annotations.yaml
aos-eval taxonomy  ->  failure modes, ranked
aos-eval export    ->  display.json
```

Go turns annotations into the canonical record, so totals and verdicts come from the pack rule
rather than from the annotator.

## The dataset is derived

`evalkit.matrix` reads the roster and prints the cases it implies. Boundaries and their owners
produce the pairs, adjacency produces the role-fit targets, and each role's meld produces the
personality cases, so adding a boundary, flipping an adjacency edge, or swapping a personality
moves the sample list on its own and the dataset cannot drift from the roster. Adjacency
reasons become the role-fit descriptors directly, which is the text a generator needs to build
the right confusion. Execution is one case per session, because batching a role's cases would
save machine time and no human time while manufacturing the reflexive deferral the in-out pair
exists to catch.

## The pair is the scoring unit

A boundary case comes in halves: one inside where the role must own the work, one outside where
it must defer. A role passing one and failing the other is a boundary failure rather than fifty
percent, so a degenerate always-defer policy scores zero instead of perfect conformance, and
`aos-eval` reports pair results rather than half results.

## There is no mechanical scorer

Samples once carried a `discriminator`, a regex list for the failing behavior, and item
analysis used the match count to pick which samples reached the annotator. Across nine
pass-or-fail cases the patterns and the grader agreed on nothing that mattered: one false
positive, two false negatives, six agreements where nothing happened. It measured something,
but not what the grading measures, so it was **deleted rather than tuned**. The model-graded
predecessor went the same way in `agent-compose#262`, its gating lane unable to fail and its
producer and reviewer sharing a model. One authored case per slot now, all annotated, and the
retired records stay under `evaluations/retired-*`. Epochs stay at five, the annotator sees
epoch 1, and the rest stay in the log as a failure-spread estimate at no grading cost.

## Annotation

Boundary and role fit are pass or fail with a 50-word cap. Personality is fit, undecided, or
does not fit, with a 100-word cap. Notes are recorded only on a deduction. Caps were measured:
below roughly 25 words a deferring half drops the factual handoff its boundary requires, and 50
words fit one slide at large type. `undecided` is a signal rather than a hedge, so a cluster of
it is item analysis for the one tier with no mechanical filter.

Samples are ordered **role-major**, so an annotator loads one charter and holds it across that
role's samples. `--roster` prints purpose, boundaries, adjacency reasons, and personalities
once per group, `--role` annotates a subset, and grading saves after every decision. A
deduction records a critique and, where one exists, a verbatim span from the output, verified
before it is accepted. `aos-eval taxonomy` is the axial step: it groups deductions by
structural axis, then by shared critique terms, and ranks by frequency, producing a list of
failure modes rather than a score.

## Export is one way

`aos-eval export` projects a committed run into a display payload and nothing returns, so the
surface reading it is a rebuildable projection rather than a second home for evidence. Pairs
travel as their own structure, so a renderer gets `complete` and `passed` rather than a
half-graded pair wrong, and `critique` and `evidence` stay out unless `--include-private` asks
for them. Export **refuses rather than scrubs** when a record looks like it carries a secret,
because a scrubber that misses a pattern ships the secret, and withheld text is not scanned
since text that never leaves cannot leak.

## Commands

Through `just`: `evalkit-matrix` prints the case list, `evalkit-run` runs the subject through
Inspect, `evalkit-filter` writes the dataset, `evalkit-annotate` grades it, `evalkit-export
<run>` emits the payload, `evalkit-check` runs ruff, mypy, and pytest.

## References

* **CheckList**, ACL 2020, <https://arxiv.org/abs/2005.04118> - a capability by test-type
  matrix. The dataset is that shape and the in-out pair is its Directional Expectation test.
* **RULERS**, January 2026, <https://arxiv.org/abs/2601.08654> - immutable criteria, evidence
  anchoring, scale calibration, all three already here.
* **Inspect**, <https://inspect.aisi.org.uk/> - dataset, task, solver, scorer, adopted for the
  run leg, so `evalkit/task.py` is an Inspect task.
* **Practitioner writing**, <https://hamel.dev/blog/posts/evals-faq/> and
  <https://newsletter.pragmaticengineer.com/p/evals> - both argue for binary pass or fail, one
  expert annotator, and a custom annotation surface, which is why no platform was adopted for
  grading.
