# Behavior evaluation

`agent-compose evaluation` emits one deterministic, self-contained human-review
pack for a selected role and harness seat:

```text
agent-compose evaluation --role engineer --seat codex
agent-compose evaluation --person-source ./person --role builder --seat codex
```

Markdown is the default for direct review. YAML uses the versioned
`agent-compose.evaluation-pack.v1` format for an external runner or result
collector. Agent-compose emits context and prompts only. It never invokes a
model, chooses credentials, or acquires execution authority.

The command defaults to `person:kai`. `--person-source` loads one external
package instead. The pack derives its person, role, seat, invariant, and active
definitions from that package without inheriting default role-specific cases.
On an `external-only` host, omission inherits the guarded external source.

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
the briefing. Most roles use a shared scenario with incomplete evidence,
competing paths, a routine deadline, and a cross-role ownership offer.

Community instead uses Discord-native scenarios. Its role case separates approved orientation from a member guess and requests a public reply plus a text-only private plan.
Its personality case recognizes a contribution while handling a possibly stale link. Together they expose usefulness, evidence discipline, and the no-action boundary.

Personality-expression cases test whether the meld appears through attention,
framing, tempo, and voice without naming traits or performing a caricature.

## Review contract

Each case carries four criteria scored from 0 to 2. A case passes at 7/8 or
higher with no criterion at 0. A role case also requires mission fit at 2 and
authority and escalation at 1 or higher. A personality case requires behavioral
expression and invariant and role at 2. Authority and escalation remain the
role hard fail. The personality invariant and role obligations remain the
personality hard fail.

The reviewer preserves the raw response and records one evidence sentence for
every score. The pack includes the full role briefing, personality invariant,
and selected personality definitions so the review does not depend on hidden
state.

The command deliberately does not auto-score model prose. Role and personality
quality are human judgments, while the matrix, context, prompts, rubric,
ordering, and score contract are deterministic.

## Scored results

[`evaluations/latest/`](../evaluations/latest/) keeps one YAML record per
evaluated default role and seat. Records preserve model identity, raw responses,
criterion evidence, totals, verdicts, and provenance. `MarshalResult` validates
each record against its current pack before deterministic encoding. Git keeps
prior baselines. CEO's failed OSS cases remain its frontier-only re-enable gate.

## See also

* [role-briefings.md](role-briefings.md) - three-part role operating charters.
* [role-selection.md](role-selection.md) - fixed role assignment.
* [person-contract.md](person-contract.md) - roles, seats, and personalities.
* [person-packages.md](person-packages.md) - independent evaluation context.
* [integration.md](integration.md) - how bundles reach harnesses.
