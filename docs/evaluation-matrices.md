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
* `evidence-acquisition`

The engine expands each scenario into frontier, commodity, and OSS cases.
Every lane carries the same complete context. Keeping lanes and rubrics in the
engine prevents role assets from restating the same evaluation configuration.
`human-communication-ownership` is required from the roles declaring the
`comms` boundary and from its owner, which owes the case from the owning side
without ever receiving the body. A declaring role may carry more than one when
distinct cases are needed to preserve both a recommendation deferral and a
role-owned mechanical-record regression.
`evidence-acquisition` is required from exactly the roles that declare the
`evidence` boundary, and rejected from the roles that do not, so the roster decides
coverage instead of a second list here. The scenario adds an
`evidence-acquisition` criterion, scored but not a hard fail, because a
partially grounded claim is a quality deduction rather than an authority breach.
Case prompts are the only context a driver session receives, so these cases
score whether the response treats opening the authoritative source as required
work, not whether a file was read. Staging real artifacts for a driver to open
is a separate methodology change.
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
