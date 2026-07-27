# agent-compose

Agent-compose is the context substrate between knowledge providers and native
or isolated agent harnesses. It selects, compiles, and installs the context an
agent starts with while keeping executable authority outside the bundle. The
public product is intentionally opinionated: it includes Kai's public-safe
operating roster, personality catalog, and composition defaults. An independent
deployment can select one complete external person package instead, with no
role or definition inheritance from that default.

The intended product accepts a role, a model class, a delivery mode, and
optional capability sources. The role activates its complete ordered
personality set.
Agent-compose resolves every bound personality skill from the selected person
package, tells the agent every component and melded favorite color, injects
compact selected-role metadata, and emits an immutable context bundle.

## Ownership boundary

Agent-compose owns the context boundary and its bundled public-safe person
configuration:

* role-driven personality meld resolution
* one selected person package, with Kai's twelve-role roster as the default
* external person packages that fully replace the default roster
* role-neutral personality catalog bindings, definitions, invariant, and
  curated compatibility
* complete selected-role terminal metadata, including identity primitives,
  seats, inspiration records, appearances, and renderer expressions
* credited role and personality inspirations with sourced public appearances
* ordinary and role-composed skill selection
* model-class-aware pruning controlled by each skill's own frontmatter
* fail-closed per-role model-class compatibility
* native-skill and compiled-context delivery with source entry-point promotion
* immutable bundle materialization and caching
* harness load-point adapters and launch-time refresh
* host doctrine convergence, native skill installation, and native MCP projection
* bundle inspection, validation, and compatibility reporting
* compact text and JSON identity overlays with caller-supplied state
* deterministic Markdown and YAML behavior evaluation packs with YAML scored baselines

Knowledge providers own reusable doctrine, general skills, capability sources,
and editorial validation. Agent-compose combines those sources with its
selected person provider to build the concrete context surface for each
harness. External person packages and private overlays remain outside this
public repo. Launch consumers own executable authority, runtime facts, mounts,
and lifecycle. Personality and
organizational framing never alter consumer permissions. Infrastructure
installs the resulting system across hosts.

## Personal by design

The initial person configuration is Kai's twelve-role operating roster and
sixteen-personality catalog described in
[agentic-os#602](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/issues/602)
and tracked for implementation in
[agent-compose#10](https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/issues/10).
The engine keeps this package as its opinionated zero-config default. Another
deployment can select a complete package using the same validated layout. The
selection is exclusive: an external package contributes its own roles, seats,
personality definitions, and evaluation context without loading `person:kai`.
An `external-only` policy makes that boundary fail closed across the machine.
See [person packages](docs/person-packages.md).

## Status

Current releases ship the Go composition engine, verified deterministic
bundles, transactional repo and container-home projection, decision
inspection, refresh-then-exec, the absorbed AOS cascade, host convergence, and
package-manager distribution. A configured mcporter inventory projects into
both Claude Code and Codex native MCP registries during host convergence.
Provider roots contribute ordinary skills for
every role plus `COMPOSED.md` sources selected by `.agents/roles.kdl`. The
default `person:kai` provider supplies the personality invariant and all 16
canonical definitions, so its host roster convergence needs no external
personality source. A configured external person package replaces that provider
as one unit. The selected person policy carries a normalized catalogue of
credited inspirations and representative public appearances without adding
biography to runtime prompts. The repository also ships a local personality
palette explorer, identity overlay, and four-case behavior evaluation pack.
Bare `acompose` converges the host and atomically refreshes the selected
versioned person snapshot at
`~/.agent-compose/sources/personality/person.json`.
`acompose -- <command>` converges context before launching the command.
`acompose --reapply` rewrites the generated compose outputs and recreates
global load-point links even when they are current. `acompose --verbose`
prints every composition source and load-point file as
`source => destination`.

Composition adapters can project a verified bundle into an empty staged home,
remove agent-compose's projection state, and wrap the remaining selected load
points in their own schema. See [staged-home.md](docs/staged-home.md).

## Install

Via Homebrew (macOS and Linux):

```sh
brew tap coilyco-flight-deck/tap https://forgejo.coilysiren.me/coilyco-flight-deck/homebrew-tap.git
brew install coilyco-flight-deck/tap/agent-compose
```

Via Scoop (Windows):

```sh
scoop bucket add coilyco https://forgejo.coilysiren.me/coilyco-flight-deck/scoop-bucket.git
scoop install coilyco/agent-compose
```

Both managers also install `acompose`, which is the compose verb directly:
bare for host convergence, `acompose -- <command>` for refresh-then-exec.

Release binaries (darwin-arm64, linux-amd64/arm64, windows-amd64) also attach
to tagged
[Forgejo releases](https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/releases)
directly. From source, `ward exec install` builds into GOBIN.
`agent-compose version` reports the build you are running.
Every push to canonical `main` validates and publishes the next minor release.

## Development

Development commands are declared in [`.ward/ward.yaml`](.ward/ward.yaml).
`ward exec test` runs the Go and palette tests plus the full pre-commit sweep.
`ward exec smoke` builds the real `acompose` entry point and converges an
isolated temporary home twice, covering roster, cascade, skills, MCP projection,
and idempotence without touching live host state or the network. It reports each
stage. `ward exec smoke-verbose` also prints both captured convergence transcripts.
`build`, `lint`, `install`, and `tidy` cover the remaining Go verbs.

`ward exec palette-serve` generates browser data from the embedded person
source and starts the local explorer. `palette-build`, `palette-test`, and
`palette-tidy` cover its remaining development lifecycle. See
[the personality palette walkthrough](docs/personality-palette.md).

## License

Agent-compose is available under the [MIT License](LICENSE).

## See also

* [AGENTS.md](AGENTS.md) - repo-specific operating rules.
* [docs/FEATURES.md](docs/FEATURES.md) - inventory of what exists today.
* [docs/architecture.md](docs/architecture.md) - shipped composition boundary.
* [docs/staged-home.md](docs/staged-home.md) - provider-neutral adapter handoff.
* [docs/person-packages.md](docs/person-packages.md) - independent roster and evaluation packages.
* [docs/inspiration-catalogue.md](docs/inspiration-catalogue.md) - credited influence and provenance contract.
* [docs/evaluation.md](docs/evaluation.md) - four-case human behavior review.
* [docs/personality-palette.md](docs/personality-palette.md) - local color explorer.
* [docs/release.md](docs/release.md) - automatic Forgejo release pipeline.
* [`.ward/ward.yaml`](.ward/ward.yaml) - allowlisted development commands.
* [Catalog trifecta convention](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/src/branch/main/docs/features-release-tooling.md) - shared entry-point structure.
