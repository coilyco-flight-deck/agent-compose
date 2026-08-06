# Roster package contract

The binary embeds one ordered `roster:core` source set as the default. A caller
may select one external package with the same contract:

```kdl
roster "core" {
    role "engineer" {
        purpose "Build and land work across Kai's real repository portfolio."
        model-tier "frontier" "commodity" "oss"
        skill "role-engineer"
        method "eval-role-comms"
        personality "curious" "grounded" "meticulous" "tenacious"
        identity name="opal engineer" pronouns="she"
        agent "claude" tier="frontier"
        agent "codex" tier="frontier"
    }
    personality "curious" skill="personality-curious" color="#d98e48" motif="map-paper"
}
```

A package splits the manifest, roles, personalities, invariant, and definitions
into the [external layout](person-packages.md). The loader validates it before
it becomes a source.

A role names its display name, purpose, role skill, [methods](role-methods.md),
ordered personality meld, and seats. Its optional `model-tier` list restricts
composition, while omission supports all three tiers. Core roles declare the
list explicitly. Core uses exactly three per role, covers every canonical
personality, caps usage at three roles, and requires legible, distinct derived
colors. External packages retain any nonempty ordered meld. The [role skill](role-briefings.md)
needs valid frontmatter and at least three body paragraphs. The loader rejects
invalid tiers or personality sets.
The default has eight roles and sixteen personalities. Its explicit
[role-by-tier matrix](model-tiers.md) groups complex, foundational, and
high-security roles without changing their authority.

A personality entry is a catalog binding from its canonical name to a stable
skill id (`personality-<name>`). Every role reference needs one or the loader
rejects it. The same selected person source supplies one complete `SKILL.md`
tree for every binding plus the personality invariant. A missing, empty, extra,
or mismatched definition fails source validation. Roster output therefore
carries every selected definition without a capability provider.

The [identity primitives](identity-primitives.md) define renderer semantics.

## Favorite colors

Every selected personality declares one hex `color` in its person source, which
owns the exact palette. Bundles tell the agent every active personality's
name, skill, and color plus the melded favorite. The parse gate requires OKLab
lightness 0.60-0.80 and chroma of at least 0.05. Each role derives its favorite
as the OKLab centroid of every component, restores chroma to their minimum, and
clamps it into the legible band - the perceptual middle, never gray.

## Agent seats

A Core role declares one `identity` with a name and pronoun pair. Every
`agent` node is a harness routing selector for that identity. Optional
`channel` and `tier` properties describe routing, and a tier must be
canonical and supported by the role.
Launch consumers keep permissions, models, and reasoning effort on their side.
Nothing here grants authority.

Every Core Roster role carries harness seats. Seat keys remain stable join
points while the role-owned name and pronouns remain identical across them.
Selecting another seat changes routing metadata only.

External packages authored before role-level identity may keep `name` and
`pronouns` on every seat. A role must use one form consistently. Mixing
role-level identity with per-seat identity fails validation.

Seats are personality-neutral. A compose request selects a role, which
activates its role skill, methods, and ordered personality set. See
[role-skill delivery](role-briefings.md).

A private overlay may add scoped instructions or selection rules. It may not
redefine selected roles, personalities, seats, definitions, or role
personality sets. An external person package replaces the embedded default as
one unit. AOS owns no copy of either package or its personality definitions.
