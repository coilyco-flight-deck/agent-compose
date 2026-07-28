# Remote skill catalogs

Before composing host context, agent-compose can hydrate ordinary skill
catalogs from Git and project them through the same native load points as local
eligible repositories. The feature belongs to bare host convergence. Bundle
requests and person-package loading remain local and deterministic.

## Configuration

Declare repositories in `~/.agent-compose/agent-compose.yaml`:

```yaml
skill_load_points:
  claude: ~/.claude/skills
  codex: ~/.agents/skills

remote_skill_cache_ttl: 10m
remote_skill_sources:
  - url: https://example.com/owner/shared-skills.git
    ref: main
  - url: ssh://git@example.com/owner/private-skills.git
    ref: v2:catalog/skills
```

`url` is required. `ref` uses Git's native `<tree-ish>:<path>` object syntax.
The tree-ish accepts a branch, tag, or full commit ID. Omitting `ref` selects
`HEAD:.agents/skills`. A bare tree-ish such as `main` also selects the
conventional `.agents/skills` path. Use `main:skills` or
`v2:catalog/skills` when an upstream repository keeps its catalog elsewhere.

Remote catalogs are harness-neutral inputs. Every configured catalog projects
to every configured `skill_load_points` destination after local eligible
repositories.

`remote_skill_cache_ttl` accepts a Go duration and defaults to `10m`. `0s`
refreshes movable refs on every convergence.
Git uses the calling process's normal credential chain. Configuration names a
clone URL and never embeds a credential.

## Cache lifecycle

Each normalized URL and Git revision-path receives its own SHA-256 cache key
under `~/.agent-compose/cache/remote-skills/`. The cache stores a bare mirror
and one verified working checkout. The URL and locator do not become directory
names.

One source lock serializes mirror and checkout changes across concurrent
convergences. A fresh mirror plus a matching verified checkout performs no
network or filesystem work. Once the TTL elapses, agent-compose refreshes the
mirror and replaces the working checkout only after the requested ref resolves
and the catalog path exists. A full commit ID already present in the mirror is
immutable, so it skips the refresh.

## Failure behavior

A source that has never hydrated must clone, resolve its ref, and expose a
valid catalog directory before native skill links change. Clone errors,
unknown refs, and missing or escaping catalog paths fail the convergence.

If a stale refresh fails after an earlier successful hydration, agent-compose
warns and serves the last verified checkout. This keeps launches usable
offline without hiding first-use failures. The next convergence after the TTL
tries the refresh again.

## Projection and precedence

Local eligible repositories keep their existing order. Remote catalogs apply
after them in configuration order, so a later remote catalog wins a duplicate
skill name at every configured load point.

Existing unowned entries at a native load point still win over every managed
catalog. Agent-compose removes a stale managed link only when its target still
matches the ownership sidecar. Remote hydration therefore changes where skills
can come from without weakening local projection safety.

## See also

* [cascade.md](cascade.md) - host composition and native skill load points.
* [FEATURES.md](FEATURES.md) - shipped capability inventory.
