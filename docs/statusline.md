# Projected composition status line

`acompose statusline` is Agent Compose's compact renderer for the immutable
bundle active at a repository or staged-home projection. It is read-only and
self-suppresses when no projection applies.

The ordinary row carries the composition facts worth keeping visible:

```text
🧭 🪨 📐 ⛏️  terran engineer · engineer@codex · frontier · 99 skills / ~96k catalog · ✓ composed
```

* Personality emblems and the named seat come from identity metadata retained
  in the selected immutable bundle.
* `role@harness` names the actual projection choice instead of inferring role
  from the current task.
* Model class records the compatibility boundary Agent Compose evaluated. It
  is not a model route or runtime permission.
* The skill and token footprint comes from selected provider reports in the
  stored decision trace. The token value measures the discoverable selected
  catalogue, including lazy skill content. It is not prompt-context usage.
* Health counts only sources classified as warnings during composition.
  Providers excluded because they belong to another role remain ordinary
  decisions and do not create false alerts.

The command walks upward from `--target` until it finds
`.agent-compose/projection.json`, then reads that projection's bundle manifest
and decision trace. It never recomposes, refreshes, verifies every bundle file,
or reads a mutable person source. New bundles retain the renderer metadata so
the row cannot drift from the context the agent actually received. Older
bundles degrade to the role and harness without inventing a seat or emblems.

Use `--color` when a status-line composer captures stdout through a pipe but
still accepts ANSI color. Direct redirected output remains plain by default.

## Provider integration

A status-line composer should invoke:

```text
acompose statusline --target <project-directory> --color
```

Agent Compose owns the rendered payload and bundle semantics. The caller owns
provider discovery, the project-directory runtime fact, row ordering, and
whether the output is shown. A missing projection emits no output. A recorded
projection with an unreadable manifest or trace emits a compact warning row.

## See also

* [Native role launch](native-role-launch.md) - how assigned bundles are composed and projected.
* [Projection](projection.md) - load-point and ownership sidecar contract.
* [Catalogues and export](catalogues-and-export.md) - detailed inspection beyond the compact row.
* [FEATURES.md](FEATURES.md) - shipped capability inventory.
