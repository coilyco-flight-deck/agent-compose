# The bundle protocol and contract review

How a bundle is assembled, and how its contract is reviewed. Every successful
composition produces one immutable tree, and consumers enter it through
`manifest.json` and otherwise treat the tree as opaque.

## Bundle protocol

* `manifest.json` - what was composed, the delivery entry points, and
  `delivery.body_bytes`, the size of what that mode hands a consumer.
* `trace.json` - decisions, provider outcomes, and context-budget contributions.
* `content/instructions.md` - selected instructions and one role identity card.
* `content/skills/<source-id>/<skill>/...` - canonical selected skill trees.
* `delivery/compiled.md` - only when the adapter compiles selected skill bodies
  into one document. Canonical skill trees stay beside it.

Every path uses slash-separated relative form, trees contain regular files and
directories only, symlinks and escaping paths are invalid, and harness load-point
paths never appear inside the generic tree. A source id is percent-encoded per
path segment while `manifest.json` and `trace.json` keep the raw id, so
`roster:core` is `roster%3Acore` on disk.

### Immutability and atomicity

The materializer stages beside the final location, verifies the tree is complete,
then renames it into place atomically. A bundle is never rewritten in place:
refresh builds a new tree and swaps it in, and a failed refresh leaves the
previous bundle live rather than partially replacing it.

`agent-compose verify <bundle-dir>` is that same read-only check, and cache hits
re-verify before reuse, so `manifest.json` alone never blesses a tree. It prints
bounded counts only, leaving identity detail in `trace.json` for `describe`.
`bundle export` verifies first, then writes sorted names with normalized gzip and
tar metadata, so identical trees produce byte-identical archives. Runtime
telemetry never lands under the bundle root.

### One identity card per bundle

The operating base leads `content/instructions.md`, and it is also the host global
load point, where it carries every role's identity card. A native session launched
with no assigned role self-selects from that table, so the host global keeps it.

A bundle assigns a role and its own instructions refuse to switch, so the other
cards are unreachable. The assembler drops every card from the base and renders
the assigned one back. That single copy is authoritative: it carries the resolved
melded color and boundary list, which the roster's copy does not, so keeping both
would show two different cards for one role. Measured against the shipped roster,
the strip removes 40,468 bytes.

The strip runs from a card's heading to the next top-level heading. No role ships
a copy contract today, and one that starts to needs it carried to the re-render.

### Producer contract

`testdata/handmade-bundle` is hand-authored, `agent-compose` never composed it,
and every consumer surface accepts it. Beyond the path rules above, `verify` wants
a `content/skills` tree equal to the skill set its trace selects, and one selected
`profile` decision per role, personality, and boundary the manifest names with no
others. `providers` is optional and exact once present: `context_bytes` must equal
the byte sum of that source's selected skill trees, and `approximate_tokens` must
equal `(context_bytes + 3) / 4`. It is the only field a producer cannot guess.

Four fields a producer owes are unenforced, each confirmed by mutation.

* `content[].digest` is checked for shape, never recomputed against bytes.
* `role_skill_digest` is checked for presence only, not even for shape.
* `delivery.body_bytes` is never read.
* `identity` is optional, and omitting it degrades renderers to the role slug.

## v0.1 contract review

The human review record for issue #2, reviewed in issue #13. Implementation
consumes these decisions, not the earlier proposal.

### Review decisions

* Agent-compose is a personality engine. It owns personality, source
  selection, and delivery. It is not a security boundary.
* Repositories are not an agent-compose concept. A repo is at best a place
  capability files happen to live, reached through a source locator like any
  other directory. Privacy scopes, target repositories, repo declarations, and
  per-repo capability resolution are removed from the contract.
* Agent, model, harness, reasoning effort, and interactivity belong to the
  caller and launcher and never enter a compose request. Legacy `density "full"`
  is ignored and brief density is rejected (#59).
* Delivery mode, native skills or compiled context, is load-bearing and stays.
* A compose request selects a role, not one personality. The role activates
  every personality in its ordered set, and their component colors derive one
  melded favorite for the bundle.
* Personality definitions live inside `SKILL.md` trees, bound by name. Their
  bodies are freeform prose, like role and boundary bodies.
* The invariant is shared instruction prose, not schema surface.
* No schema-version fields and no digest ceremony. Immutability and atomic
  refresh remain, and a bundle is not a trust artifact.
* The decision trace stays, as a plain ordered list with readable reasons.
* Byte-identical duplicate content deduplicates; non-identical collisions for
  one delivery slot still fail in v0.1 instead of adding an override grammar.
* Agent identity is named seats: `agent` nodes under each role. Names are opaque
  to the engine, and launchers keep permissions, models, and effort.

### Consumer integration record

A consumer may build the compose request and adapt the resulting bundle or home
projection while treating the source tree as immutable. Authority claims,
credentials, permissions, mutable harness state, and task acceptance stay with
that consumer, and either product can run independently.

### Knowledge-provider integration record

Knowledge providers publish reusable ordinary skills and instructions under
stable source ids and relative paths. They do not publish Kai's person source,
personality definitions, harness load points, launch policy, or installation
paths, and agent-compose resolves local declarations without fetching.

### Compatibility fixtures

`native.kdl` selects instructions plus native skill trees and `compiled.kdl`
selects instructions and skill bodies in one document, so the two together prove
delivery mode varies without agent-compose knowing which harness or model sits
behind it.
