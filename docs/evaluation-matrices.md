# Profile-owned evaluation matrices

The engine owns the run protocol, scoring rule, frontier and OSS lanes, and
generic rubrics. A Core Roster role owns `evaluations/<role>.yaml` with ordered
scenario inputs. Each scenario names a stable id, one supported kind, its
prompt, and an optional reviewer question. Adjacent-role scenarios also name
the role whose boundary they test.

Supported Core Roster scenario kinds are:

* `mission-fit`
* `personality-expression`
* `authority-boundary`
* `completion-ownership`
* `portfolio-replay`
* `adjacent-role-discrimination`
* `human-communication-ownership`

The engine expands each scenario into paired frontier and OSS cases. Keeping
lanes and rubrics in the engine prevents eight role assets from restating the
same evaluation configuration. A pack-level `disabled_model_tiers` marker
temporarily pauses a lane without deleting its cases or changing role-owned
scenario files.

External roster packages may instead provide a complete custom matrix with
`run_protocol`, `review_rule`, and arbitrary `cases`. A complete matrix replaces
the generic matrix as one unit. A role without an asset receives the generic
fallback. Legacy prompt-only assets remain readable for external v1 packages.

## See also

* [evaluation.md](evaluation.md) - pack and scored-result contract.
* [content-ownership.md](content-ownership.md) - engine and profile boundaries.
