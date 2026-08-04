# Evaluation scorecard

`agent-compose scorecard` validates scored result records against their current
evaluation packs, then renders one dense Markdown page:

```text
ward exec evaluation-scorecard
agent-compose scorecard --results evaluations/latest --seat codex
```

Each role occupies one row. `FR`, `FP`, `OR`, and `OP` are the frontier role,
frontier personality, OSS role, and OSS personality cases. V2 scorecards add
`FA` and `OA` for adjacent-role discrimination when any result contains those
cases. A cell aggregates its cases and carries points plus `✓` or `×`. A missing
adjacent lane renders `-`, which is expected for QA. The header carries
aggregate case and point totals.

Human-communication ownership cases remain in the role columns. Their explicit
fifth hard-fail criterion makes those role-cell maximums larger than standard
8-point cases, so affected cells render `points/maximum`.

The Ward verb refreshes the committed
[`evaluation-scores.md`](evaluation-scores.md). Direct CLI use emits
the page to standard output unless `--out` names a file. `--check` fails when
that file differs from a fresh render.

Uniform frontier and OSS model names appear once in the header. Mixed fields
name both models on each affected row. The YAML records remain the owning score
source, while the committed page is a renderer-verified view.

## See also

* [evaluation.md](evaluation.md) - evaluation packs and scored-result contract.
* [evaluation-matrices.md](evaluation-matrices.md) - profile-owned matrices.
* [FEATURES.md](FEATURES.md) - shipped capability inventory.
