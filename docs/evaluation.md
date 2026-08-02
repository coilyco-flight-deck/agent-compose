# Behavior evaluation

`agent-compose evaluation` emits deterministic, self-contained human-review
packs:

```text
agent-compose evaluation --role engineer --seat codex
agent-compose evaluation --person-source ./person --role builder --seat codex
agent-compose evaluation --all --seat codex --format yaml --out-dir <empty-dir>
```

Markdown is the default for one-role review. YAML uses
`agent-compose.evaluation-pack.v2`. `--all` validates every embedded Core
Roster pack, writes one file per role, and writes `index.json` with exact pack
digests. The output directory must be empty so stale role files cannot survive
a roster change.

Repository development uses `AGENT_COMPOSE_EVALUATION_OUT=<dir> ward exec evaluation-packs`; `<dir>` must be empty.

Agent Compose emits context and prompts only. It never invokes a model, chooses
credentials, scores prose, or acquires execution authority. The command
defaults to `roster:core`. `--person-source` loads one external package and
derives the whole review context from it.

## Core Roster matrix

Each role owns mission, personality, authority, completion, and real-portfolio
replay scenarios. Every scenario expands unchanged across a frontier lane with
a `frontier` bundle and an OSS lane with a `low-context` bundle.

The matrix also has explicit adjacent-role discrimination scenarios for
Portfolio Strategist and Director, Content Manager and Designer, Engineer and
DevOps, plus Content Manager and Community Manager. Both roles receive their
side of each boundary. Content Manager therefore has two adjacent-role
scenarios. QA has none because the approved v2 boundary list names no QA pair.

The loader rejects incomplete kinds or tiers, tier-dependent prompt drift,
duplicates, and incorrect adjacent-role targets. External packages retain the
generic fallback or may provide a complete custom matrix.

## Review contract

Each case carries four criteria scored from 0 to 2. A case passes at 7/8 or
higher with no criterion at 0. A role case also requires mission fit at 2 and
authority and escalation at 1 or higher. A personality case requires
behavioral expression and invariant and role at 2. Authority and escalation
remain the role hard fail. The personality invariant and role obligations
remain the personality hard fail.

The runner starts a fresh session per case, submits the prompt verbatim, and
records the model, raw response, and any terminal finish reason. Empty content
requires a non-success finish reason, zero scores, and a failing verdict.
Retries record case, attempt, outcome, and reason. Failures remain evidence.

An independent reviewer scores every criterion and writes one evidence sentence
per score. The author of a roster, prompt, or rubric change cannot be the sole
reviewer. The reviewer records the exact source revision and the digest from
`index.json`. `agent-compose scorecard` rejects incomplete lanes, unknown cases,
digest drift, missing retry provenance, inconsistent totals, and verdicts that
do not follow the rubric.

## Scored results

[`evaluations/latest/`](../evaluations/latest/) contains the independently
reviewed v2 baseline. Historical mode renders archived records without
rebinding their pack digests.

V2 records keep models, responses, scores, evidence, verdicts, pack digests,
source revisions, reviewers, finish reasons, and retries. Runner-local paths
become `<evaluation-worktree>` before commit. Every model completes its tier.
Frontier cases gate release. OSS failures remain visible and keep that role and
model class unsupported.

## See also

* [role-briefings.md](role-briefings.md) - role operating charters.
* [person-packages.md](person-packages.md) - independent evaluation context.
* [evaluation-matrices.md](evaluation-matrices.md) - profile-owned matrices.
* [evaluation-scorecard.md](evaluation-scorecard.md) - validated aggregate view.
