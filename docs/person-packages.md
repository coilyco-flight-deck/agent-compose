# External person packages

One agent-compose installation can use a person package that is completely
independent of the shipped `person:kai` default. Selection is exclusive.
Agent-compose never merges roles, seats, personalities, definitions, or
evaluation context across the two packages.

The package owns identity and operating policy. Capability providers still own
general skills and reusable doctrine. Launchers still own models, permissions,
credentials, tools, and execution authority.

Authors follow [person-package-authoring.md](person-package-authoring.md).

## Bundle composition

A portable request selects a package relative to the request file:

```kdl
compose {
    person-policy "external-only"
    person-source "person"
    role "builder"
    model-class "frontier"
    delivery "native-skills"
    source "knowledge" root="knowledge"
}
```

Omitting both person nodes selects the embedded `person:kai` default.
`person-policy "external-only"` requires its paired source. The bundle manifest
records `person:<name>`, never the local package path.

## Host convergence

The host config selects a package with `person_source`:

```yaml
person_policy: external-only
person_source: /path/to/person
sources:
  - /path/to/AGENTS.md
roots:
  - ~/.config/agent-compose/sources
load_points:
  claude: ~/.claude/CLAUDE.md
  codex: ~/.codex/AGENTS.md
```

The path may be absolute, config-relative, or home-relative. `external-only`
makes the selection machine-wide. Requests and direct person commands inherit
the source when they omit one. A request may name another external package.
Missing or invalid sources abort before projection. Refresh-then-exec also
refuses last-known-good fallback because that projection may use the default.
The direct `project` command rejects embedded-person bundles under the guard.

Without `person_policy`, removing `person_source` returns bare convergence to
the embedded default. Existing installations retain that behavior.

The custom package and its machine rollout belong in their own repository or
host configuration. They do not belong in the public agent-compose engine.

## Direct identity and evaluation commands

Person-dependent commands accept the same package explicitly:

```text
agent-compose evaluation \
  --person-source /path/to/person \
  --role builder --seat codex
```

`overlay`, `roster`, and `palette-data` accept the same flag. Under the host
guard they inherit its source when the flag is absent.

Evaluation packs include the selected person name, role briefing, seat,
invariant, and active definitions. External packages use the generic
four-case frontier and OSS matrix, so they do not inherit a role-specific case
from `person:kai`. Agent-compose emits the deterministic pack and validates
scored results. An external runner or human still owns model invocation,
credentials, response capture, and scoring.
