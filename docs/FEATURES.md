# agent-compose features

Inventory of what this repository ships today. Planned behavior belongs in the
canonical issue tracker until an implementation lands.

## Repository foundation

* Public Forgejo-canonical repository under `coilyco-flight-deck`, MIT
  licensed, positioned as Kai's opinionated personal context system.
* Audience-specific README, agent instructions, and shipped-feature inventory.
* Ward-gated repository validation backed by the agentic-os hook catalog.

## Composition engine (first slice)

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

* `agent-compose launch` refreshes (compose plus project) then execs the real
  command, with an environment sentinel preventing wrapper recursion.
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

## Roster artifact

* `agent-compose roster --out <dir>` renders the seat dispatch table as a
  v1-cascade source: every named seat's identity line, compatible
  personalities with favorite colors, personality bodies from the given
  sources, and a claude override with mechanical `@` imports.
* Degrades gracefully - roles without personalities and personalities
  without bodies render as pending until #10 - and writes under the same
  sidecar ownership rules as projection.

## Release

* Semver git tags with a stamped `version` verb; `ward exec release-build`
  cross-compiles darwin-arm64 and linux-amd64/arm64 into `dist/`.
* v0.1.0 is published on Forgejo with all three binaries attached.

## Product status

No complete personality roster or Ward consumer ships yet.

## See also

* [../README.md](../README.md) - product boundary and current status.
* [../AGENTS.md](../AGENTS.md) - repo-specific operating rules.
* [../.ward/ward.yaml](../.ward/ward.yaml) - allowlisted development commands.
* [Catalog trifecta convention](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/src/branch/main/docs/features-release-tooling.md) - shared entry-point structure.
