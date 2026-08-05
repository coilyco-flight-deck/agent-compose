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

1. Explicit operating-context repositories.
2. Global repositories.
3. Direct repositories and providers admitted for the selected role.

Duplicate paths collapse first-wins. Byte-identical skills may shadow the
earlier copy, while different bodies with the same name fail closed.

Composed-skill selectors within one provider role form a set union. When two
selectors match the same composed skill, Agent Compose selects that skill once,
emits a warning, and retains every matching selector in `trace.json`. Overlap
within one role does not create a content collision. Different skill bodies
with the same name still fail closed during cross-provider resolution.
An empty role selects only operating context and global repositories. Bare `acompose`
uses the host-residency union. Assigned launch consumers hide that global skill
mount before projecting the role bundle.

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

* [Repository policy](repository-policy.md) - direct role and residency selection.
* [Ordinary-skill selectors](ordinary-skill-selectors.md) - bounded catalogue slices.
* [Integration](integration.md) - host and isolated ownership tiers.
