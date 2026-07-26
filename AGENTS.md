# Agent instructions

## Scope

Agent-compose is the context substrate between knowledge providers and native
or isolated harness consumers. It selects, compiles, and materializes agent
context for hosts and staged homes. It is public source and embeds Kai's public-safe
company roster, personalities, and composition defaults. Keep private identity
detail, machines, credentials, and deployment values out of the repo.

## Project shape

The repository ships a Go engine under `cmd/agent-compose` and `internal/`,
with contract fixtures in `testdata/contracts`. The shipped inventory is
[`docs/FEATURES.md`](docs/FEATURES.md). The KDL schema, bundle format, and
package layout land together so the code never precedes its public boundary.

## Repo boundaries

* `agent-compose` owns the compiler, schema, resolver, cache, bundle format,
  harness adapters, diagnostics, and Kai's public-safe person configuration.
* The person configuration owns organizational purpose, personality catalog
  bindings, curated compatibility, credited inspirations, and
  context-selection policy.
* `agent-compose` owns the personality invariant and canonical personality
  definitions alongside the person configuration that binds them.
* `agentic-os` owns reusable knowledge sources, general skills, capability
  providers, and editorial validators.
* Private overlays stay outside this public repo and may extend the embedded
  person source without replacing its canonical public-safe content.
* Launch consumers own execution permissions, role authority, runtime-fact
  resolution, and their handoff schemas. Shared role slugs do not transfer
  permission ownership into agent-compose.
* `infrastructure` owns installation, binary shadowing rollout, host paths, and
  fleet convergence.
* Product repos are not an agent-compose concept. A repo may host capability
  files that a source locator references, and it owns any bespoke,
  foundational, or exceptional local skills.

## Commands

Route development through Ward using [`.ward/ward.yaml`](.ward/ward.yaml). The
current command surface is:

* `ward exec smoke` / `smoke-verbose` - test isolated host convergence.
* `ward exec test` - run all repository validation (Go tests plus hooks).
* `ward exec build` / `fmt` / `lint` / `install` / `tidy` - Go engine verbs.
* `ward exec pre-commit` - explicit spelling of the hook sweep alone.

Do not invoke a language tool that has no verb here. Add new build, test,
lint, and install verbs to `.ward/ward.yaml` before using them.

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

The input contract names the requested role, delivery, and sources
without importing launcher policy. The role activates its complete ordered
personality set. The output contract is a manifest plus a filesystem tree that
consumers treat as opaque. Agent-compose combines its embedded person policy
with capability providers and scoped overlays. Native harness wrappers and
staged-home adapters consume the output contract. Changes to either contract
need compatibility tests against both paths before release.

## Release

Canonical development, releases, and issues live on Forgejo. Every push to
`main` queues the single-stage release workflow, validates the commit, bumps the
minor version, and publishes version-stamped cross-platform binaries plus
Homebrew and Scoop metadata. Manual dispatch may select patch or major instead.
Protocol compatibility is deliberately thin: consumers check the manifest
`format` marker and nothing else.

## Agent rules

Keep one issue per independently verifiable vertical slice. Kai-specific,
public-safe role, personality, and inspiration policy is first-class product
content. Do not generalize it prematurely, copy ordinary AOS skills into this
repo, or allow personality to alter truthfulness, authority, safety, rollback,
or completion.
Generated bundles and rendered references stay uncommitted. Update
[`docs/FEATURES.md`](docs/FEATURES.md) only when a significant capability
actually ships.

## See also

* [README.md](README.md) - human-facing product boundary and status.
* [docs/FEATURES.md](docs/FEATURES.md) - current shipped inventory.
* [`.ward/ward.yaml`](.ward/ward.yaml) - allowlisted development commands.
* [Catalog trifecta convention](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/src/branch/main/docs/features-release-tooling.md) - shared entry-point structure.
