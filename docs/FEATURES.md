# agent-compose features

Inventory of what ships today; planned behavior lives in the issue tracker.

## Repository foundation

* Public, MIT-licensed AOS and Ward context substrate with personal policy.
* Audience-specific README, agent instructions, and shipped-feature inventory.
* Ward-gated repository validation backed by the agentic-os hook catalog.

## Composition engine

* Go CLI (`agent-compose compose`) turns a KDL request into an immutable
  bundle without network access.
* Embedded person source carrying the six active roles, the social and sales
  stubs, and twelve named agent seats with pronouns; personalities beyond the
  engineer set land with issue #10.
* Resolver emits the decision trace while choosing, covering selected,
  excluded, shadowed, and fallback outcomes.
* Atomic, content-keyed materialization: identical inputs reuse the cached
  bundle, failed finalization leaves no partial tree.
* Native-skill and compiled-context delivery with density-aware compiled
  prose, exercised by the four public fixtures.

## Load-point projection

* `agent-compose project` places bundle content at harness load points via
  the fixed v0.1 layout registry (claude, codex, goose, opencode).
* Sidecar-tracked ownership: projection never overwrites files it did not
  create, and re-projection removes only its own stale files.

## Launch-time refresh

* `compose ... -- <command>` refreshes (bundle projection, or full host
  convergence bare) then execs, sentinel-guarded against wrapper recursion.
* Refresh failure warns loudly and falls back to a validated last-known-good
  projection; concurrent launches converge on one cache entry via the
  materializer's rename race and a per-target projection lock.

## Inspection

* `agent-compose describe` renders a bundle's decision tree in scannable
  sections with collapse for large exclusion groups; `--why` follows one item
  from consideration to outcome, including what would have selected it.
* `agent-compose diff` reports semantic decision changes between two bundles;
  `compose --explain` appends the full tree to the one-screen summary.
* Color only on a TTY with NO_COLOR unset; redirected output stays plain and
  deterministic, and trace.json is the machine-readable surface.
* Favorite colors: personality hex colors gated at parse time into the
  terminal-legible OKLab band; the composed favorite rides the manifest, and
  multi-component favorites derive as the chroma-restored OKLab centroid.

## Roster artifact and cascade

* `agent-compose roster --out <dir>` renders the seat dispatch table as a
  cascade source: identity lines per seat, compatible personalities with
  favorite colors, personality bodies, and a claude `@`-import override,
  degrading to pending markers until #10 and writing under sidecar ownership.
* `agent-compose cascade` is the absorbed v1 composer: doctrine sources into
  per-harness COMPOSED files, load-point symlinks, scope and harness
  filtering, section overrides, the mount-eligibility manifest ward reads,
  and --dry-run/--check - byte-compatible with the Python outputs.
* Bare `acompose` converges roster and cascade; `-- <command>` refreshes before exec.
* Configured skill roots mount safely into native harness skill directories.

## Release

* Main pushes queue validated minor releases with cross-platform binaries,
  checksums, Homebrew, and Scoop output. See [release.md](release.md).

## Product status

The complete personality roster and Ward container consumer remain open.

## See also

* [../README.md](../README.md) - product boundary and current status.
* [../AGENTS.md](../AGENTS.md) - repo-specific operating rules.
* [../.ward/ward.yaml](../.ward/ward.yaml) - allowlisted development commands.
* [Catalog trifecta convention](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/src/branch/main/docs/features-release-tooling.md) - shared entry-point structure.
