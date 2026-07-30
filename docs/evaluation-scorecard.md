# Evaluation scorecard

`agent-compose scorecard` validates scored result records against their current
evaluation packs, then renders one dense Markdown page:

```text
ward exec evaluation-scorecard
agent-compose scorecard --results evaluations/latest --seat codex
```

Each role occupies one row. `FR`, `FP`, `OR`, and `OP` are the frontier role,
frontier personality, OSS role, and OSS personality cases. A cell carries its
points plus `✓` or `×`. The header carries aggregate case and point totals.

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
