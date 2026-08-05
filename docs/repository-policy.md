# Repository policy

Agent Compose is the sole compiler for role-aware repository availability.
Trusted operating-context roots declare policy in `.agents/roles.kdl`. Host
configuration names those roots through `operating_context`. Doctrine sources,
harness load points, and repositories merely present on disk never grant
availability.

## Grammar

The optional top-level `repositories` node declares stable repository ids and
their `owner/repository` paths:

```kdl
repositories {
    repository lore path="coilysiren/lore"
    repository voice-corpus path="coilysiren/voice-corpus"
    repository profile path="coilysiren/coilysiren"

    global lore
    resident-only profile
}

roles {
    role content {
        use-repository voice-corpus
    }
}
```

Policy has three distinct scopes:

* `global` makes a repository available to every canonical role.
* `use-repository` makes a repository available only to the containing role.
* `resident-only` keeps a checkout on the host without granting it to any role.

Role providers remain repository selections with skill-selector provenance.
Agent Compose rejects unknown ids, duplicate paths, duplicate use, unsafe
paths, conflicting trusted definitions, and provider cycles.

## Compiled plan

`agent-compose cascade` writes
`~/.agent-compose/repository-plan.yaml` in
`agent-compose.repositories.v2` format. The file is machine-owned but
intentionally reviewable. It records each trusted policy source identity, full
Git revision, `.agents/roles.kdl` path, and SHA-256 digest before the role and
residency selections.

The plan uses a bounded safe-YAML subset and absolute paths beneath one
`projects_root`. Consumers reject unsafe YAML, unsafe paths, unsorted
identities, and incomplete provenance. Consumers never parse `roles.kdl`
themselves. Native skill linking uses the residency projection. A role launch
reads only its role selection. See [Repository plan](repository-plan.md).

## Verified bundle handoff

Selected repository identities and provenance are sealed into `manifest.json`
and the decision trace. `agent-compose verify` checks ordering, identity safety,
provenance, and manifest-to-trace agreement.

`agent-compose bundle materialize --role <role> --harness <harness>` converges
the host and returns a verified immutable bundle selection without starting a
harness. This is the adapter surface for AOS and other launch consumers.

Repository metadata expresses context reachability only. It grants no runtime
authority, network access, credentials, or mutation permission.

## See also

* [Cascade](cascade.md) - host convergence and plan emission.
* [Repository plan](repository-plan.md) - YAML v2 contract and provenance.
* [Native role launch](native-role-launch.md) - role selection and launch.
* [Bundle protocol](bundle-protocol.md) - immutable consumer contract.
* [Role-scoped providers](role-scoped-providers.md) - provider selection.
