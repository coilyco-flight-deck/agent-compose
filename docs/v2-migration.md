# Agent Compose v2 roster migration

Agent Compose v2 renames the baked provider from `person:kai` to
`roster:core` and emits the Core Roster slugs, including AI Engineer as `ai`.
This is an intentional major-version break. There are no compatibility aliases
for old role identifiers. The Core Roster now returns to eight roles after
audience-facing work proved more coherent as one Content Creator feedback loop.

## Role destinations

* `engineer` - remains `engineer`.
* `director` - remains `director`.
* `qa` - remains `qa`.
* `ops` - remains `ops`, displayed as DevOps.
* `designer` - becomes `design`, displayed as Designer.
* `advisor`, `ceo`, and `pm` - become `exec`, displayed as Executive
  Strategist.
* `technical-writer`, `social`, `content`, `community`, `outreach`,
  `sales`, and `customer-success` become `creator`, displayed as Content
  Creator.

Update launch commands, Ward role selections, composed-skill bindings,
evaluation inputs, and any bundle source checks to use the new identifiers.
Existing v1 scored records remain historical evidence and retain their
original role and provider identities.

Move `remote_skill_sources`, `remote_skill_cache_ttl`, and `mcp_inventory`
configuration to AOS before installing v2. AOS hydrates and verifies remote
catalogues, projects native MCP and Codex approval policy, then passes
`skill_catalog_manifest` to Agent Compose. Removed v1 keys fail strict config
loading instead of being ignored.

## External packages

External `person "<name>"` packages remain supported through the validated
package contract and keep `person:<name>` provenance. A package may adopt a
`roster "<name>"` manifest when it intends to emit `roster:<name>` provenance.
The package selector remains exclusive and never inherits Core Roster roles or
definitions.
