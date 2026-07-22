# Agent instructions

## Scope

Agent-compose is Kai's personal engine for selecting, compiling, and
materializing agent context for host harnesses and warded containers. It is
public source and may embed her public-safe company roster, personalities, and
composition defaults. Keep private identity detail, machines, credentials, and
deployment values out of the repo.

## Project shape

The repository is currently a documentation and tracker shell. The shipped
inventory is [`docs/FEATURES.md`](docs/FEATURES.md). The first implementation
must establish the package layout, KDL schema, and bundle contract together so
the code does not precede its public boundary.

## Repo boundaries

* `agent-compose` owns the compiler, schema, resolver, cache, bundle format,
  harness adapters, diagnostics, and Kai's public-safe person configuration.
* The person configuration owns organizational purpose, role-neutral
  personalities, curated compatibility, and context-selection policy.
* `agentic-os` owns reusable knowledge sources, general skills, capability
  providers, and editorial validators.
* Private overlays stay outside this public repo and may extend the embedded
  person source without replacing its canonical public-safe content.
* `ward` owns execution permissions, role authority, runtime-fact resolution,
  and the generic read-only bundle mount. Shared role slugs do not transfer
  permission ownership into agent-compose.
* `infrastructure` owns installation, binary shadowing rollout, host paths, and
  fleet convergence.
* Product repos are not an agent-compose concept. A repo may host personality
  files that a source locator references, and it owns any bespoke,
  foundational, or exceptional local skills.

## Commands

Route development through Ward using [`.ward/ward.yaml`](.ward/ward.yaml). The
current command surface is:

* `ward exec test` - run all repository validation.
* `ward exec pre-commit` - explicit spelling of the same pre-commit sweep.

Do not invoke a language tool directly. Add the implementation language and its
build, test, lint, and install verbs to `.ward/ward.yaml` before using them.

## Validation

Run `ward exec test` before every commit. The agentic-os hook catalog enforces
the documentation trifecta, flat docs, cross-links, public-safe prose, comment
discipline, and secret scanning. Add focused engine tests with each executable
slice once implementation begins.

## Safety

Context is not a credential transport. Bundles may contain instructions,
skills, and declarative routing metadata, but never tokens, auth stores, opaque
host identifiers, or mutable harness state. Materialized bundles are immutable
and mounted read-only. A failed refresh must not partially replace a known-good
bundle.

## Cross-repo contracts

The input contract names the requested role, personality, density, delivery,
and sources without importing Ward policy. The output contract is a manifest
plus a filesystem tree that consumers treat as opaque. Agent-compose combines its embedded person policy with AOS
capability providers and scoped overlays. Ward and native harness wrappers
consume the output contract. Changes to either contract need compatibility
tests against both consumers before release.

## Release

Canonical development and issues live on Forgejo. The project is pre-v0.1 and
has no release pipeline yet. The release slice must define versioning, binary
distribution, and protocol compatibility before publishing the first artifact.
Do not infer release behavior from agentic-os or Ward.

## Agent rules

Keep one issue per independently verifiable vertical slice. Kai-specific,
public-safe role and personality policy is first-class product content. Do not
generalize it prematurely, copy reusable AOS skills into this repo, or allow
personality to alter truthfulness, authority, safety, rollback, or completion.
Generated bundles and rendered references stay uncommitted. Update
[`docs/FEATURES.md`](docs/FEATURES.md) only when a significant capability
actually ships.

## See also

* [README.md](README.md) - human-facing product boundary and status.
* [docs/FEATURES.md](docs/FEATURES.md) - current shipped inventory.
* [`.ward/ward.yaml`](.ward/ward.yaml) - allowlisted development commands.
* [Catalog trifecta convention](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/src/branch/main/docs/features-release-tooling.md) - shared entry-point structure.
