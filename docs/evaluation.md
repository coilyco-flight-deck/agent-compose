# Behavior evaluation

How role and personality behavior is measured. The board is derived from the
roster, run against one subject, and graded by a human.

## The triple

Three parties, none holding two seats.

* **Generator** - authors candidate cases. Currently an agent working with Kai.
* **Subject** - produces responses. `evaluation/deepseek-v4-flash` through
  Agent Proxy, at the `commodity` tier.
* **Grader** - scores them. Kai, by hand.

A role's own `model-tier` declaration is a deployment compatibility claim and is
not what the board tests. Model tier does not change selected context, so one
subject measures the composed text for every role. See
[model-tiers.md](model-tiers.md).

## What replaced the driver and reviewer

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

## Running it

`evalkit` owns the pipeline and [eval orchestration](eval-orchestration.md)
describes it. [Eval annotation](eval-annotation.md) covers labels, ordering, and
what reaches the grader.

## See also

* [eval-orchestration.md](eval-orchestration.md) - the pipeline and its seam.
* [eval-annotation.md](eval-annotation.md) - scoring tiers and ordering.
* [eval-references.md](eval-references.md) - where the method comes from.
* [role-adjacency.md](role-adjacency.md) - the axis role-fit cases read.
