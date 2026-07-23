# Cascade

The cascade (bare `agent-compose compose`, or the hidden `cascade` verb for scripting) is the absorbed v1 composer: it turns doctrine
sources into each harness's global context. Presence of
`~/.agent-compose/agent-compose.yaml` activates it; without the file
the verb is a documented no-op, so a host behaves exactly as if
agent-compose were not installed.

Bare `acompose` always summarizes its roster, outputs, harness load points,
eligibility manifest, and skill links. Detailed lines identify repaired drift.

All state lives under `~/.agent-compose`: the config, COMPOSED outputs, the
mount-eligibility manifest, `sources/` (including the roster artifact), and
the bundle cache at `bundles/`. A legacy `~/.config/agent-compose` directory
migrates wholesale on first use, leaving a compatibility symlink behind so
ward's manifest read and fleet config references keep resolving until the
fleet cutover tracked in agentic-os#618.

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
mount-eligibility manifest ward reads is emitted beside the composed output:
per harness, the repos backing its selected sources unioned with the default
mount set, as deterministic JSON.

`--dry-run` previews only real changes; `--check` verifies every output
against a fresh compose and fails with a diff on drift. Writes happen only
on change, so a converged host recomposes silently.

Behavior was cross-checked against the Python composer on shared fixtures:
outputs are byte-identical, including symlink-resolved link absolutization.

## Native skill roots

Bare compose can also link authored skill catalogs into harness-native skill
directories:

```yaml
skill_load_points:
  codex: ~/.codex/skills
```

Each harness uses the eligible repository paths already recorded in
`mount-eligibility.json`, including the default AOS and AOSK roots. A repository
contributes skills when it contains `.agents/skills`. Defaults compose first,
then additional eligible repositories in stable order. Existing unowned
entries at a load point always win. Agent-compose records its links in
`~/.agent-compose/skill-mounts.json` and removes only stale links that still
match that ownership record. Fleet pointer aggregation, conditional category
gating, and per-repo capability pulls remain rollout policy outside this
substrate operation.

## See also

* [integration.md](integration.md) - how roster and cascade fit together.
* [projection.md](projection.md) - repo and home load-point projection.
