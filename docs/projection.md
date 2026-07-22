# Load-point projection

Composition is harness-blind. Projection is the one layer that knows harness
vocabulary: it places a materialized bundle's content at the load points a
harness actually reads, beneath a chosen target directory.

```
agent-compose project <bundle-dir> --layout <name> --target <dir>
```

## v0.1 layout registry

* `claude` - native-skills: instructions to `CLAUDE.md`, each selected skill
  tree to `.claude/skills/<skill-id>/`.
* `codex` - native-skills: instructions to `AGENTS.md`, skills to
  `.agents/skills/<skill-id>/`.
* `goose` - compiled: the compiled context document to `.goosehints`.
* `opencode` - compiled: the compiled context document to `AGENTS.md`.

A layout requires the matching bundle delivery mode and fails with a
diagnostic otherwise. Layout names and load-point paths live only in this
layer; they never appear in the resolver, the request, the manifest, or the
bundle tree.

## Ownership and safety

Projection records every file it writes in `.agent-compose/projection.json`
beneath the target. It refuses to overwrite any file it did not create, so a
hand-authored CLAUDE.md or AGENTS.md is never clobbered. Re-projection
replaces its own previous files, removes ones no longer projected, prunes the
directories they emptied, and leaves foreign files untouched. The bundle
itself is read-only input and is never modified.

## See also

* [bundle-protocol.md](bundle-protocol.md) - the tree projection consumes.
* [manifest-schema.md](manifest-schema.md) - the entry points it reads.
* [architecture.md](architecture.md) - why composition stays harness-blind.
