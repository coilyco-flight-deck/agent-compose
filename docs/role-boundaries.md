# Role boundaries

A boundary is one behavior removed from several roles and allocated to exactly
one owner. Use one when restating that allocation per charter would let the
copies drift apart. A body every role declares is a roster-wide rule instead.

## Why boundaries exist

Role skill bodies have a 400-word ceiling. Before boundaries, every role sharing
an allocation restated it in its own charter, so shared policy competed with
role-specific prose for one budget. Extraction freed 602 words across the Core
Roster with no doctrine lost, and each side is bounded separately.

## How a boundary differs from its neighbours

* Personality - shared and eager too, but carries disposition, a color, and identity primitives, and never overrides a role obligation.
* Method - progressive disclosure, owned by exactly one role, inactive until role and task both match.
* Boundary - one behavior removed from several roles and allocated to exactly one owner.

## Package layout

Declare the catalog entry in a library, then reference it from each role:

```kdl
boundary "modify-live-system" skill="boundary-modify-live-system" owner="ops" summary="DevOps changes running systems, other roles observe and hand the action over"
```

```kdl
role "engineer" {
    skill "role-engineer"
    boundary "modify-live-system" "seek-external-validation"
    personality "curious" "meticulous" "tenacious"
}
```

Store each body beside the personality definitions:

```text
libraries/kai-core/boundaries/01-modify-live-system.kdl
libraries/kai-core/definitions/skills/boundary-modify-live-system/SKILL.md
```

The catalog id, `skill` property, and frontmatter name must agree, so boundary
`modify-live-system` binds `boundary-modify-live-system`. An unreferenced boundary fails loading.

## Selection and delivery

Agent Compose selects a boundary with every role that declares it and with its
owner. The identity card lists them under `Boundaries`, marking each as owned or
deferred, and repeats them in `Active doctrine` so the agent loads them before
acting. Sharing one body is expected, not a collision.

The bundle manifest records the selected boundaries, the decision trace carries one
`boundary:<id>` entry per role, and bundle verification fails when the manifest and
trace disagree.

## Evaluation

Packs carry boundary bodies in a `boundaries` block beside the briefing, so
doctrine that left a charter still reaches the driver rather than scoring an
incomplete role. Both sides receive the body, so both owe its case. Changing a
body moves the pack digest for every role on either side, retiring those results
until an independently reviewed re-run.

## Core Roster boundaries

Each slug names the behavior that moves, not the domain it sits in.

* `modify-live-system`, owner DevOps - deferred by engineer, qa, ai, gamedev, and exec, the roles sealed against live mutation. Director keeps it, since directing committed work reaches as far as a release.
* `build-software`, owner Engineer - deferred by exec, director, and creator, the roles that shape what software should do without producing it. Designer neither defers nor owns it, since its charter already grants a bounded visual-only slice.
* `suggest-human-comms`, owner Content Creator - deferred by design, exec, and ops, the roles holding externally facing content that carries a social tone.
* `seek-external-validation`, owner Executive Strategist - deferred by engineer, director, qa, ops, design, and ai. Content Creator neither defers nor owns it, since its charter already sends it outside for audience evidence.

## See also

* [Role skills](role-briefings.md) - charter and progressive-disclosure model.
* [Boundary owners](ownership.md) - the two-sided relationship.
* [Role methods](role-briefings.md) - single-owner lazy procedures.
* [Personality libraries](personality.md) - the shared disposition axis.
* [Role-skill context budget](ai-engineer.md) - measured budget effects.
* [Evaluation](evaluation.md) - deterministic packs and review policy.
