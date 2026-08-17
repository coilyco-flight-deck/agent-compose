# Evaluation

What the evaluation surface is and how a run is annotated.

## Behavior evaluation

How role and personality behavior is measured. The board is derived from the
roster, run against one subject, and graded by a human.

### The triple

Three parties, none holding two seats.

* **Generator** - authors candidate cases. Currently an agent working with Kai.
* **Subject** - produces responses. `evaluation/deepseek-v4-flash` through
  Agent Proxy, at the `commodity` tier.
* **Grader** - scores them. Kai, by hand.

A role's own `model-tier` declaration is a deployment compatibility claim and is
not what the board tests. Model tier does not change selected context, so one
subject measures the composed text for every role. See
[model-tiers.md](harness-vendoring.md).

### What replaced the driver and reviewer

An earlier system drove cases with one model and scored them with another, then
recorded rubric totals against a digest-bound pack. It was retired for three
reasons recorded in `agent-compose#262`: the release-gating lane could not fail,
the lane judged itself because producer and reviewer shared a model, and no
measured agreement with a human grader existed.

A regex tier survived into the replacement and was deleted too, after the first
human-graded board showed it disagreeing with the grader on every case where
either of them deviated from a pass. There is now no mechanical scorer anywhere
in the loop.

Retired records are preserved rather than deleted. `evaluations/retired-*`
holds the last baseline earned under the old contract, and
`evaluations/pilot/` holds graded pilot boards and probes.

### Running it

`evalkit` owns the pipeline and [eval orchestration](eval-pipeline.md)
describes it. [Eval annotation](evaluation.md) covers labels, ordering, and
what reaches the grader.

## Eval annotation

How a sample is labelled, in what order, and which candidate reaches the
dataset. The pipeline is in [eval orchestration](eval-pipeline.md).

### Label sets

* Boundary and role fit - pass or fail, 50-word cap.
* Personality - fit, undecided, or does not fit, 100-word cap.

Notes are recorded only on a deduction. `undecided` is a signal rather than a
hedge: a case returning it is usually a bad case, so a cluster is item analysis
for the one tier with no mechanical filter.

Caps were measured against example responses running 28 to 31 words and about
68. Below roughly 25 words a deferring half drops the factual handoff its
boundary requires. Stage time binds, since 50 words fit one slide at large type.

### The pair is the scoring unit

A boundary case comes in halves: one inside the boundary where the role must own
the work, one outside where it must defer. A role passing one and failing the
other is a boundary failure, not fifty percent. The pair catches a degenerate
always-defer policy that would otherwise score as perfect conformance.
`annotate.py` reports pair results, never half results.

### Annotation is role-major

Samples are ordered by role, so an annotator loads one charter and holds it
across all of that role's samples. `--roster` prints purpose, boundaries,
adjacency reasons, and personalities once per group.

Test-type-major degrades more gracefully, leaving every role partly scored if a
session stops early. Role context is the expensive thing to reload and grading
is resumable, so speed wins. `--role` annotates a subset.

### Every authored case is graded

There is no mechanical scorer and no item analysis. One authored case per slot,
all of them annotated. The only sample a run drops is one the subject never
answered, and that drop is reported.

The regex tier that used to select cases was deleted after the first graded
board, where it disagreed with the grader on every case either of them failed.
[Eval orchestration](eval-pipeline.md) carries that evidence.

The annotator sees epoch 1. The other four stay in the Inspect log, so a reader
can check whether an answer was stable across runs without anything scoring it
for them.

### Axial coding into a failure taxonomy

A critique is an open code. `evalkit.taxonomy` is the axial step: it groups
every deduction by its structural axis, then by shared critique terms, and ranks
by frequency. The output is a list of failure modes rather than a score, which
is what you act on, and both practitioner references call error analysis the
highest-return activity.

`undecided` counts as a deduction, so a cluster of them surfaces as a failure
mode of the samples themselves rather than of the roles.

### Evidence anchoring

A deduction records a critique and, where one exists, a verbatim span from the
output, verified before it is accepted, so a critique is auditable rather than
impressionistic. RULERS uses the same rule to catch hallucinated justification.
