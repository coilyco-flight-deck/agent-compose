# Global skill load points

`skill_load_points` names the harness-native directory each harness reads for
global skills. Converge links the compiled residency set into every wired
destination.

## Defaults

Claude and codex are wired by default, matching `load_points`:

* `claude` - `~/.claude/skills`
* `codex` - `~/.agents/skills`

Claude Code reads `.claude/skills` and never the portable `.agents/skills`
directory, so one shared path cannot serve both harnesses. Goose and opencode
do read the portable directory, but like their instruction load points they
stay opt-in through config.

## Overrides

A configured entry replaces its default. A null or false value opts that
harness out entirely, the same falsy rule `load_points` uses.

Naming one harness leaves the other on its default. Before this, config that
set only `codex` unwired claude silently, and a claude session started with no
global skills at all while codex had the full set.

## See also

* [cascade.md](cascade.md) - host composition and native skill roots.
* [projection.md](projection.md) - per-harness load points and home scope.
