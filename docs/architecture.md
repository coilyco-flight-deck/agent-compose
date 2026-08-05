# Architecture

Agent-compose sits between knowledge providers and context consumers. Providers
author reusable knowledge. Agent-compose selects and installs it for a harness.
Consumers supply runtime facts and executable authority outside the bundle.

## Composition inputs

The caller supplies every input. Agent-compose infers nothing about the agent:

* `person source` - the caller may name one complete external package. Omission
  selects the embedded `roster:core` default.
* `role` - the caller names it, the selected person source validates it, and the role
  activates every personality it declares.
* `model tier` - the caller selects `frontier`, `commodity`, or `oss`, and the
  role may reject it before bundle materialization.
* `model class` - the caller independently selects full `frontier` or pruned
  `low-context` bundle density.
* `delivery` - native skills or compiled context.
* `source locators` - where optional capability files live. AOS's knowledge
  inventory is inferred from its provider root.

Model identity, harness, reasoning effort, interactivity, permissions, and task
acceptance stay with the launcher and consumer. Repositories only host
capability files reached through source locators.
Consumers map concrete models into the stable tiers. Claude and Codex are
frontier examples, DeepSeek is commodity, and Ornith or Mistral are OSS.
The Core Roster applies those tiers through its
[role compatibility matrix](model-tiers.md).

## Policy ownership

Agent-compose embeds the eight-role Core Roster, `roster:core`, as its
public-safe default. A caller may
select one complete [external package](person-packages.md) instead. Selection
is exclusive. The package owns roles, personalities, compatibility, invariant,
definitions, and inspirations. Capability sources add knowledge but cannot
redefine those names or bodies.

Personality definitions live inside person-package skills. Agent-compose
discovers ordinary skills and a trusted `.agents/roles.kdl` graph for
role-only providers and composed skills. Imported graphs do not recurse. Overlays may use
explicit source declarations. An optional legacy AOS invariant and personality
copy remains readable during rolling upgrades. Byte-identical copies shadow
behind the selected person source.

The resolver evaluates admitted private overlays in request order, then AOS
sources in request order. Byte-identical candidates for one delivery slot
deduplicate to the highest-precedence copy. Non-identical collisions fail in
v0.1 rather than adding an override grammar.

## Composition flow

Agent-compose loads exactly one person source, validates the role, selects
matching instructions, ordinary skills, active personalities, and composed
role skills, then chooses delivery and materializes the bundle. It records
what it picked and why as each decision occurs.

## Integration obligations

Knowledge providers publish reusable doctrine, ordinary skills, capability
sources, instructions, and composed-skill bindings under stable relative
paths. They carry no copy of the selected person source or personality
definitions.

A consumer may build the compose request and adapt a verified home projection,
treating the source bundle as immutable. Authority and credentials never enter
the request. Agent-compose is a context producer, not a permission engine.
Consumers can run without composed context, and agent-compose can serve native
harnesses without a composition adapter.

## See also

* [kdl-contracts.md](kdl-contracts.md) - human-authored input grammar.
* [person-contract.md](person-contract.md) - validated person-package policy.
* [person-packages.md](person-packages.md) - isolated external package boundary.
* [bundle-protocol.md](bundle-protocol.md) - immutable output contract.
* [decision-trace.md](decision-trace.md) - retained decision evidence.
* [projection.md](projection.md) - the harness-aware load-point layer.
* [integration.md](integration.md) - the v1 cascade seam and delivery tiers.
