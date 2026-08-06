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
    model-tier "commodity"
    delivery "native-skills"
    source "aos-public" root="agentic-os" required=#true
}
```

`person-source` names a request-relative package and fully replaces the
embedded package. `person-policy "external-only"` requires it and prohibits
fallback. Omitting both selects `roster:core`, unless the host guard supplies it.

The role activates its personality set, ordinary skills, and composed-skill allowlist. `delivery` is `native-skills` or `compiled`.
`model-tier` is `frontier`, `commodity`, or `oss`, defaults to `frontier`, and must be supported by the role.
Model tier never changes selected context. Every supported tier receives the
complete role, personality, ordinary-skill, and composed-skill selection.

Legacy `density "full"` is ignored and other densities fail. Sources run in
request order. `root` and `declaration` only locate files.

## Capability sources

The public AOS provider needs only its root. Agent-compose discovers ordinary skills and reads one `.agents/roles.kdl` graph:

```kdl
repositories {
    repository hardware path="example/hardware-knowledge" {
        skill "machine-*"
    }
}
roles {
    role "engineer" {
        use-repository hardware
        composed-skill "coding-*"
    }
}
```

Repository IDs are document-local, paths use `owner/repository`, and `skill` marks a selected repository as an ordinary-skill provider with a bounded catalogue.
Selected skill-provider repositories fail closed when their checkout or `.agents/skills` catalogue is unavailable. Only trusted roots widen eligibility, and imported graphs do not recurse.
See [role-scoped providers](role-scoped-providers.md) for resolution and provenance.

Each `composed-skill` admits `.agents/composed/<name>/COMPOSED.md` by exact name or glob. Globs expand lexically. Invalid or overlapping selections fail.
Materialization renames admitted entry points to `SKILL.md`. Nested `SKILL.md` files and ordinary/composed name collisions fail.
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

The request admits it with `source "aos-public" declaration="source-public.kdl"`. Paths stay beneath the declaration root.
Symlinks and escaping paths fail. Required missing sources fail, while optional
ones produce trace decisions.

During rolling upgrades, identical legacy invariant and personality copies
shadow behind the person source. Different copies conflict.

## See also

* [person-packages.md](person-packages.md) - external package selection.
