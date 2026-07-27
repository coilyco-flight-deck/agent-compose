# External person packages

One agent-compose installation can use a person package that is completely
independent of the shipped `person:kai` default. Selection is exclusive.
Agent-compose never merges roles, seats, personalities, definitions, or
evaluation context across the two packages.

The package owns identity and operating policy. Capability providers still own
general skills and reusable doctrine. Launchers still own models, permissions,
credentials, tools, and execution authority.

Package authors use the validated layout and naming rules in
[person-package-authoring.md](person-package-authoring.md).

## Bundle composition

A portable request selects a package relative to the request file:

```kdl
compose {
    person-source "person"
    role "builder"
    model-class "frontier"
    delivery "native-skills"
    source "knowledge" root="knowledge"
}
```

Omitting `person-source` selects the embedded `person:kai` default. The bundle
manifest records the selected package as `person:<name>`. It never records the
local package path.

## Host convergence

The host config selects a package with `person_source`:

```yaml
person_source: /path/to/person
sources:
  - /path/to/AGENTS.md
roots:
  - ~/.config/agent-compose/sources
load_points:
  claude: ~/.claude/CLAUDE.md
  codex: ~/.codex/AGENTS.md
```

The path may be absolute, relative to the config file, or home-relative. Bare
`acompose` validates the package, atomically replaces the generated roster and
snapshot, then runs normal cascade convergence. Removing `person_source`
returns that host to the embedded default on its next convergence.

The custom package and its machine rollout belong in their own repository or
host configuration. They do not belong in the public agent-compose engine.

## Direct identity and evaluation commands

Person-dependent commands accept the same package explicitly:

```text
agent-compose evaluation \
  --person-source /path/to/person \
  --role builder --seat codex
```

`overlay`, `roster`, and `palette-data` accept the same `--person-source` flag.

Evaluation packs include the selected person name, role briefing, seat,
invariant, and active definitions. External packages use the generic
four-case frontier and OSS matrix, so they do not inherit a role-specific case
from `person:kai`. Agent-compose emits the deterministic pack and validates
scored results. An external runner or human still owns model invocation,
credentials, response capture, and scoring.

## See also

* [person-contract.md](person-contract.md) - complete package schema.
* [person-package-authoring.md](person-package-authoring.md) - filesystem layout.
* [kdl-contracts.md](kdl-contracts.md) - compose request grammar.
* [evaluation.md](evaluation.md) - behavior review and scoring.
