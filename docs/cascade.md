# Cascade

The cascade turns doctrine sources into each harness's global context when
`~/.agent-compose/agent-compose.yaml` exists. Missing config is a no-op.

Bare `acompose` summarizes its roster, outputs, load points, manifest, skill
links, and repaired drift.
Bare `acompose --reapply` recreates outputs and load-point links.
`acompose --verbose` emits each source, override, manifest, and link as
`source => destination`.

`person_policy: external-only` requires `person_source`. A bad package aborts
before roster or cascade projection can restore the embedded default.

All state lives under `~/.agent-compose`: config, outputs, manifest, roster,
and cache. A legacy `~/.config/agent-compose` migrates on first use and leaves
a compatibility symlink through the cutover tracked in agentic-os#618.

Explicit `sources` compose first in listed order, then each `roots` entry is
walked for `AGENTS.COMPOSE.md` files, appended sorted. That filename is the
disjoint-source convention: always-global doctrine that no harness's own
AGENTS.md/CLAUDE.md cascade loads, so composing it in never double-loads.

## Selection and rewrites

A machine may declare `scopes`; a source declares its own in YAML
frontmatter and composes only when the two intersect. Omitting the machine
key disables filtering entirely; under active filtering an untagged source
never leaks in. Frontmatter `harnesses` restricts a source to named
harnesses. Composed bodies are rewritten for their new home: frontmatter
stripped, `## See also` navigation dropped, and relative markdown links
absolutized against the source's own directory.

A sibling `AGENTS.<harness>.md` beside a source patches it for one harness:
sections replace by verbatim heading, new headings append, and an ambiguous
heading fails the compose loudly. When harness slices diverge - by selection
or by override - output splits into `COMPOSED.<harness>.md` files; identical
slices share one `COMPOSED.md`, and obsolete banner-carrying outputs are
removed on convergence.

## Outputs

Each configured load point (claude and codex by default, others via
`load_points`, `null` to opt out) is symlinked at its harness's composed
file, backing up any pre-existing regular file to `.bak`. The
mount-eligibility manifest is emitted beside the composed output:
per harness, the repos backing its selected sources unioned with the default
mount set, as deterministic JSON.

`--dry-run` previews only real changes; `--check` verifies every output
against a fresh compose and fails with a diff on drift. Writes happen only
on change, so a converged host recomposes silently.

## Native skill roots

Bare compose can also link authored skill catalogs into harness-native skill
directories through `skill_load_points`, such as `codex: ~/.agents/skills`.

Each harness uses the eligible repository paths already recorded in
`mount-eligibility.json`, including the default AOS and AOSK roots. A repository
contributes skills when it contains `.agents/skills`. Defaults compose first,
then additional eligible repositories in stable order, followed by configured
remote catalogs in declaration order. Existing unowned entries at a load point
always win. Agent-compose records its links in
`~/.agent-compose/skill-mounts.json` and removes only stale links that still
match that ownership record. Fleet pointer aggregation, conditional category
gating, and per-repo capability pulls remain rollout policy outside this
substrate operation.

`remote_skill_sources` adds Git catalogs to those load points.
`ref`, `path`, and `harnesses` are optional. See
[remote skills](remote-skills.md) for configuration, caching, and failures.

## See also

* [integration.md](integration.md) - how roster and cascade fit together.
* [projection.md](projection.md) - repo and home load-point projection.
* [remote-skills.md](remote-skills.md) - Git hydration, caching, and offline reuse.
