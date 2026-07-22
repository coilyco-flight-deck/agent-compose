# v0.1 contract review

This is the human review record for issue #2. Kai reviewed the proposed
contract in issue #13 and the decisions below are the outcome. Implementation
issues consume this reviewed contract rather than the earlier proposal.

## Review decisions

* Agent-compose is a personality engine. It owns personality, source
  selection, and delivery. It is not a security boundary.
* Repositories are not an agent-compose concept. A repo is at best a place
  personality files happen to live, reached through a source locator like any
  other directory. Privacy scopes, target repositories, repo declarations, and
  per-repo capability resolution are removed from the contract.
* Agent, model, harness, reasoning effort, and interactivity are the realm of
  AOS and Ward and never enter a compose request. The one surviving
  harness-adjacent input is `density`, which a caller may derive from model
  class to size the personality prose - one sentence for a frontier model, a
  one-pager for a small local model. No model name reaches agent-compose.
* Delivery mode - native skills or compiled context - is load-bearing and
  stays.
* Personality definitions live inside skills. The person contract binds
  personality names to skills and drops the presence, attention, tempo, and
  voice fields. The format inside the skill (KDL, YAML, or a metadata block in
  SKILL.md) is deliberately undecided in v0.1.
* The person contract drops its invariants section and renames
  `allows-personality` to `personality`. The rule that personality never
  alters truthfulness, authority, safety, rollback, or completion remains a
  repo rule in AGENTS.md rather than schema surface.
* No schema-version fields and no digest ceremony. Immutability and atomic
  refresh remain because they are cheap and prevent partially replaced
  bundles, not because bundles are a trust artifact.
* The decision trace stays, but as a plain ordered list of decisions with
  human-readable reasons rather than a protocol-grade specification.
* Byte-identical duplicate content deduplicates; non-identical collisions for
  one delivery slot still fail in v0.1 instead of adding an override grammar.
* Agent identity entered the person contract as named seats - `agent` nodes
  with `name` and `pronouns` nested under each role, adapting the shape ward's
  roles.kdl comments sketched. Names are opaque strings to the engine; ward
  keeps guardfiles, models, and reasoning effort, joined by the shared role
  slug.

## Ward integration record

Ward may build the compose request and mount the resulting bundle read-only,
treating the tree as opaque. Authority claims, guardfiles, credentials,
permissions, mutable harness state, and task acceptance stay entirely on
Ward's side. Ward can run without agent-compose, and agent-compose can run
without Ward.

## AOS integration record

AOS publishes reusable personality skills and instructions under stable
source ids and relative paths. It does not publish Kai's person source,
harness load points, Ward policy, or installation paths. Agent-compose
resolves AOS declarations locally without fetching.

## Compatibility fixtures

* `native-full.kdl` - native skills at full density.
* `native-brief.kdl` - native skills at brief density.
* `compiled-full.kdl` - compiled context at full density.
* `compiled-brief.kdl` - compiled context at brief density.

The four fixtures prove that delivery mode and context density vary
independently without agent-compose knowing which harness or model sits
behind them.

## See also

* [architecture.md](architecture.md) - composition inputs and ownership.
* [kdl-contracts.md](kdl-contracts.md) - request and source grammar.
* [person-contract.md](person-contract.md) - embedded personal policy.
