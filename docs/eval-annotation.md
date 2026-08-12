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

## Item analysis reports, it does not gate

Every sample runs five times. Two compete per slot and the one closest to the
midpoint wins. That loser, and a sample with no subject output, are the only
things a run drops, and every drop is reported.

Nothing is dropped for failing to discriminate. A pattern cannot see polarity,
so a sample it never fires on may be an easy case or a blind regex, and the
filter cannot tell those apart. Dropping it removes the one thing that could,
the human reading it. A sample outside the one-to-four band is noted instead,
as a lead for the next generation pass rather than a verdict.

**A negative control is exempt.** An in-half and a within-role case exist to
catch a degenerate always-defer policy, so passing every run is the control
working. Scoring that as non-discriminating deleted whole pairs, including the
most informative shape a boundary produces: a role owning its own work and
over-claiming on the far side. A pair whose halves behave identically is still
broken, or its bundle is, and is now noted rather than deleted.

The annotator sees run 1. The other four supply a failure-spread estimate at no
human cost, answering the single-sample gap in the retired baseline.

## Axial coding into a failure taxonomy

A critique is an open code. `evalkit.taxonomy` is the axial step: it groups
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

## See also

* [Eval references](eval-references.md) - where this method comes from.
* [Eval orchestration](eval-orchestration.md) - the pipeline and its seam.
* [Evaluation](evaluation.md) - packs, records, and review policy.
