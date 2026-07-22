# agent-compose features

Inventory of what this repository ships today. Planned behavior belongs in the
canonical issue tracker until an implementation lands.

## Repository foundation

* Public Forgejo-canonical repository under `coilyco-flight-deck`.
* Personal-product positioning: Kai's opinionated context system is published
  in the open rather than presented as a neutral enterprise framework.
* Audience-specific README, agent instructions, and shipped-feature inventory.
* Ward-gated repository validation backed by the agentic-os hook catalog.
* MIT licensing for the public source.

## Composition engine (first slice)

* Go CLI (`agent-compose compose`) turns a KDL request into an immutable
  bundle without network access.
* Embedded fixture-grade person source validating role-personality pairing;
  the full roster is issue #10.
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

## Product status

No complete personality roster, Ward consumer, or release artifact ships
yet.

## See also

* [../README.md](../README.md) - product boundary and current status.
* [../AGENTS.md](../AGENTS.md) - repo-specific operating rules.
* [../.ward/ward.yaml](../.ward/ward.yaml) - allowlisted development commands.
* [Catalog trifecta convention](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/src/branch/main/docs/features-release-tooling.md) - shared entry-point structure.
