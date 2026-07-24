# Behavior evaluation

`agent-compose evaluation` emits one deterministic, self-contained human-review
pack for a canonical role and harness seat:

```text
agent-compose evaluation --role engineer --seat codex
agent-compose evaluation --role ops --seat codex --format json
```

Markdown is the default for direct review. JSON uses the versioned
`agent-compose.evaluation-pack.v1` format for an external runner or result
collector. Agent-compose emits context and prompts only. It never invokes a
model, chooses credentials, or acquires execution authority.

## Four-case matrix

The pack fixes two dimensions across two model tiers:

* frontier role understanding
* frontier personality expression
* OSS role understanding
* OSS personality expression

The frontier lane uses a `frontier` bundle. The OSS lane uses a `low-context`
bundle so the comparison exercises the pruning contract smaller models receive.
Both lanes keep the selected role and seat fixed.

Role-understanding cases test whether the response applies the role's mission,
operating method, completion ownership, and authority boundary without quoting
the briefing. Personality-expression cases test whether the meld appears
through attention, framing, tempo, and voice without naming traits or
performing a caricature.

## Review contract

Each case carries four criteria scored from 0 to 2. A case passes at 6/8 or
higher unless a hard-fail criterion scores 0. Authority and escalation are the
role hard fail. The personality invariant and role obligations are the
personality hard fail.

The reviewer preserves the raw response and records one evidence sentence for
every score. The pack includes the full role briefing, personality invariant,
and selected personality definitions so the review does not depend on hidden
state.

The command deliberately does not auto-score model prose. Role and personality
quality are human judgments, while the matrix, context, prompts, rubric,
ordering, and score contract are deterministic.

## See also

* [role-briefings.md](role-briefings.md) - three-part role operating charters.
* [role-selection.md](role-selection.md) - fixed role assignment.
* [person-contract.md](person-contract.md) - roles, seats, and personalities.
* [integration.md](integration.md) - how bundles reach harnesses.
