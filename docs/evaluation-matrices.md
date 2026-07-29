# Profile-owned evaluation matrices

The engine ships one complete generic evaluation fallback. A profile may add
`evaluations/<role>.yaml` with its own complete `run_protocol`, `review_rule`,
and arbitrary named `cases`.

Each custom case names its stable seat-oriented lane, bundle model class,
dimension, prompt, reviewer question, and rubric. A complete custom matrix
replaces the generic matrix as one unit. A role without one receives the
generic fallback.

Legacy prompt-only profile assets remain readable during v1.x. They retain the
generic protocol, review rule, lanes, and rubrics while replacing only the
role-specific prompts.

## See also

* [evaluation.md](evaluation.md) - pack and scored-result contract.
* [content-ownership.md](content-ownership.md) - engine and profile boundaries.
