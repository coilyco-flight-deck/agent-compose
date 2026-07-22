# agent-compose

Agent-compose is Kai's personal, profile-aware compiler for agent context. It
is public source, but intentionally opinionated: the product includes her
synthetic company, personality catalog, and composition defaults rather than
presenting itself as a neutral enterprise framework.

The intended product accepts a role, a personality, a context density, a
delivery mode, and the locations of personality sources. It resolves those
inputs against embedded personal policy plus scoped overlays, then emits an
immutable context bundle for a host harness or a warded container to consume.

## Ownership boundary

Agent-compose owns composition mechanics and its bundled public-safe person
configuration:

* role and personality resolution
* the ten-seat company roster and organizational purpose
* role-neutral personalities and curated role compatibility
* native-skill and compiled-context delivery
* immutable bundle materialization and caching
* harness load-point adapters and launch-time refresh
* bundle inspection, validation, and compatibility reporting

Agentic-os owns reusable doctrine, general skills, capability providers, and
editorial validation. Private overlays remain outside this public repo. Ward
owns executable authority and supplies runtime facts while mounting an opaque
bundle. Personality and organizational framing never alter Ward permissions.
Infrastructure installs and converges the resulting system.

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

The first executable slice ships: a Go CLI composes a public fixture profile
through the embedded person source into a deterministic, atomically written
bundle with a decision trace. The full personality roster (#10), Ward
consumption (#7), and the release pipeline remain open on the canonical Forgejo
[issue tracker](https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/issues).

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

Release binaries (darwin-arm64, linux-amd64/arm64, windows-amd64) also attach
to tagged
[Forgejo releases](https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/releases)
directly. From source, `ward exec install` builds into GOBIN.
`agent-compose version` reports the build you are running.

## Development

Development commands are declared in [`.ward/ward.yaml`](.ward/ward.yaml).
`ward exec test` runs the Go test suite and the full pre-commit sweep;
`build`, `lint`, `install`, and `tidy` cover the remaining Go verbs.

## License

Agent-compose is available under the [MIT License](LICENSE).

## See also

* [AGENTS.md](AGENTS.md) - repo-specific operating rules.
* [docs/FEATURES.md](docs/FEATURES.md) - inventory of what exists today.
* [docs/architecture.md](docs/architecture.md) - proposed v0.1 composition boundary.
* [`.ward/ward.yaml`](.ward/ward.yaml) - allowlisted development commands.
* [Catalog trifecta convention](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/src/branch/main/docs/features-release-tooling.md) - shared entry-point structure.
