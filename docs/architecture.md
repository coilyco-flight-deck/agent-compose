# v0.1 architecture

Agent-compose resolves explicit agent facts and human-authored policy into an
immutable context bundle. The product works alone. A launcher such as Ward may
mount the same bundle beside its own independent authority surface.

## Layer boundary

The cli-guard `pkg/agentclaim` package owns the shared structural vocabulary.
It treats role names as opaque and separates `context` from `authority` roles.

Agent-compose requires one `context` role and rejects an `authority` role. Ward
may resolve authority separately. Matching names create no grant or roster.

## Input facts and owners

The caller supplies resolved facts. Agent-compose never infers one from another:

* `agent`, `model`, `model class`, `harness`, `reasoning effort`, and `mode` -
  the launcher owns them. Model class controls density.
* `context role` - the caller owns it. The person source validates it.
* `personality` - the caller owns it. The person source validates the pairing.
* `privacy scopes` - the operator or deployment owns source-admission tags.
* `target repositories` - the caller owns ordered slugs and declaration locators.
* `source locators` - the installation owns local AOS and overlay paths.

Locators never enter identity. Stable source ids and digests replace them in
deterministic evidence.

Role, personality, model, model class, harness, reasoning effort, permissions,
and task acceptance remain separate dimensions. Permission policy and task
acceptance are deliberately absent from the request. A consumer owns both
after composition.

## Policy ownership

Agent-compose compiles the schema and embeds one canonical public-safe person
source. That source owns a reserved namespace for roles, personalities,
compatibility, and invariants. No external source can replace those definitions.
A private overlay may add scoped context and selection rules.

AOS owns reusable instructions, skills, and capability mappings. A target repo
owns ranked language and product declarations. Compose performs no fetches.

The resolver evaluates additive content from most specific to broadest: target
repos in request order, admitted private overlays in request order, then AOS
sources in request order. Source ids must be unique and content ids are
source-qualified. Precedence determines evaluation and deterministic
deduplication, not override authority. Byte-identical candidates for one
delivery key collapse to the highest-precedence candidate and record the rest
as `shadowed`. Non-identical candidates for one delivery key fail in v0.1.

## Composition flow

Agent-compose loads the embedded person source, validates the profile, reads
repo declarations, evaluates candidates, chooses delivery, and materializes a
bundle. The resolver emits trace nodes as each decision occurs. The bundle
protocol defines atomic materialization and cache reuse.

## Integration obligations

AOS publishes stable source ids, relative paths, and scope-safe content. AOS
does not carry Kai's person source.

Ward may supply facts and mount the bundle read-only. Ward treats the tree as
opaque and keeps authority and credentials outside the request.

## See also

* [kdl-contracts.md](kdl-contracts.md) - human-authored input grammar.
* [person-contract.md](person-contract.md) - embedded public-safe policy.
* [bundle-protocol.md](bundle-protocol.md) - immutable output contract.
* [manifest-schema.md](manifest-schema.md) - stable manifest fields.
* [decision-trace.md](decision-trace.md) - deterministic explanation data.
* [contract-review.md](contract-review.md) - cross-product decisions and gates.
