# Person profiles and personality libraries

Agent Compose selects exactly one local person profile. A profile owns roles,
seats, briefings, the invariant, copy contracts, role inspirations, and
optional role evaluation matrices.

## Profile layout

```text
person.kdl
roles/NN-role.kdl
personalities/NN-local.kdl
inspirations/NN-inspiration.kdl
definitions/INVARIANT.md
definitions/skills/<skill>/SKILL.md
evaluations/<role>.yaml
libraries/<local-library>/
```

The established complete-person layout remains valid during v1.x. Its local
personalities act as an implicit package-local library.

## Library layout

```text
library.kdl
personalities/NN-personality.kdl
inspirations/NN-inspiration.kdl
definitions/skills/<skill>/SKILL.md
```

`library.kdl` has one stable logical library name. Libraries contain only
personality-owned content. They do not own roles or a profile invariant.

## Admission and ordering

The profile discovers `libraries/` children in lexical order. Callers may
admit further local roots with repeatable `--personality-library` flags,
`personality-library` request nodes, or ordered host
`personality_libraries`. Profile-local roots resolve first, followed by the
caller order.

All roots are local directories. Agent Compose does not accept URLs, git refs,
release identifiers, or fetch instructions.

## Conflicts and compatibility

Roles reference personality slugs, not a library name. A profile may therefore
meld local and admitted personalities. A role may have any nonempty ordered
meld, including one personality.

Byte-identical personality definitions deduplicate. A divergent duplicate slug
or duplicate skill binding fails before materialization. Missing personality
references fail after all admitted libraries merge. Alias collisions remain
visible as ordered candidate sets.

Generated v1-compatible `person.json` remains available. The additive
`person.v4.json` and `personality-index.md` provide aliases, affinities,
logical provenance, and derived melds for consumers that need them.
