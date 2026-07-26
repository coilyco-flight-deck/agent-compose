# Behavior evaluation

`agent-compose evaluation` emits one deterministic, self-contained human-review
pack for a canonical role and harness seat:

```text
agent-compose evaluation --role engineer --seat codex
agent-compose evaluation --role ops --seat codex --format yaml
```

Markdown is the default for direct review. YAML uses the versioned
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

[`evaluations/latest/`](../evaluations/latest/) keeps one versioned YAML record
per evaluated role and seat. Each record preserves model identity, raw
responses, criterion scores and evidence, totals, verdicts, and issue
provenance. Repository validation derives the expected cases, criteria, totals,
and pass decisions from the current generated pack.

A new accepted evaluation replaces that role and seat's latest file. Git
history preserves prior baselines, while issue comments retain the review
discussion that produced them. `MarshalResult` validates a scored record against
its current generated pack before encoding deterministic YAML. CEO keeps its
failed OSS cases as the evidence and re-enable gate for its frontier-only rule.

Current accepted Codex coverage includes Engineer, Director, QA, Advisor, Ops,
PM, Designer, Social, Sales, Customer Success, and CEO. Community has no accepted scored baseline yet, so its absence makes no pass or fail claim.

## See also

* [role-briefings.md](role-briefings.md) - three-part role operating charters.
* [role-selection.md](role-selection.md) - fixed role assignment.
* [person-contract.md](person-contract.md) - roles, seats, and personalities.
* [integration.md](integration.md) - how bundles reach harnesses.
