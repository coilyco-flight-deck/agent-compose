# Profile-owned evaluation matrices

The engine owns the run protocol, scoring rule, frontier, commodity, and OSS
lanes, and generic rubrics. A Core Roster role owns
`evaluations/<role>.yaml` with ordered scenario inputs. Each scenario names a
stable id, one supported kind, its prompt, and an optional reviewer question.
Adjacent-role scenarios also name the role whose boundary they test.

Supported Core Roster scenario kinds are:

* `mission-fit`
* `personality-expression`
* `authority-boundary`
* `completion-ownership`
* `portfolio-replay`
* `adjacent-role-discrimination`
* `human-communication-ownership`
* `external-validation-deferral`

The engine expands each scenario into frontier, commodity, and OSS cases.
Every lane carries the same complete context. Keeping lanes and rubrics in the
engine prevents role assets from restating the same evaluation configuration.
A boundary-bound kind is required from every role that receives that boundary,
which is the roles declaring it plus its owner, and rejected from roles that do
not, so the roster decides coverage instead of a second list here.
`human-communication-ownership` belongs to `suggest-human-comms` and
`external-validation-deferral` to `seek-external-validation`. A role may carry
more than one when distinct cases are needed, as with a recommendation deferral
beside a role-owned mechanical-record regression. The
`external-validation-deferral` criterion is scored but not a hard fail, since a
partially grounded claim is a quality deduction rather than an authority breach.
Case prompts are the only context a driver session receives, so these cases
score stated behavior rather than observed tool use.
A pack-level
`disabled_model_tiers` marker temporarily pauses a lane without deleting its
cases or changing role-owned scenario files.

External roster packages may instead provide a complete custom matrix with
`run_protocol`, `review_rule`, and arbitrary `cases`. A complete matrix replaces
the generic matrix as one unit. A role without an asset receives the generic
fallback. Legacy prompt-only assets remain readable for external v1 packages.

## See also

* [evaluation.md](evaluation.md) - pack and scored-result contract.
* [creator-ownership.md](creator-ownership.md) - engine and profile boundaries.
