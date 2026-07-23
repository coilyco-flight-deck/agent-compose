# Architecture

Agent-compose is the context substrate between AOS and Ward. AOS authors
reusable doctrine, skills, and capability providers. Agent-compose selects,
compiles, and installs that knowledge for a harness. Ward supplies runtime facts
and executable authority without placing either inside the context bundle.

Personality is the first opinionated policy domain built on that substrate.
Agent-compose resolves a role into its personality set, ordinary provider
skills, and explicitly composed role skills, then materializes the combined
context as an immutable bundle. Native harnesses can use it without Ward.

## Composition inputs

The caller supplies every input. Agent-compose infers nothing about the agent:

* `role` - the caller names it, the person source validates it, and the role
  activates every personality it declares.
* `density` - how much personality prose the bundle carries. A caller usually
  derives it from model class: a frontier model needs one sentence where a
  small local model needs the one-pager. No model name enters a request.
* `delivery` - native skills or compiled context.
* `source locators` - where files live. AOS's knowledge inventory is inferred
  from its provider root.

Everything else about the agent stays outside the request. Model, harness,
reasoning effort, interactivity, permissions, and task acceptance belong to
the launcher and the consumer. A repository is not an agent-compose concept:
a repo is only a place personality files sometimes live, reached through a
source locator like any other directory.

## Policy ownership

Agent-compose embeds one canonical public-safe person source. That source owns
roles, personalities, and role-personality compatibility. External sources add
knowledge but cannot redefine those names. A private overlay may add scoped
instructions and selection rules.

Personality definitions live inside skills. Agent-compose discovers AOS's
invariant and ordinary `SKILL.md` trees from its root. `.agents/roles.kdl`
admits role-specific `.agents/composed/*/COMPOSED.md` sources. Overlays may
use explicit source declarations.

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

AOS publishes reusable doctrine, skills, capability providers, instructions,
and composed-skill bindings under stable relative paths. AOS carries no copy
of Kai's person source or Ward authority.

Ward may build the compose request and mount the finished bundle read-only,
treating the tree as opaque. Authority and credentials never enter the
request. Agent-compose is Ward's context substrate, not its permission engine.
Ward can run without composed context, and agent-compose can serve native
harnesses without Ward.

## See also

* [kdl-contracts.md](kdl-contracts.md) - human-authored input grammar.
* [person-contract.md](person-contract.md) - embedded public-safe policy.
* [bundle-protocol.md](bundle-protocol.md) - immutable output contract.
* [manifest-schema.md](manifest-schema.md) - stable manifest fields.
* [decision-trace.md](decision-trace.md) - retained decision evidence.
* [projection.md](projection.md) - the harness-aware load-point layer.
* [integration.md](integration.md) - the v1 cascade seam and delivery tiers.
* [contract-review.md](contract-review.md) - review decisions of record.
