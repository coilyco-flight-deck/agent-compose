# agent-compose

Eval driven composer for roles and personas

![agent-compose and $ acompose, eval driven composer for roles and personas](assets/banner/agent-compose-banner.jpg)

agent-compose compiles the context an agent harness loads. It selects a role,
the personality meld that role carries, the skills that role can see, and the
tool inventory it gets, then materializes one immutable bundle of plain files.
Claude Code, Codex, Goose, and OpenCode all take the same bundle.

**A role is context, never permission.** The bundle carries no credential, no
mount, and no command, and a role slug shared with a launch consumer transfers
no authority into agent-compose. Execution permissions, runtime facts, and
lifecycle stay with whatever launches the agent. You can read and diff the whole
bundle before a run, and `verify` will tell you it is complete before you do.

## Why eval driven

The board is derived from the roster rather than written beside it.
housecast's `evalkit.matrix` reads the roster and prints the cases it implies: boundaries
and their owners produce the pairs, adjacency produces the role-fit targets, and
each role's meld produces the personality cases. Add a boundary, flip an
adjacency edge, or swap a personality and the challenge list moves on its own.
You cannot write a case that does not correspond to the roster, and you cannot
change the roster without changing what gets tested.

The hard cases are generated on purpose. Role adjacency names each role's two
likeliest absorptions, and those reasons become the descriptors a generator uses
to build exactly the confusion a seat is most at risk of.

Three parties and none of them holds two seats: a generator authors the cases, a
subject answers them through Agent Proxy, and a human grades them. The grading
half ships separately as `aos-eval` in agentic-os, so it holds no runner and no
model client, and grading never spends a token or touches a deployed system.
Details in [docs/evaluation.md](docs/evaluation.md).

## Install

```sh
brew tap coilyco-flight-deck/tap https://forgejo.coilysiren.me/coilyco-flight-deck/homebrew-tap
brew install coilyco-flight-deck/tap/agent-compose
```

```powershell
scoop bucket add coilyco-flight-deck https://forgejo.coilysiren.me/coilyco-flight-deck/scoop-bucket
scoop install coilyco-flight-deck/agent-compose
```

Both also install `acompose`, the compose verb directly. Tagged
[Forgejo releases](https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/releases)
attach darwin-arm64, linux-amd64, linux-arm64, and windows-amd64 binaries. From
source, `just install` builds into `GOBIN`. `agent-compose version` reports the
build you are running.

## Use it

```sh
acompose                              # converge the host from the local catalogue
acompose -- <command>                 # refresh context, then exec the command
agent-compose launch <role> <harness> # launch one harness with an assigned role
```

`--reapply` rewrites the host compose layout even when it is already current,
and `--verbose` prints every `source => destination` mapping it places.
`--explain` adds the role briefing, the credits, the expression vocabulary,
and the full decision tree to the summary.

When you want to see what a bundle actually decided rather than trust it,
`describe` renders its stored decision tree, `diff` reports the semantic change
between two bundles, and `verify` checks one is complete and safe to consume.
Composition adapters can project a verified bundle into an empty staged home
and wrap it in their own schema. See
[docs/staged-home.md](docs/staged-home.md).

## The Core Roster

The zero-config default, `roster:core`, ships seven seats: Agentic Platform
Engineer, Systems Administrator, Applied Scientist, Design Engineer, Game
Developer, Portfolio Director, and Developer Advocate. Each melds one signature
trait with one bond it shares with a sibling seat.

Selection is exclusive. An external person package contributes its own roles,
seats, personality definitions, and evaluation context, and it replaces
`roster:core` wholesale rather than merging with it. An `external-only` policy
makes that boundary fail closed across the machine. See
[docs/person-packages.md](docs/person-packages.md).

## Development

Dev verbs are recipes in the [justfile](justfile). `just test` runs the Go tests
plus the full hook sweep. `just smoke` builds the real `acompose` entry point and
converges an isolated temporary home twice, covering roster, cascade, skills,
load points, and idempotence without touching live host state or the network.
`just palette-serve` starts the local personality palette explorer.

Every push to canonical `main` validates and publishes the next minor release.
Forgejo is canonical and the GitHub mirror is verified.

## License

MIT. See [LICENSE](LICENSE).

## See also

- [AGENTS.md](AGENTS.md) - agent-facing operating rules.
- [docs/FEATURES.md](docs/FEATURES.md) - inventory of what ships today.
- [docs/architecture.md](docs/architecture.md) - the composition boundary.
- [docs/ownership.md](docs/ownership.md) - who owns which boundary, and why.
- [docs/role-selection.md](docs/role-selection.md) - role-scoped providers and assigned bundles.
- [docs/personality.md](docs/personality.md) - the personality catalog and palette.
- [docs/release.md](docs/release.md) - the automatic Forgejo release pipeline.
- [justfile](justfile) - development recipes.
- [.ward/ward.yaml](.ward/ward.yaml) - catalog metadata only.
