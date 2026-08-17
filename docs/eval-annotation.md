# Eval annotation

How a sample is labelled, in what order, and which candidate reaches the
dataset. The pipeline is in [eval orchestration](eval-orchestration.md).

## Label sets

* Boundary and role fit - pass or fail, 50-word cap.
* Personality - fit, undecided, or does not fit, 100-word cap.

Notes are recorded only on a deduction. `undecided` is a signal rather than a
hedge: a case returning it is usually a bad case, so a cluster is item analysis
for the one tier with no mechanical filter.

Caps were measured against example responses running 28 to 31 words and about
68. Below roughly 25 words a deferring half drops the factual handoff its
boundary requires. Stage time binds, since 50 words fit one slide at large type.

## The pair is the scoring unit

A boundary case comes in halves: one inside the boundary where the role must own
the work, one outside where it must defer. A role passing one and failing the
other is a boundary failure, not fifty percent. The pair catches a degenerate
always-defer policy that would otherwise score as perfect conformance.
`annotate.py` reports pair results, never half results.

## Annotation is role-major

Samples are ordered by role, so an annotator loads one charter and holds it
across all of that role's samples. `--roster` prints purpose, boundaries,
adjacency reasons, and personalities once per group.

Test-type-major degrades more gracefully, leaving every role partly scored if a
session stops early. Role context is the expensive thing to reload and grading
is resumable, so speed wins. `--role` annotates a subset.

## Every authored case is graded

There is no mechanical scorer and no item analysis. One authored case per slot,
all of them annotated. The only sample a run drops is one the subject never
answered, and that drop is reported.

The regex tier that used to select cases was deleted after the first graded
board, where it disagreed with the grader on every case either of them failed.
[Eval orchestration](eval-orchestration.md) carries that evidence.

The annotator sees epoch 1. The other four stay in the Inspect log, so a reader
can check whether an answer was stable across runs without anything scoring it
for them.

## Axial coding into a failure taxonomy

A critique is an open code. `aos-eval taxonomy` is the axial step: it groups
every deduction by its structural axis, then by shared critique terms, and ranks
by frequency. The output is a list of failure modes rather than a score, which
is what you act on, and both practitioner references call error analysis the
highest-return activity.

`undecided` counts as a deduction, so a cluster of them surfaces as a failure
mode of the samples themselves rather than of the roles.

## Evidence anchoring

A deduction records a critique and, where one exists, a verbatim span from the
output, verified before it is accepted, so a critique is auditable rather than
impressionistic. RULERS uses the same rule to catch hallucinated justification.

## Authoring fails before grading does

The shared `Sample` enforces only what is true of every deployment, so
`evalkit.filter.load_samples` validates against this repo's profile and refuses
a board whose cases would otherwise read as coverage they do not have.

## See also

* [Eval references](eval-references.md) - where this method comes from.
* [Eval orchestration](eval-orchestration.md) - the pipeline and its seam.
* [Evaluation](evaluation.md) - packs, records, and review policy.
