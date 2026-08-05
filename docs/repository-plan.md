# Repository plan

`agent-compose cascade` writes the compiled repository availability contract at
`~/.agent-compose/repository-plan.yaml`. The format marker is
`agent-compose.repositories.v2`.

The file is machine-owned, deterministic YAML. It remains readable so an
operator can inspect the projects root, trusted inputs, per-role selections,
and residency union without replaying `.agents/roles.kdl`.

```yaml
format: agent-compose.repositories.v2
projects_root: /absolute/projects/root

inputs:
  - identity: example/operating-context
    revision: 0123456789012345678901234567890123456789
    policy:
      path: .agents/roles.kdl
      sha256: sha256:0123456789012345678901234567890123456789012345678901234567890123

roles:
  example-role:
    - identity: example/context
      path: /absolute/projects/root/example/context
      source: example/operating-context
      scope: global
      reason: repository policy makes this repository available to every role

residency:
  - identity: example/context
    path: /absolute/projects/root/example/context
    source: example/operating-context
    scope: role-union
    reason: repository is selected by at least one canonical role
```

The document contains:

* `inputs` - trusted policy source identity, full Git revision, policy path,
  and SHA-256 digest.
* `roles` - the exact sorted selection for every canonical role.
* `residency` - every role-selected repository plus resident-only pins.
* selection provenance - source, scope, reason, and provider selector details.

Consumers validate the same bounded YAML subset Agent Compose writes. The
loader rejects aliases, custom tags, duplicate keys, unknown fields, ambiguous
scalar coercion, unsafe paths, unsorted identities, and incomplete provenance.

Consumers never parse `roles.kdl`. Native skill linking reads `residency`.
Native role launch reads only the selected role.

## See also

* [Repository policy](repository-policy.md) - KDL policy grammar.
* [Cascade](cascade.md) - host convergence and plan emission.
* [Native role launch](native-role-launch.md) - role selection and launch.
