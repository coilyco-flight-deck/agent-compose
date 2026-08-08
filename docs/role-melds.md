# Role melds

A meld is one shared doctrine body that any number of roles activate
identically. Use a meld when the same obligation binds several roles and
restating it per charter would let the copies drift apart.

## Why melds exist

Role skill bodies have a 400-word ceiling. Before melds, every role sharing the
communication-ownership or live-operations boundary restated it in its own
charter, so shared policy competed with role-specific prose for one budget.
Extracting both boundaries freed 602 words across the Core Roster, a 22%
reduction, without removing doctrine from any role. A meld body is bounded
separately at 400 words and never enters `Role.Briefing`.

## How a meld differs from its neighbours

* Personality - shared and eager like a meld, but carries disposition, a color, and identity primitives. A personality never overrides a role obligation.
* Method - progressive disclosure like an ordinary skill, but owned by exactly one role and inactive until role and task both match.
* Meld - shared by many roles, eagerly active, and carries obligation rather than disposition or procedure.

## Package layout

Declare the catalog entry in a library, then reference it from each role:

```kdl
meld "live-ops" skill="meld-live-ops" summary="observation is approved, mutation and promotion are not"
```

```kdl
role "engineer" {
    skill "role-engineer"
    meld "live-ops" "comms"
    personality "curious" "meticulous" "tenacious"
}
```

Store each body beside the personality definitions:

```text
libraries/kai-core/melds/01-live-ops.kdl
libraries/kai-core/definitions/skills/meld-live-ops/SKILL.md
```

The catalog id, `skill` property, and frontmatter name must agree, so meld
`live-ops` binds skill `meld-live-ops`. An unreferenced meld fails loading.

## Selection and delivery

Agent Compose selects a meld with every role that declares it. The identity card
names melded skills under `Shared doctrine` and lists them in `Active doctrine`
between the role charter and the personalities, so the agent loads them before
acting. Native delivery installs each body once as an ordinary skill, and
compiled delivery appends the same selected bodies. Several roles sharing one
body is the expected case rather than a collision.

The bundle manifest records the selected melds, the decision trace carries one
`meld:<id>` entry per role, and bundle verification fails when the manifest and
trace disagree.

## Evaluation

Evaluation packs carry melded bodies in a `melds` block beside the briefing, so
doctrine that left a charter still reaches the driver rather than scoring an
incomplete role. Changing a melded body moves the pack digest for every
declaring role, retiring those results until an independently reviewed re-run.

## Core Roster melds

* `live-ops` - melded into engineer, qa, and ai, the roles sealed against live mutation. DevOps owns the opposite authority and does not meld it.
* `comms` - melded into every role except creator, which owns the other side of the boundary.
* `evidence` - melded into every role. Acquiring the source that settles a claim precedes ranking the evidence already held, and no role is exempt from opening it.

## See also

* [Role skills](role-briefings.md) - charter and progressive-disclosure model.
* [Role methods](role-methods.md) - single-owner lazy procedures.
* [Personality libraries](personality-libraries.md) - the shared disposition axis.
* [Role-skill context budget](role-skill-context-budget.md) - measured budget effects.
* [Evaluation](evaluation.md) - deterministic packs and review policy.
