# agent-compose features

Inventory of what this repository ships today. Planned behavior belongs in the
canonical issue tracker until an implementation lands.

## Repository foundation

* Public Forgejo-canonical repository under `coilyco-flight-deck`.
* Audience-specific README, agent instructions, and shipped-feature inventory.
* Ward-gated repository validation backed by the agentic-os hook catalog.
* MIT licensing for reuse as a generic context-composition tool.

## Product status

No compiler, KDL schema, bundle protocol, harness adapter, cache, or release
artifact ships yet.

## See also

* [../README.md](../README.md) - product boundary and current status.
* [../AGENTS.md](../AGENTS.md) - repo-specific operating rules.
* [../.ward/ward.yaml](../.ward/ward.yaml) - allowlisted development commands.
* [Catalog trifecta convention](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/src/branch/main/docs/features-release-tooling.md) - shared entry-point structure.
