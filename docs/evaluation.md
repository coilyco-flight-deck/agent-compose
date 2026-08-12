# Behavior evaluation

`agent-compose evaluation` emits deterministic, self-contained execution
packs:

```text
agent-compose evaluation --role engineer --seat codex
agent-compose evaluation --person-source ./person --role builder --seat codex
agent-compose evaluation --all --seat codex --format yaml --out-dir <empty-dir>
```

Markdown is the one-role default. YAML uses
`agent-compose.evaluation-pack.v2`. `--all` validates every Core Roster pack,
writes one file per role, and writes `index.json` with exact digests. Its output
directory must be empty.

Repository development uses `AGENT_COMPOSE_EVALUATION_OUT=<dir> ward exec evaluation-packs`; `<dir>` must be empty.

Agent Compose emits prompts and context, never model calls, credentials, scores,
or authority. `--person-source` selects one complete external package. The Core
[role boundaries](role-boundaries.md) decide which roles owe a boundary case
without replacing this pack or reviewer contract.

Execution packs are ephemeral. Their digest binds the exact role, personality,
[boundary](role-boundaries.md), policy, and case context used by one durable review
record. Editing a shared boundary retires every declaring role's results.

The [evaluation policy](evaluation-policy.md) defines model capability,
reasoning effort, session isolation, review, and evidence requirements.

## Core Roster matrix

Each role owns mission, personality, authority, completion, portfolio replay,
and adjacent-role scenarios. Each becomes exactly one case against the single
`commodity` subject tier.

The board does not read a role's own `model-tier` declaration. That declaration
is a deployment compatibility claim and roles still run on frontier and OSS
models, tested or not. Model tier does not change selected context, so one
subject measures the composed text for every role. See
[model-tiers.md](model-tiers.md).

Every Core Roster role also owns a human-communication scenario. Non-Creator
cases cover email, private messages, public and social posts, interviews,
meetings, and community conversations. They require the role to stop before
drafting or advising and give Content Creator only a factual handoff. Additional Ops,
Engineer, QA, and Director regressions require an authorized factual rollout
ledger, implementation checkpoint, verdict, or decision record without
deferring that mechanical artifact to Content Creator. Content Creator's
matching case requires complete recommendations without claiming send,
publication, account, moderation, or commitment authority.

Paired scenarios cover the declared Strategist, Director, Content Creator,
Designer, Engineer, Ops, and AI Engineer boundaries. QA has no other approved
adjacent pair.

The loader rejects incomplete kinds, a case off the subject tier, duplicates,
and incorrect adjacent-role targets. External packages retain the generic
fallback or may provide a complete custom matrix.

## Review records

[Review records](evaluation-records.md) own the compact v3 format, its
provenance requirements, release gating, and unreviewed driver output.

## See also

* [role-briefings.md](role-briefings.md) - role operating charters.
* [person-packages.md](person-packages.md) - independent evaluation context.
* [evaluation-matrices.md](evaluation-matrices.md) - profile-owned matrices.
* [evaluation-policy.md](evaluation-policy.md) - driver and reviewer contract.
* [evaluation-driver.md](evaluation-driver.md) - executable driver and reviewer.
* [evaluation-scorecard.md](evaluation-scorecard.md) - validated aggregate view.
