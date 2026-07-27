# agent-compose features

Inventory of what ships today. Planned behavior lives in the issue tracker.

## Repository foundation

* Public MIT source with Ward-gated AOS validation and catalog docs.

## Composition engine

* `agent-compose compose` turns a KDL request into an immutable offline bundle.
* `person:kai` carries a portfolio-native default with 12 roles, 16
  personalities, melds, and 40 seats. It grounds work in Kai's real projects
  while treating commercial paths as evidence-qualified possibilities.
* [External person packages](person-packages.md) and
  [local personality libraries](personality-libraries.md) replace the default
  across every person-dependent command.
* AOS roots expose ordinary skills. `.agents/roles.kdl` owns composed admission.
  `low-context: optional` prunes optional skills.
* Role compatibility fails closed before materialization. CEO supports only
  `frontier`, while its failed OSS evaluation remains the re-enable gate.
* Materialization promotes admitted `COMPOSED.md` files to native `SKILL.md`.
* Resolver traces selected, excluded, shadowed, and delivered outcomes.
* Atomic materialization verifies staged and reused bundles.
* Both delivery modes use canonical personality bodies. Identical legacy
  copies shadow during upgrades and divergent copies fail closed.

## Load-point projection

* `agent-compose project` places verified bundles transactionally at repo or
  container-home load points for all four harnesses.
* Sidecar ownership protects foreign files and restores prior owned state.

## Launch-time refresh

* `compose ... -- <command>` refreshes then execs, guarded against recursion.
* Refresh uses validated fallback unless `external-only` forbids it.
  Concurrent launches share locks.

## Inspection

* `agent-compose describe` renders a collapsible decision tree. `--why`
  follows one item from consideration to outcome.
* `agent-compose diff` reports semantic changes. `verify` checks entry points,
  delivery, trace integrity, and selected identities.
* `compose` renders complete role metadata. `--explain` adds decisions.
* [Evaluation](evaluation.md) - guarded person review packs and multi-model baselines.
* TTY colors use canonical identity. Redirects and `NO_COLOR` stay plain.
* Colors pass an OKLab legibility gate. Each role carries the chroma-restored centroid.

## Identity surfaces

* The local [palette](personality-palette.md) shows melds and identity primitives.
  `agent-compose overlay` emits text or JSON for one selected member. See [overlay.md](overlay.md).

## Roster artifact and cascade

* `agent-compose roster --out <dir>` renders definitions, seats, briefings,
  colors, and confirmed [native personality swaps](native-personality-swaps.md).
* Bare convergence emits deterministic [`person.json`](person-snapshot.md) for
  roles, seats, compatibility, identities, inspirations, and appearances.
* `agent-compose cascade` is the absorbed v1 composer: doctrine sources into
  per-harness files, symlinks, filtering, overrides, a mount manifest,
  and dry-run/check behavior compatible with the Python outputs.
* Bare `acompose` converges hosts. `--reapply` forces the layout, `--verbose`
  traces `source => destination`, and `-- <command>` refreshes then execs.
  Ward smoke proves idempotence and its test verb runs full validation.
* Configured skill roots mount safely into native harness skill directories.
* [Native MCP + approval projection](native-mcp.md).

## Release

* Main pushes publish binaries, checksums, Homebrew, and Scoop. See [release.md](release.md).

## See also

* [../README.md](../README.md) - product boundary and current status.
* [../AGENTS.md](../AGENTS.md) - repo-specific operating rules.
* [../.ward/ward.yaml](../.ward/ward.yaml) - allowlisted development commands.
* [Catalog trifecta convention](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/src/branch/main/docs/features-release-tooling.md) - shared entry-point structure.
