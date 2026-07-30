# Evaluation scorecard

`agent-compose scorecard` validates scored result records against their current
evaluation packs, then emits one dense Markdown page to standard output:

```text
ward exec evaluation-scorecard
agent-compose scorecard --results evaluations/latest --seat codex
```

Each role occupies one row. `FR`, `FP`, `OR`, and `OP` are the frontier role,
frontier personality, OSS role, and OSS personality cases. A cell carries its
points plus `✓` or `×`. The header carries aggregate case and point totals.

Uniform frontier and OSS model names appear once in the header. Mixed fields
name both models on each affected row.

The rendered page stays uncommitted. The YAML records remain the only owning
source, and rerunning the command incorporates every current selected result.

## See also

* [evaluation.md](evaluation.md) - evaluation packs and scored-result contract.
* [evaluation-matrices.md](evaluation-matrices.md) - profile-owned matrices.
* [FEATURES.md](FEATURES.md) - shipped capability inventory.
