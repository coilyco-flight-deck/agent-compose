# Architecture

Agent-compose sits between knowledge providers and context consumers. Providers
author reusable knowledge. Agent-compose selects and installs it for a harness.
Consumers supply runtime facts and executable authority outside the bundle.

Personality is the first opinionated policy domain built on that substrate.
Agent-compose resolves a role into its personality set, ordinary provider
skills, and explicitly composed role skills, then materializes the combined
context as an immutable bundle. Native harnesses can use it without an
isolated launch consumer.

## Composition inputs

The caller supplies every input. Agent-compose infers nothing about the agent:

* `role` - the caller names it, the person source validates it, and the role
  activates every personality it declares.
* `model class` - the caller names it and embedded role compatibility may
  reject it before bundle materialization.
* `delivery` - native skills or compiled context.
* `source locators` - where optional capability files live. AOS's knowledge
  inventory is inferred from its provider root.

Model identity, harness, reasoning effort, interactivity, permissions, and task
acceptance stay with the launcher and consumer. Repositories only host
capability files reached through source locators.

## Policy ownership

Agent-compose embeds one canonical public-safe person source set. Its small
manifest and ordered per-node KDL files own roles, personalities,
role-personality compatibility, the personality invariant, every canonical
personality definition, and the credited inspiration catalogue. External
sources add knowledge but cannot redefine those names or bodies. A private
overlay may add scoped instructions and selection rules.

Personality definitions live inside embedded skills. Agent-compose discovers
ordinary `SKILL.md` trees from an AOS root. `.agents/roles.kdl` admits
role-specific `.agents/composed/*/COMPOSED.md` sources. Overlays may use
explicit source declarations. An optional legacy AOS invariant and personality
copy remains readable during rolling upgrades. Byte-identical copies shadow
behind `person:kai`.

The resolver evaluates admitted private overlays in request order, then AOS
sources in request order. Byte-identical candidates for one delivery slot
deduplicate to the highest-precedence copy. Non-identical collisions fail in
v0.1 rather than adding an override grammar.

## Composition flow

Agent-compose loads the embedded person source, validates the role, selects
matching instructions, ordinary skills, active personalities, and composed
role skills, then chooses delivery and materializes the bundle. It records
what it picked and why as each decision occurs.

## Integration obligations

Knowledge providers publish reusable doctrine, ordinary skills, capability
sources, instructions, and composed-skill bindings under stable relative
paths. They carry no copy of Kai's person source or personality definitions.

A consumer may build the compose request and adapt a verified home projection,
treating the source bundle as immutable. Authority and credentials never enter
the request. Agent-compose is a context producer, not a permission engine.
Consumers can run without composed context, and agent-compose can serve native
harnesses without a composition adapter.

## See also

* [kdl-contracts.md](kdl-contracts.md) - human-authored input grammar.
* [person-contract.md](person-contract.md) - embedded public-safe policy.
* [bundle-protocol.md](bundle-protocol.md) - immutable output contract.
* [manifest-schema.md](manifest-schema.md) - stable manifest fields.
* [decision-trace.md](decision-trace.md) - retained decision evidence.
* [projection.md](projection.md) - the harness-aware load-point layer.
* [staged-home.md](staged-home.md) - launcher-neutral adapter boundary.
* [integration.md](integration.md) - the v1 cascade seam and delivery tiers.
* [contract-review.md](contract-review.md) - review decisions of record.
