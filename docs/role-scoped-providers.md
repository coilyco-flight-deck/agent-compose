# Role-scoped skill providers

A launch consumer can admit local skill providers for one selected role without
adding those providers to every role or to the host-global skill surface. The
consumer owns which roles receive which repositories. Agent Compose treats role
slugs and provider paths as opaque configuration.

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

Native launch remaps the same ordered repository set into an isolated workspace
when a consumer reproduces `projects_root` there. Agent Compose composes one
immutable role bundle from that set. Native projection and `project --scope
home` consume that same bundle, so a staged home receives the same selected
skill inventory without changing bundle ownership or launch authority.

`agent-compose describe <bundle>` shows selected and excluded providers.
`agent-compose describe <bundle> --why source:<provider-id>` explains provider
admission, and `--why skill:<skill-name>` follows a provider skill to its
selected, excluded, or shadowed outcome.

## Ownership

AOS or another launch consumer hydrates and verifies local catalogue roots.
Agent Compose performs offline selection, composition, collision checking, and
read-only bundle projection. Agent Compose does not fetch provider content,
choose operational authority, or add repository-specific mappings to the Core
Roster.

## See also

* [Native role launch](native-role-launch.md) - assigned native delivery.
* [Staged-home handoff](staged-home.md) - isolated home projection.
* [Cascade](cascade.md) - generated mount eligibility and bare convergence.
* [Integration](integration.md) - host and isolated ownership tiers.
