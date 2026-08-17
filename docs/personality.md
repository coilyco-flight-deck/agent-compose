# Personalities

## Person profiles and personality libraries

Agent Compose selects exactly one local person profile. A profile owns role
skills, structured role metadata, role identity, seats, the invariant, copy contracts, and
optional role evaluation matrices.

### Profile layout

```text
person.kdl
roles/NN-role.kdl
roles/<role>/SKILL.md
personalities/NN-local.kdl
[inspirations/NN-inspiration.kdl]
definitions/INVARIANT.md
definitions/skills/<skill>/SKILL.md
evaluations/<role>.yaml
libraries/<local-library>/
```

The established complete-person layout remains valid during v1.x. Its local
personalities act as an implicit package-local library.

### Library layout

```text
library.kdl
personalities/NN-personality.kdl
[inspirations/NN-inspiration.kdl]
definitions/skills/<skill>/SKILL.md
```

`library.kdl` has one stable logical library name. Libraries contain only
personality-owned content. They do not own roles or a profile invariant.
The inspirations directory is optional compatibility data. If present, its
records and personality references are validated as one complete graph.

### Admission and ordering

The profile discovers `libraries/` children in lexical order. Callers may admit
further local roots with repeatable `--personality-library` flags,
`personality-library` request nodes, or ordered host `personality_libraries`.
Profile-local roots resolve first, followed by the caller order. All roots are
local directories. Agent Compose does not accept URLs, git refs, release
identifiers, or fetch instructions.

### Conflicts and compatibility

Roles reference personality slugs, not a library name. A profile may therefore
boundary local and admitted personalities. A role may have any nonempty ordered
boundary, including one personality. Byte-identical personality definitions
deduplicate. A divergent duplicate slug or duplicate skill binding fails before
materialization. Missing personality references fail after all admitted
libraries merge. Alias collisions remain visible as ordered candidate sets.
Generated v1-compatible `person.json` remains available. The additive
`person.v4.json` and `personality-index.md` provide aliases, affinities,
logical provenance, and derived boundaries for consumers that need them.

### Cues and affinities

Libraries declare aliases with `alias "cue"` inside a personality entry. Lookup
applies Unicode NFKC, lowercases, trims surrounding whitespace, and normalizes
whitespace, underscores, and hyphens to one hyphen. A canonical slug match
wins. Otherwise every matching alias candidate remains visible in deterministic
catalogue order. Affinities derive from the effective profile only. Each v4
personality entry records its roles and each complete ordered meld, or an empty
affinity list when no selected role uses it. A cue never changes a role,
authority, permissions, or the native confirmation and lifetime rules for an
interactive personality swap.

## Personality palette explorer

The repository ships a local visual explorer for the canonical personality
colors and each role's complete meld. It preserves color as expression only.
Authority, safety, and completion remain outside personality.

### Run it

From the repository root:

```sh
just palette-serve
```

Ward generates `web/personality-palette/public/palette.json` from the embedded
person source, installs the pinned browser toolchain, and starts Vite on its
local address. The JSON, dependency directory, and production build stay
uncommitted. The remaining lifecycle verbs are:

* `just palette-build` - type-check and build the production assets.
* `just palette-test` - build and verify the generated artifact.
* `just palette-tidy` - reconcile `package-lock.json`.

`just test` includes the palette test, so every release validates the
committed explorer.

### Data ownership and interaction

The embedded KDL person source remains the only owner of personality colors,
identity primitives, role membership, role order, and boundary inputs. The
hidden `palette-data` command projects that source into versioned JSON. The Go
color package derives each role's boundary before the browser sees it. The
TypeScript layer owns only visual presentation metadata such as friendly color
names, short associations, and spectrum ordering. Startup validation fails
visibly if that presentation list drifts from the canonical catalog.

The explorer provides:

* the full eighteen-personality spectrum with emblem, motif, form, and sound
* role filters with complete personality melds
* component colors and the derived role boundary
* day and night previews
* spectrum and alphabetical ordering
* one-click copying for component and melded colors
* responsive layout and reduced-motion support

The app is a framework-free Vite and TypeScript project under
`web/personality-palette`. It is a local source tool, not a deployment target.
