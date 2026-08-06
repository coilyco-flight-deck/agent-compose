# Roster package contract

The binary embeds one ordered `roster:core` source set as the default. A caller
may select one external package with the same contract:

```kdl
roster "core" {
    role "engineer" {
        purpose "Build and land work across Kai's real repository portfolio."
        model-tier "frontier" "commodity" "oss"
        skill "role-engineer"
        personality "curious" "grounded" "meticulous" "tenacious"
        agent "claude" name="opal engineer" pronouns="she" tier="frontier"
        agent "codex" name="terran engineer" pronouns="he" tier="frontier"
    }
    personality "curious" skill="personality-curious" color="#d98e48" motif="map-paper"
}
```

A package splits the manifest, roles, personalities, inspirations, invariant,
and definitions into the layout documented in
[person-packages.md](person-packages.md). The loader assembles and validates the
package before it becomes a source.

A role names its display name, purpose, role skill, nonempty ordered
personality meld, and seats. Its optional `model-tier` list restricts
composition, while omission supports all three tiers. Core roles declare the
list explicitly. The [role skill](role-briefings.md) needs valid frontmatter
and at least three body paragraphs. The loader rejects invalid tiers or
personality sets.
The default has eleven roles and sixteen personalities. Its explicit
[role-by-tier matrix](model-tiers.md) groups complex, foundational, and
high-security roles without changing their authority.

A personality entry is a catalog binding from its canonical name to a stable
skill id (`personality-<name>`). Every role reference needs one or the loader
rejects it. The same selected person source supplies one complete `SKILL.md`
tree for every binding plus the personality invariant. A missing, empty, extra,
or mismatched definition fails source validation. Roster output therefore
carries every selected definition without a capability provider.

The [identity primitives](identity-primitives.md) define renderer semantics. The [inspiration catalogue](inspiration-catalogue.md) defines credits.

## Favorite colors

Every selected personality declares one hex `color` in its person source, which
owns the exact palette. Bundles tell the agent every active personality's
name, skill, and color plus the melded favorite. The parse gate requires OKLab
lightness 0.60-0.80 and chroma of at least 0.05. Each role derives its favorite
as the OKLab centroid of every component, restores chroma to their minimum, and
clamps it into the legible band - the perceptual middle, never gray.

## Agent seats

An `agent` node is a named seat. The harness is its join key, while `name` and
`pronouns` are its identity. Optional `tier` must be canonical and supported by
the role.
Launch consumers keep permissions, models, and reasoning effort on their side.
Nothing here grants authority.

Every Core Roster role carries named harness seats. Seat keys remain stable
join points while display names and pronouns remain roster-owned identity.

Seats are personality-neutral. A compose request selects a role, and that role
activates its role skill and ordered personality set. Roster delivery is documented
in [role-briefings.md](role-briefings.md).

A private overlay may add scoped instructions or selection rules. It may not
redefine selected roles, personalities, seats, definitions, or role
personality sets. An external person package replaces the embedded default as
one unit. AOS owns no copy of either package or its personality definitions.

## See also

* [person-snapshot.md](person-snapshot.md) - complete machine-readable export.
* [model-tiers.md](model-tiers.md) - Core Roster compatibility matrix.
* [person-packages.md](person-packages.md) - external layout and selection.
* [role-briefings.md](role-briefings.md) - role charter and delivery contract.
* [kdl-contracts.md](kdl-contracts.md) - request and source grammar.
