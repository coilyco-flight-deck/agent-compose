# agent-compose features

Inventory of what ships today; planned behavior lives in the issue tracker.

## Repository foundation

* Public, MIT-licensed AOS and Ward substrate with personal policy.
* Audience-specific README, agent instructions, and shipped-feature inventory.
* Ward-gated repository validation backed by the agentic-os hook catalog.

## Composition engine

* Go CLI (`agent-compose compose`) turns a KDL request into an immutable
  bundle without network access.
* Embedded source carries ten roles, the approved meld matrix, 16 catalog
  bindings, and twelve named seats. A role activates its full ordered
  personality set; unavailable skill bodies stay pending in roster output.
* Resolver emits the decision trace while choosing, covering selected,
  excluded, shadowed, and fallback outcomes.
* Atomic materialization verifies staged and reused bundles; failures leave no partial tree.
* Native-skill and compiled-context delivery with density-aware compiled
  prose, exercised by the four public fixtures.

## Load-point projection

* `agent-compose project` transactionally places verified bundles at repo or
  container-home load points for claude, codex, goose, and opencode.
* Sidecar ownership protects foreign files and restores prior owned state.

## Launch-time refresh

* `compose ... -- <command>` refreshes (bundle projection, or full host
  convergence bare) then execs, sentinel-guarded against wrapper recursion.
* Refresh failure falls back to a validated last-known-good projection.
  Concurrent launches share cache and projection locks.

## Inspection

* `agent-compose describe` renders a bundle's decision tree in scannable
  sections with collapse for large exclusion groups; `--why` follows one item
  from consideration to outcome, including what would have selected it.
* `agent-compose diff` reports semantic changes; `verify` checks safe entry
  points, delivery data, trace integrity, and the complete selected identity
  set.
  `compose --explain` appends the full tree to the one-screen summary.
* Color only on a TTY with NO_COLOR unset; redirected output stays plain and
  deterministic, and trace.json is the machine-readable surface.
* Colors pass an OKLab legibility gate. Each role carries the chroma-restored centroid.

## Personality palette

* Local Vite/TypeScript explorer shows component colors, role melds, filters,
  previews, and copy controls. Ward derives its JSON from the embedded person source. See [personality-palette.md](personality-palette.md).

## Roster artifact and cascade

* `agent-compose roster --out <dir>` renders the seat dispatch table as a
  cascade source: seats, melded personalities, component and derived colors,
  bodies, and a claude `@`-import override. Missing bodies stay pending.
* `agent-compose cascade` is the absorbed v1 composer: doctrine sources into
  per-harness COMPOSED files, load-point symlinks, scope and harness
  filtering, section overrides, the mount-eligibility manifest ward reads,
  and --dry-run/--check - byte-compatible with the Python outputs.
* Bare `acompose` converges roster and cascade; `-- <command>` refreshes before exec.
* Configured skill roots mount safely into native harness skill directories.

## Release

* Main pushes publish binaries, checksums, Homebrew, and Scoop. See [release.md](release.md).

## Product status

AOS skill bodies and Ward-side container invocation remain open.

## See also

* [../README.md](../README.md) - product boundary and current status.
* [../AGENTS.md](../AGENTS.md) - repo-specific operating rules.
* [../.ward/ward.yaml](../.ward/ward.yaml) - allowlisted development commands.
* [Catalog trifecta convention](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/src/branch/main/docs/features-release-tooling.md) - shared entry-point structure.
