# Role adjacency

Adjacency names the two roles whose work a role is most likely to absorb, and
why. It is the axis the evaluation board reads to author role-fit cases.

## Directed, not symmetric

Absorption risk runs one way. DevOps sequencing follow-up work after an
incident is a live confusion, while the Director rarely reaches for a runbook.
Declaring that pair symmetrically would buy a case nobody fails.

So a role names who it drifts toward, and the reverse edge is a separate
decision. Do not add a symmetry check.

## Out-degree is fixed at two

Two forces the roster to pick the sharpest confusions rather than list every
plausible neighbour. Every role declares exactly two, or the roster declares
none at all. The all-or-nothing rule keeps external person packages authored
before this axis loading unchanged.

## The reason is generator input

Each edge carries a `reason`. It is not commentary. An adjacency case has to
construct one specific confusion, and a generator handed a bare pair will
invent the wrong one on exactly the edges that matter. An edge nobody would
guess is either a mistake or load-bearing, and the reason is how a reader tells
which.

## Package layout

Declare one node per edge, since each edge carries its own reason:

```kdl
role "ops" {
    skill "role-ops"
    boundary "suggest-human-comms" "seek-external-validation"
    adjacent "engineer" reason="implementing the fix instead of handing it back"
    adjacent "director" reason="sequencing and assigning follow-up work after an incident instead of surfacing it as findings"
    personality "protective" "grounded" "reflective"
}
```

Loading fails on a self-edge, a repeated target, a missing reason, an unknown
target role, or an out-degree other than two.

## How adjacency differs from a boundary

A boundary removes one behavior from several roles and allocates it to an
owner, so it already tests each member against that owner. Adjacency covers the
confusions no boundary allocates.

Spend adjacency slots accordingly. DevOps defers `seek-external-validation` to
the Executive Strategist, so an `ops` edge pointed at `exec` would restate a
case the boundary tier already owns.

## Core Roster graph

```text
engineer -> qa, ai
qa       -> engineer, ai
ops      -> engineer, director
ai       -> engineer, qa
director -> exec, engineer
exec     -> director, creator
design   -> creator, engineer
creator  -> design, exec
```

In-degree is deliberately uneven. Engineer is defended by five edges because it
is the most-absorbed role, and DevOps is defended by none because
`modify-live-system` already seals engineer, qa, and ai against live mutation.

## See also

* [Role boundaries](role-boundaries.md) - shared behavior allocated to one owner.
* [Boundary owners](ownership.md) - the two-sided relationship.
* [Role skills](role-briefings.md) - charter and progressive-disclosure model.
* [Evaluation](evaluation.md) - deterministic packs and review policy.
