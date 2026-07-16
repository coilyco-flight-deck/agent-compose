# agent-compose

Agent-compose is a profile-aware compiler for agent context. It is being split
from [agentic-os](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os)
so context selection and delivery can evolve as a reusable product without
turning Ward into an information-architecture engine.

The intended product accepts explicit facts about the agent, role, model class,
privacy scope, and target repositories. It resolves those facts against
operator-owned policy, then emits an immutable context bundle for a host harness
or a warded container to consume.

## Ownership boundary

Agent-compose owns composition mechanics:

* profile and repo-capability resolution
* native-skill and compiled-context delivery
* immutable bundle materialization and caching
* harness load-point adapters and launch-time refresh
* bundle inspection, validation, and compatibility reporting

It does not own the knowledge itself. Agentic-os owns public doctrine, skills,
and capability policy. Private overlays remain outside this public repo. Ward
supplies runtime facts and mounts an opaque bundle. Infrastructure installs and
converges the resulting system.

## Status

The repository is in architecture and bootstrap. No compiler, bundle protocol,
or release artifact ships yet. The canonical Forgejo
[issue tracker](https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/issues)
holds the v0.1 implementation plan.

## Development

Development commands are declared in [`.ward/ward.yaml`](.ward/ward.yaml). The
current shell is documentation-only, so `ward exec test` runs the full
pre-commit suite. Implementation work must add language-specific build and test
verbs before invoking those tools.

## License

Agent-compose is available under the [MIT License](LICENSE).

## See also

* [AGENTS.md](AGENTS.md) - repo-specific operating rules.
* [docs/FEATURES.md](docs/FEATURES.md) - inventory of what exists today.
* [`.ward/ward.yaml`](.ward/ward.yaml) - allowlisted development commands.
* [Catalog trifecta convention](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/src/branch/main/docs/features-release-tooling.md) - shared entry-point structure.
