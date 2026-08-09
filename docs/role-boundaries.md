# Role boundaries

A boundary is one shared doctrine body that any number of roles activate
identically. Use one when the same obligation binds several roles and restating
it per charter would let the copies drift apart. A body every role declares is
a roster-wide rule, not a boundary.

## Why boundaries exist

Role skill bodies have a 400-word ceiling. Before boundaries, every role sharing the
communication-ownership or live-operations boundary restated it in its own
charter, so shared policy competed with role-specific prose for one budget.
Extracting both freed 602 words across the Core Roster, a 22% reduction, with no
doctrine removed. A boundary body is bounded separately and never enters the briefing.

## How a boundary differs from its neighbours

* Personality - shared and eager like a boundary, but carries disposition, a color, and identity primitives. A personality never overrides a role obligation.
* Method - progressive disclosure like an ordinary skill, but owned by exactly one role and inactive until role and task both match.
* Boundary - shared by many roles, eagerly active, and carries obligation rather than disposition or procedure.

## Package layout

Declare the catalog entry in a library, then reference it from each role:

```kdl
boundary "live-ops" skill="boundary-live-ops" owner="ops" summary="observation is approved, mutation and promotion are not"
```

```kdl
role "engineer" {
    skill "role-engineer"
    boundary "live-ops" "evidence"
    personality "curious" "meticulous" "tenacious"
}
```

Store each body beside the personality definitions:

```text
libraries/kai-core/boundaries/01-live-ops.kdl
libraries/kai-core/definitions/skills/boundary-live-ops/SKILL.md
```

The catalog id, `skill` property, and frontmatter name must agree, so boundary
`live-ops` binds `boundary-live-ops`. An unreferenced boundary fails loading.

## Selection and delivery

Agent Compose selects a boundary with every role that declares it. The identity card
names boundary skills under `Shared doctrine` and lists them in `Active doctrine`
between the charter and the personalities, so the agent loads them before acting.
Native delivery installs each body once as an ordinary skill, and compiled
delivery appends the same bodies. Sharing one body is expected, not a collision.

The bundle manifest records the selected boundaries, the decision trace carries one
`boundary:<id>` entry per role, and bundle verification fails when the manifest and
trace disagree.

## Evaluation

Packs carry boundary bodies in a `boundaries` block beside the briefing, so doctrine
that left a charter still reaches the driver rather than scoring an incomplete
role. Changing a body moves the pack digest for every declaring role, retiring
those results until an independently reviewed re-run.

## Core Roster boundaries

* `live-ops` - bound into engineer, qa, and ai, the roles sealed against live mutation. Owner DevOps holds the opposite authority.
* `comms` - bound into design, exec, and ops, the roles holding externally facing content that carries a social tone. Owner Content Creator owns the recommendations they defer.
* `evidence` - bound into engineer, exec, and ops, the roles whose diligence reaches past the context handed to them. No owner, since no role owns the opposite. QA, Designer, Content Creator, Director, and AI Engineer are excluded, each because acquisition would override a designated source or duplicate a rule the charter already carries.

## See also

* [Role skills](role-briefings.md) - charter and progressive-disclosure model.
* [Boundary owners](boundary-owners.md) - the two-sided relationship.
* [Role methods](role-methods.md) - single-owner lazy procedures.
* [Personality libraries](personality-libraries.md) - the shared disposition axis.
* [Role-skill context budget](role-skill-context-budget.md) - measured budget effects.
* [Evaluation](evaluation.md) - deterministic packs and review policy.
