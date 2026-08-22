# Release and the v2 migration

How agent-compose releases, and what the v2 migration changed.

## Release

Agent Compose releases are Forgejo-canonical. Every push to `main` enters a
no-cancel queue and validates the exact commit. The owning
`scripts/release-impact.sh` classifier then decides whether publication runs.

Automatic publication occurs when the unreleased diff from the latest reachable
`v*` release tag changes shipped product inputs:

* the Go command or internal engine and embedded Core Roster
* Go module dependencies
* release binary construction
* Homebrew or Scoop rendering

Documentation, scored evaluation results, examples, tests, and development
workflow changes still validate but do not create a product version. The
event base is the fallback before the first release tag. This roll-forward
window keeps a failed product release eligible when a later main push contains
only recovery evidence. The classifier fails closed to publication when its
base revision is unavailable. Its fixture suite covers documentation, results,
product code, roll-forward recovery, initial pushes, the major hold, and manual
dispatch.

An automatic release bumps the minor version, cross-compiles macOS, Linux, and
Windows binaries, creates the Forgejo release, uploads checksums and package
files, and updates Homebrew and Scoop when their write tokens are present.

### Major release hold

A tracked `.release-major` pauses automatic publication while a breaking stack
lands. Main validation continues. Manual workflow dispatch ignores the hold and
remains the only major-version and explicit-tag path.

For v2.0, keep the hold until the Core Roster stack, complete frontier and OSS
evidence, independent QA verdict, generated scorecard, and clean-main smoke are
all present. Then an operator dispatches the exact landed revision with a major
bump. Remove the hold in a follow-up commit only after v2.0.0 and both package
channels are verified. Removing the hold alone does not publish another
version.

## Agent Compose v3 roster migration

v3 takes the Core Roster from nine seats to seven and renames every slug. There
are no compatibility aliases either way, so the whole set moves in one break
rather than drifting through two: `ai` and `engineer` named a technology and a
generic category while every other slug named a practice.

### Role destinations

* `engineer` becomes `platform`, displayed as Developer Platform Engineer.
* `ops` becomes `sysadmin`, displayed as Systems Administrator.
* `ai` becomes `eval`, displayed as Agent Evaluation Engineer.
* `design` becomes `frontend`, displayed as Frontend Design Engineer.
* `creator` becomes `devrel`, displayed as Developer Advocate.
* `gamedev` is unchanged.
* `director` and `exec` merge into `tpm`, displayed as Technical Program
  Manager, on the finding that they split on seniority rather than practice.
* `qa` is cut. Its done-condition folds into `eval`, and code review folds into
  `tpm` as a gate decision.

Three boundaries rename with it: `build-software` to
`build-foundational-software`, `modify-live-system` to `modify-live-backend`,
and `suggest-human-comms` to `suggest-external-comms`. `modify-live-backend` is
re-ownered to sysadmin and `seek-external-validation` to tpm. Every boundary now
reaches all seven seats as one owner, two scoped, and four deferring.

Melds drop from three traits to two, one signature plus one bond shared with a
sibling seat, and the trait vocabulary falls from 19 to 10. The nine retired
bodies are preserved verbatim at
[#316](https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/issues/316).

Update launch commands, Ward role selections, composed-skill bindings,
`AGENT_GIT_ATTRIBUTION_ROLE`, per-role Claude themes, and evaluation inputs to
the new identifiers. Scored records from earlier rosters remain historical
evidence and retain their original role identities.

## Agent Compose v2 roster migration

Agent Compose v2 renames the baked provider from `person:kai` to
`roster:core` and emits the Core Roster slugs, including AI Engineer as `ai`.
This is an intentional major-version break. There are no compatibility aliases
for old role identifiers. The Core Roster returned to eight roles at that point, after
audience-facing work proved more coherent as one Content Creator feedback loop.
Game Developer joined later as the ninth.

### Role destinations

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

### External packages

External `person "<name>"` packages remain supported through the validated
package contract and keep `person:<name>` provenance. A package may adopt a
`roster "<name>"` manifest when it intends to emit `roster:<name>` provenance.
The package selector remains exclusive and never inherits Core Roster roles or
definitions.
