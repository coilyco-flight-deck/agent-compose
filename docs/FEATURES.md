# agent-compose features

Inventory of what ships today. Planned behavior lives in the issue tracker.

## Repository foundation

* Public, MIT-licensed context substrate with personal policy.
* Audience-specific README, agent instructions, and shipped-feature inventory.
* Ward-gated repository validation backed by the agentic-os hook catalog.

## Composition engine

* `agent-compose compose` turns a KDL request into an immutable offline bundle.
* Embedded `person:kai` source carries ten roles, required long-form briefings,
  the approved meld matrix, 16 catalog bindings and definitions,
  the personality invariant, twelve named seats, and the normalized
  [inspiration catalogue](inspiration-catalogue.md). Every bundle prefixes
  instructions with the selected role's complete briefing.
* AOS provider roots expose ordinary skills. `.agents/roles.kdl` owns composed
  skill admission and intent-to-harness routes.
* Materialization promotes admitted `COMPOSED.md` files to native `SKILL.md`.
* Resolver traces selected, excluded, shadowed, and delivered outcomes.
* Atomic materialization verifies staged and reused bundles.
* Native-skill and compiled-context delivery use the same embedded canonical
  personality bodies. Identical legacy provider copies shadow during rolling
  upgrades and divergent copies fail closed.

## Load-point projection

* `agent-compose project` transactionally places verified bundles at repo or
  container-home load points for claude, codex, goose, and opencode.
* Sidecar ownership protects foreign files and restores prior owned state.

## Launch-time refresh

* `compose ... -- <command>` refreshes then execs, guarded against recursion.
* Refresh failure falls back to a validated last-known-good projection.
  Concurrent launches share cache and projection locks.

## Inspection

* `agent-compose describe` renders a collapsible decision tree. `--why`
  follows one item from consideration to outcome.
* `agent-compose diff` reports semantic changes. `verify` checks safe entry
  points, delivery data, trace integrity, and the complete selected identity
  set.
  `compose --explain` appends the full tree to the one-screen summary.
* TTY color honors NO_COLOR. Redirected output is plain and deterministic.
* Colors pass an OKLab legibility gate. Each role carries the chroma-restored centroid.

## Personality palette

* Local Vite/TypeScript explorer shows component colors, role melds, filters,
  previews, and copy controls. See [personality-palette.md](personality-palette.md).

## Roster artifact and cascade

* `agent-compose roster --out <dir>` renders the embedded invariant and
  definitions, overlay instructions, seats, long-form role briefings, melded
  personalities, colors, bodies, and a claude `@`-import override. See
  [role-briefings.md](role-briefings.md).
* `agent-compose cascade` is the absorbed v1 composer: doctrine sources into
  per-harness COMPOSED files, load-point symlinks, scope and harness
  filtering, section overrides, the mount-eligibility manifest ward reads,
  and --dry-run/--check - byte-compatible with the Python outputs.
* Bare `acompose` converges roster and cascade. `-- <command>` refreshes before exec.
* Configured skill roots mount safely into native harness skill directories.
* [Native MCP + approval projection](native-mcp.md).

## Release

* Main pushes publish binaries, checksums, Homebrew, and Scoop. See [release.md](release.md).

## See also

* [../README.md](../README.md) - product boundary and current status.
* [../AGENTS.md](../AGENTS.md) - repo-specific operating rules.
* [../.ward/ward.yaml](../.ward/ward.yaml) - allowlisted development commands.
* [Catalog trifecta convention](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/src/branch/main/docs/features-release-tooling.md) - shared entry-point structure.
