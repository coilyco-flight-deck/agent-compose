# Agent Compose v2 roster migration

Agent Compose v2 renames the baked provider from `person:kai` to
`roster:core` and emits the Core Roster slugs, including AI Engineer as `ai`.
This is an intentional major-version break. There are no compatibility aliases
for old role identifiers. The earlier fixed eight-role decision changed when
portfolio evidence established distinct AI Engineer, Outreach, Sales, and
Content ownership boundaries.

## Role destinations

* `engineer` - remains `engineer`.
* `director` - remains `director`.
* `qa` - remains `qa`.
* `ops` - remains `ops`, displayed as DevOps.
* `designer` - becomes `design`, displayed as Designer.
* `community` - remains `community`, displayed as Community Manager.
* `advisor`, `ceo`, and `pm` - become `strats`, displayed as Portfolio
  Strategist.
* `technical-writer` and `social` - become `content`, displayed as Content
  Manager.
* `sales` - is restored as Sales, with `customer-success` still requiring the
  actual mission owner or an external roster package.

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
