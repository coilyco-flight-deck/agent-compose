# agent-compose

Agent-compose is the context substrate between AOS knowledge and Ward or native
agent harnesses. It selects, compiles, and installs the context an agent starts
with while keeping executable authority outside the bundle. The public product
is intentionally opinionated: it also includes Kai's synthetic company,
personality catalog, and composition defaults rather than presenting itself as
a neutral enterprise framework.

The intended product accepts a role, a personality, a context density, a
delivery mode, and the locations of personality sources. It resolves those
inputs against embedded personal policy plus scoped overlays, then emits an
immutable context bundle for a host harness or a warded container to consume.

## Ownership boundary

Agent-compose owns the context boundary and its bundled public-safe person
configuration:

* role and personality resolution
* the ten-seat company roster and organizational purpose
* role-neutral personalities and curated role compatibility
* native-skill and compiled-context delivery
* immutable bundle materialization and caching
* harness load-point adapters and launch-time refresh
* host doctrine convergence and native skill installation
* bundle inspection, validation, and compatibility reporting

AOS owns reusable doctrine, general skills, capability providers, and editorial
validation. Agent-compose turns those sources into the concrete context surface
for each harness. Private overlays remain outside this public repo. Ward owns
executable authority, supplies runtime facts, and mounts opaque bundles.
Personality and organizational framing never alter Ward permissions.
Infrastructure installs the resulting system across hosts.

## Personal by design

The initial person configuration is Kai's ten-role synthetic company and
sixteen-personality catalog described in
[agentic-os#602](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/issues/602)
and tracked for implementation in
[agent-compose#10](https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/issues/10).
The engine may keep clean internal contracts, but v0.1 does not add abstraction
solely to make this personal configuration look generic. Other users can fork
or replace policy later without becoming the design center now.

## Status

Current releases ship the Go composition engine, verified deterministic
bundles, transactional repo and container-home projection, decision
inspection, refresh-then-exec, the absorbed AOS cascade, host convergence, and
package-manager distribution. Bare `acompose` converges the host.
`acompose -- <command>` converges context before launching the command.

The full personality roster
([#10](https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/issues/10))
and Ward-side container invocation
([#17](https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/issues/17))
remain open on the canonical Forgejo tracker.

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
`ward exec test` runs the Go test suite and the full pre-commit sweep;
`build`, `lint`, `install`, and `tidy` cover the remaining Go verbs.

## License

Agent-compose is available under the [MIT License](LICENSE).

## See also

* [AGENTS.md](AGENTS.md) - repo-specific operating rules.
* [docs/FEATURES.md](docs/FEATURES.md) - inventory of what exists today.
* [docs/architecture.md](docs/architecture.md) - shipped composition boundary.
* [docs/release.md](docs/release.md) - automatic Forgejo release pipeline.
* [`.ward/ward.yaml`](.ward/ward.yaml) - allowlisted development commands.
* [Catalog trifecta convention](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/src/branch/main/docs/features-release-tooling.md) - shared entry-point structure.
