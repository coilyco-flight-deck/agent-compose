# KDL contracts

Agent-compose uses KDL for human-authored requests and policy. Parsers reject
unknown nodes, duplicate facts, missing sources, or an empty selection.

## Compose request

A request names a role, a delivery mode, an optional person package, and any
external capability sources:

```kdl
compose {
    person-policy "external-only"
    person-source "person"
    role "engineer"
    model-class "frontier"
    delivery "native-skills"
    source "aos-public" root="agentic-os" required=#true
}
```

`person-source` names a request-relative package and fully replaces the
embedded package. `person-policy "external-only"` requires it and prohibits
fallback. Omitting both selects `roster:core`, unless the host guard supplies it.

The role activates its selected personality set, ordinary provider skills, and
its composed-skill allowlist. `delivery` is `native-skills` or `compiled`.
`model-class` defaults to `frontier`. `low-context` excludes only skills whose top-level
frontmatter says `low-context: optional`. Embedded role policy may reject a class.

Legacy `density "full"` is ignored and other densities fail. Sources run in
request order. `root` and `declaration` only locate files.

## Capability sources

The public AOS provider needs only its root. Agent-compose discovers every
ordinary skill under `.agents/skills` in lexical order. It also reads
`.agents/roles.kdl`:

```kdl
roles {
    role "engineer" {
        composed-skill "coding-*"
    }
}
```

Each `composed-skill` admits `.agents/composed/<name>/COMPOSED.md` by exact name or a `coding-*` glob. Globs expand lexically.
Invalid, unmatched, and overlapping selections fail closed. Each `intent` records one model-opaque default harness route.
Materialization renames the admitted entry point to `SKILL.md`. Nested
`SKILL.md` files and ordinary/composed name collisions fail.
The same root form works in requests, roster arguments, and `roster_sources`.
Roster sources are optional overlays. The selected person source always
supplies the invariant and bound personality bodies.

An overlay or another provider can instead carry an explicit declaration:

```kdl
source "aos-public" {
    instruction "foundation" path="content/foundation.md"
    skill "coding-go" path="skills/coding-go"
}
```

The request admits that file with
`source "aos-public" declaration="source-public.kdl"`. Paths inside the
declaration are relative to the declaration and must stay beneath its source
root. Request locator paths are relative to the request. Symlinks and escaping
paths fail validation. A required missing source fails composition. An optional
missing source is skipped with a note in the trace.

For rolling upgrades, inferred providers may still contain the former
`personality-shared/INVARIANT.md` and `personality-*` trees. Identical copies
shadow behind the selected person source. A different copy conflicts and stops
composition.

## See also

* [person-contract.md](person-contract.md) - person-package policy.
* [person-packages.md](person-packages.md) - external package selection.
