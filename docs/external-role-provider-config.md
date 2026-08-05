# Legacy external role-provider configuration

This surface exists only for the rolling migration to unified
`.agents/roles.kdl` provider graphs. New declarations belong in the trusted
knowledge root's KDL document.

A deployment can keep role-provider selection in a separately owned canonical
file without copying it into a rendered host wrapper.

## Host wrapper

The main host configuration references one external document:

```yaml
# agent-compose.yaml
role_providers_file: role-providers.yaml
```

Deployment tooling may make that conventional path a symlink to the canonical
source. The target remains product-native configuration:

```yaml
# role-providers.yaml
role_providers:
  operations:
    - path: example/infrastructure
      required: true
```

## Resolution

The external path resolves relative to `agent-compose.yaml`. `~/` expands
against the current home. Agent Compose follows symlinks before reading the
document.

A relative provider `path` identifies a repository beneath `projects_root`.
An absolute provider path must also remain beneath that root. Logical paths
stay unchanged in both YAML files. Manifest rendering resolves and
canonicalizes them beneath `projects_root`, so runtime state receives absolute
paths without rewriting source configuration.

## Validation

The external document accepts only the `role_providers` key. Inline
`role_providers` and `role_providers_file` are mutually exclusive. The same
strict provider, selector, unknown-field, and trailing-document validation
applies before the mount manifest or a bundle can be written.

If an admitted trusted root declares any KDL providers, Agent Compose rejects
either legacy YAML form. This prevents two authored graphs from silently
merging during rollout.

## Ownership

Deployment tooling owns the symlink and host wrapper. The referenced provider
owns the canonical logical mapping. Agent Compose owns strict loading, path
hydration, selection, and immutable projection. It never fetches provider
content or rewrites either configuration file.

## See also

* [Role-scoped providers](role-scoped-providers.md) - selection and delivery.
* [Cascade](cascade.md) - host convergence and mount eligibility.
