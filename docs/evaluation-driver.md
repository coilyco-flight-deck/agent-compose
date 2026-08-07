# Evaluation driver

`scripts/evaluation_driver.py` executes pack cases. `scripts/evaluation_reviewer.py`
scores the preserved responses. Both exist because pack rendering was
deterministic while execution stayed hand-driven, so no committed artifact
recorded how a result was earned.

## Driver

The driver launches one fresh harness session per case through the real
`acompose <role> <harness>` path, so the bundle under test is the bundle a
native launch would receive. Each case gets its own projection target, its own
process, and no session persistence. The driver preserves the raw response,
the resolved model, the finish reason, token usage, cost, and every retry.

One arm is one frozen driver configuration, spelled `name:model:effort`.
Repeating `--arm` runs the same case list under each configuration, which is
how a model comparison holds prompt, context, bundle, and runner constant.

```sh
python3 scripts/evaluation_driver.py \
  --packs <pack-dir> --out <run-dir> --sessions <scratch-dir> \
  --arm opus-high:claude-opus-5:high \
  --arm sonnet-medium:claude-sonnet-5:medium \
  --role ai --tier frontier --home <isolated-home>
```

## Context isolation

`--home` selects a stable isolated home. The home stays stable across cases
because the harness namespaces its stored credential by config directory, so a
per-case home would start logged out. Session freshness comes from the
per-case projection target and process, not from a new home.

`--allow-host-home` runs against the host home instead. The host global
instructions then sit in context beside the composed bundle, and a case can
pass on host instructions the bundle never supplied. Such a run stamps
`context-contaminated` in its record and cannot serve as bundle-behavior
evidence. It remains valid for comparing arms, because the same contamination
applies to every arm.

## Sealed sessions

A driver session inherits the harness's ordinary tool surface, which includes
every configured MCP server. A case that authorizes a factual record can then
write one to a real system: during the first full run, a director case posted a
decision record to a live tracking issue and reported the URL as its answer.

Sealed sessions are therefore the default. The driver launches each case with
no MCP servers and with shell, file-write, and notebook tools denied, so a case
can reason about an authorized action and state it without performing it.
Reading the repository through the harness's own read tools still works, which
is what the portfolio-replay cases depend on. Every record carries its
`tool_policy`.

`--allow-mutation` restores the full surface. A run made that way can act on
real systems and should be treated as an experiment, not as evidence.

## Reviewer

The reviewer scores in separate sessions resolved from the pack's reviewer
policy. It receives the role contract, the case prompt, the rubric, and one
preserved answer. It never learns which arm produced the answer and never sees
a competing answer to the same case. Verdicts are computed from the pack's
review rule rather than requested from the reviewer, so passing total, zero
rejection, and criterion minimums cannot drift with reviewer wording.

## See also

* [evaluation.md](evaluation.md) - pack generation and scored results.
* [evaluation-policy.md](evaluation-policy.md) - driver and reviewer contract.
* [model-tiers.md](model-tiers.md) - which lanes a role admits.
