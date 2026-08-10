# Eval grading

How a case is scored, in what order, and which candidates earn a place on the
board. The pipeline that produces them is in
[eval orchestration](eval-orchestration.md).

## Scoring tiers

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

`grade.py` reports pair results, never half results.

## Grading is role-major

Cases are ordered by role, so a grader loads one charter and holds it across
all of that role's cases. `--roster` prints purpose, owned and deferred
boundaries, adjacency reasons, and personalities once per group.

Kind-major degrades more gracefully, leaving every role partly scored if a
session stops early. Role context is the expensive thing to reload and grading
is resumable, so speed wins. `--role` grades a subset when a session needs one.

## Item analysis

Every candidate runs five times. One that passes all five measures nothing, and
one that fails all five is broken rather than hard, so both are dropped. Two
candidates compete per slot and the one closest to the midpoint wins.

A boundary pair whose halves behave identically across all five runs is a
broken pair, or its bundle is, so it yields to the other candidate pair.

Every drop is reported, because silent truncation reads as full coverage.

The grader sees run 1. The other four supply a failure-spread estimate at no
human cost, which is what answers the single-sample gap in the retired
baseline.

## See also

* [Eval orchestration](eval-orchestration.md) - the pipeline and its seam.
* [Evaluation](evaluation.md) - packs, records, and review policy.
