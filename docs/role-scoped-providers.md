# Role-scoped skill providers

A trusted root can admit local skill providers for one role without widening
the host-global surface. It owns `.agents/roles.kdl`, whose provider IDs are
document-local.

## Configuration

Declare providers beside composed skills in the trusted root's role graph:

```kdl
providers {
    provider infrastructure path="example/infrastructure"
    provider deploy path="example/deploy"
}
roles {
    role operations {
        use-provider infrastructure required=#true
        use-provider deploy required=#false
    }
}
```

Optional `skill` children select a bounded ordinary-skill slice. See
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

Duplicate paths collapse first-wins. Byte-identical skills may shadow the
earlier copy, while different bodies with the same name fail closed.

An empty role selects only defaults and harness repositories. Bare `acompose`
convergence therefore never links role-only providers into host-global skill
load points.

## Native and staged delivery

Native launch remaps ordered repositories into an isolated `projects_root`.
Native projection and `project --scope home` consume the same immutable bundle.

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

A trusted root authors logical catalogue paths. Agent Compose discovers it
through the source graph, then owns strict loading, hydration, selection,
collision checking, and read-only projection. It does not fetch content,
choose authority, or add mappings to the Core Roster.

## See also

* [Cascade](cascade.md) - generated mount eligibility and bare convergence.
* [Ordinary-skill selectors](ordinary-skill-selectors.md) - bounded catalogue slices.
* [Integration](integration.md) - host and isolated ownership tiers.
