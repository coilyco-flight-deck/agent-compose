# Agent instructions

## Scope

Agent-compose is the generic engine that selects, compiles, and materializes
agent context for host harnesses and warded containers. Keep this repository
public-safe and independent of any one operator's doctrine, identity, machines,
or deployment.

## Project shape

The repository is currently a documentation and tracker shell. The shipped
inventory is [`docs/FEATURES.md`](docs/FEATURES.md). The first implementation
must establish the package layout, KDL schema, and bundle contract together so
the code does not precede its public boundary.

## Repo boundaries

* `agent-compose` owns the generic compiler, schema, resolver, cache, bundle
  format, harness adapters, and diagnostics.
* `agentic-os` owns public knowledge sources, skills, capability vocabulary,
  provider policy, and editorial validators.
* Private overlays and operator-specific policy do not enter this repo.
* `ward` owns execution permissions, role authority, runtime-fact resolution,
  and the generic read-only bundle mount. It must not interpret composition
  policy.
* `infrastructure` owns installation, binary shadowing rollout, host paths, and
  fleet convergence.
* Product repos own their identity declarations and any bespoke, foundational,
  or exceptional local skills.

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

The input contract names resolved facts without importing Ward policy. The
output contract is a versioned manifest plus a filesystem tree that consumers
treat as opaque. AOS policy maps repo capabilities to content. Ward and native
harness wrappers consume the output contract. Changes to either contract need
compatibility tests against both consumers before release.

## Release

Canonical development and issues live on Forgejo. The project is pre-v0.1 and
has no release pipeline yet. The release slice must define versioning, binary
distribution, and protocol compatibility before publishing the first artifact.
Do not infer release behavior from agentic-os or Ward.

## Agent rules

Keep one issue per independently verifiable vertical slice. Do not copy AOS
content into engine fixtures or teach the engine Kai-specific profile names.
Generated bundles and rendered references stay uncommitted. Update
[`docs/FEATURES.md`](docs/FEATURES.md) only when a significant capability
actually ships.

## See also

* [README.md](README.md) - human-facing product boundary and status.
* [docs/FEATURES.md](docs/FEATURES.md) - current shipped inventory.
* [`.ward/ward.yaml`](.ward/ward.yaml) - allowlisted development commands.
* [Catalog trifecta convention](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/src/branch/main/docs/features-release-tooling.md) - shared entry-point structure.
