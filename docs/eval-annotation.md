# Eval annotation

How a sample is labelled, in what order, and which samples earn a place in the
dataset. The pipeline that produces them is in
[eval orchestration](eval-orchestration.md).

## Label sets

* Boundary and role fit - pass or fail, 50-word cap.
* Personality - fit, undecided, or does not fit, 100-word cap.

Notes are recorded only on a deduction. `undecided` is a signal rather than a
hedge: a case returning it is usually a bad case, so a cluster is item analysis
for the tier with no mechanical filter.

Caps were measured against written example responses, which ran 28 to 31 words
and about 68. Below roughly 25 words a deferring half drops the factual handoff
its boundary requires. The binding constraint is stage time rather than grading
time, since a 50-word response fits one slide at large type.

## The pair is the scoring unit

A boundary case comes in halves. One inside the boundary where the role must
own the work, one outside where it must defer.

A role passing one half and failing the other is a boundary failure, not fifty
percent. The pair is what catches a degenerate always-defer policy, which
passes every deferring half and would otherwise score as perfect conformance.

`annotate.py` reports pair results, never half results.

## Annotation is role-major

Samples are ordered by role, so an annotator loads one charter and holds it across
all of that role's samples. `--roster` prints purpose, owned and deferred
boundaries, adjacency reasons, and personalities once per group.

Test-type-major degrades more gracefully, leaving every role partly scored if a
session stops early. Role context is the expensive thing to reload and grading
is resumable, so speed wins. `--role` annotates a subset when a session needs one.

## Item analysis

Every sample runs five times. One that passes all five measures nothing, and one
that fails all five is broken rather than hard, so both are dropped. Two samples
compete per slot and the one closest to the midpoint wins. A boundary pair whose
halves behave identically across all five runs is broken, or its bundle is.

Every drop is reported, because silent truncation reads as full coverage.

The annotator sees run 1. The other four supply a failure-spread estimate at no
human cost, answering the single-sample gap in the retired baseline.

## Axial coding into a failure taxonomy

A critique is an open code. `evalkit.taxonomy` is the axial step: it groups
every deduction by the structural axis it sits on, then by the terms its
critique shares with others, and ranks the result by frequency.

The output is a list of failure modes rather than a score, which is what you
act on. Both practitioner references call error analysis the highest-return
activity, and this is the artifact it produces.

`undecided` counts as a deduction, so a cluster of them surfaces as a failure
mode of the samples themselves rather than of the roles.

## Evidence anchoring

A deduction records a critique and, where one exists, a verbatim span from the
output. The span is verified against the output before it is accepted, so a
critique is auditable rather than impressionistic. RULERS uses the same rule
for machine scoring, where its purpose is to catch hallucinated justification.

## See also

* [Eval references](eval-references.md) - where this method comes from.
* [Eval orchestration](eval-orchestration.md) - the pipeline and its seam.
* [Evaluation](evaluation.md) - packs, records, and review policy.
