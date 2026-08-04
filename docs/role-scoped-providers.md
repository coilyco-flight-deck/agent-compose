# Role-scoped skill providers

A launch consumer can admit local skill providers for one selected role
without adding them to every role or the host-global skill surface. The
consumer owns the mapping. Role slugs and paths are opaque configuration.

## Configuration

`role_providers` belongs in `agent-compose.yaml` beside the existing source and
load-point settings:

```yaml
role_providers:
  operations:
    - path: example/infrastructure
      required: true
    - path: example/deploy
      required: false
```

A relative `path` identifies a repository beneath `projects_root`. An absolute
path must also remain beneath that root. Configuration and generated manifests
reject unknown fields.

Optional `skills` entries select a bounded ordinary-skill slice. See
[Ordinary-skill selectors](ordinary-skill-selectors.md). Omission preserves
the whole provider.

Required providers fail the assigned launch when their `.agents/skills`
catalogue is unavailable. Optional providers leave an explicit excluded-source
decision in `trace.json` and composition continues. A provider assigned to a
different role is excluded with its provider and discoverable skill reasons in
the same trace.

## Selection

An assigned launch resolves providers in this order:

1. Default repositories.
2. Repositories admitted for the selected harness.
3. Providers admitted for the selected role.

Duplicate repository paths collapse at their first position. Within the
resolver, byte-identical duplicate skills may shadow the earlier copy. A
different copy with the same skill name fails closed.

An empty role selects only defaults and harness repositories. Bare `acompose`
convergence therefore never links role-only providers into host-global skill
load points.

## Native and staged delivery

Native launch remaps the ordered repositories when a consumer reproduces
`projects_root` in an isolated workspace. Native projection and `project
--scope home` consume one immutable bundle, so staged homes receive the same
skills without changing ownership or authority.

`agent-compose describe <bundle>` shows selected and excluded providers in a
dedicated provider section. It classifies default and harness roots as ordinary
catalogues, role roots as role providers, and the selected roster as a person
package. Its context-budget section names each provider's selected skill count,
retained bytes, and approximate tokens. Excluded providers contribute explicit
zeroes. A selected slice records its selector outcome and bounded budget.
`agent-compose describe <bundle> --why source:<provider-id>` explains provider
admission, and `--why skill:<skill-name>` follows a provider skill to its
selected, selector-excluded, role-excluded, or shadowed outcome.

## Ownership

AOS or another launch consumer hydrates and verifies local catalogue roots.
Agent Compose performs offline selection, composition, collision checking, and
read-only bundle projection. Agent Compose does not fetch provider content,
choose operational authority, or add repository-specific mappings to the Core
Roster.

## See also

* [Cascade](cascade.md) - generated mount eligibility and bare convergence.
* [Ordinary-skill selectors](ordinary-skill-selectors.md) - bounded catalogue slices.
* [Integration](integration.md) - host and isolated ownership tiers.
