# Role methods

A person package may bind progressive-disclosure method skills to one role.
Use a role method when the procedure is determined by that package's roster
policy and would become misleading if copied into a general knowledge provider.

## Package layout

Declare method ids in the owning role fragment:

```kdl
role "ai" {
    skill "role-ai"
    method "eval-role-comms"
    method "eval-role-live-ops"
}
```

Store each body under that role:

```text
roles/ai/SKILL.md
roles/ai/skills/eval-role-comms/SKILL.md
roles/ai/skills/eval-role-live-ops/SKILL.md
```

Every method directory contains only `SKILL.md`. The frontmatter name matches
the declared id. Missing, extra, malformed, duplicate, or cross-role duplicate
methods fail package loading.

## Selection and delivery

Agent Compose selects methods only with their owning role. The selected role's
identity card names them without eagerly loading their bodies. Native delivery
projects them as discoverable skills, while compiled delivery appends the same
selected bodies to preserve behavior for harnesses without a skill loader.

Bare roster convergence installs every method beside role and personality
skills so an inferred interactive role can activate it. The roster instruction
keeps methods inactive until both the current role and task match. Assigned AOS
launches consume the verified bundle and own no copy of the method source.

## Core AI evaluation methods

The Core Roster AI Engineer owns two cross-role evaluation methods:

* `eval-role-comms` derives communication-owner, factual-handoff, advice, and
  delivery-gate coverage from the selected package.
* `eval-role-live-ops` derives observation, mutation, promotion, after-state,
  rollback, recovery, and operator-handoff coverage from canonical policy.

Both methods preserve pack provenance, isolated runs, raw failures, independent
review, and QA acceptance. Live-operations cases use offline fixtures. Neither
method grants communication, deployment, credential, or runtime authority.

## See also

* [Role skills](role-briefings.md) - charter and progressive-disclosure model.
* [Person package authoring](person-package-authoring.md) - complete layout.
* [Evaluation](evaluation.md) - deterministic packs and review policy.
* [Architecture](architecture.md) - provider and consumer boundaries.
