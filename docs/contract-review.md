# v0.1 contract review

This is the human review record for issue #2. The contract remains proposed
until Kai confirms or changes the review gates below. Implementation issue #3
must consume the reviewed contract rather than silently revising it.

## Proposed decisions

* Cli-guard owns the policy-free claim structure. Agent-compose accepts exactly
  one `context` role and rejects authority roles.
* Agent-compose owns personality, mode, privacy scope, target repositories,
  source selection, and delivery. Permissions and task acceptance stay with the
  consumer and never enter the context bundle as grants.
* The binary embeds one canonical public-safe person source. External sources
  can add scoped context but cannot replace its role, personality,
  compatibility, or invariant definitions.
* Additive content precedence is target repositories, private overlays, then
  AOS sources. Request order breaks ties within a class. Identical delivery
  candidates deduplicate, while non-identical collisions fail closed in v0.1.
* Every bundle retains canonical selected instructions and skill trees.
  Compiled profiles add a derived context document rather than discarding the
  canonical inputs.
* Agent-compose owns a fixed v0.1 adapter registry. Claude and Codex use native
  skills. Goose and OpenCode use compiled context. Model class changes density,
  not delivery mode.
* The manifest records the embedded person, target repo, admitted overlay, and
  AOS source digests. It inventories every payload file and names only generic
  entry points.
* The trace records normalized decision inputs and metadata as resolution
  occurs. Volatile runtime telemetry remains outside the bundle.

## Ward integration record

Ward may supply the versioned compose request and mount the resulting bundle
read-only. Ward validates the manifest protocol, version, file digests, and
declared entry points before launch. It treats source records, selected content,
and the decision trace as opaque.

Ward keeps authority claims, guardfiles, credentials, mutable harness state,
and task acceptance outside the compose request. A shared role slug is only a
join key between independently resolved context and authority. It grants
nothing by itself. Ward can run without agent-compose, and agent-compose can run
without Ward.

## AOS integration record

AOS publishes stable source ids, source-relative declaration paths, reusable
content, and capability-to-skill mappings. It does not publish Kai's person
source, harness load points, Ward policy, or installation paths.

Agent-compose resolves AOS declarations locally without fetching. The request
supplies each locator, while the bundle and trace retain only normalized ids,
relative paths, and digests. Product repos declare ranked identity without
provider paths or selected skill names.

## Compatibility fixtures

* `native-claude.kdl` - Claude with native skills.
* `native-codex.kdl` - Codex with native skills.
* `compiled-goose.kdl` - Goose with compiled context on a frontier model.
* `compiled-qwen.kdl` - Qwen through OpenCode with compiled context on a local
  model.

The Goose and Qwen pair proves that harness capability chooses delivery while
model class independently chooses context density.

## Review gates

Kai needs to confirm or change three defaults before issue #2 lands:

* Compiled bundles retain canonical selected skill trees beside the derived
  context document.
* V0.1 fixes delivery to Claude/Codex native and Goose/OpenCode compiled.
* V0.1 rejects non-identical delivery-key collisions instead of adding an
  override grammar.

Everything else in the contract is an implementation constraint rather than an
open product fork.
