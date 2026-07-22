# V1 integration and delivery tiers

Two composers share the agent-compose name. V1 is the Python generator in
agentic-os (`generate-agent-compose`): it composes `AGENTS.COMPOSE.md`
doctrine sources into `~/.config/agent-compose/COMPOSED.<harness>.md`,
symlinks each harness's global load point at the result (claude
`~/.claude/CLAUDE.md`, codex `~/.codex/AGENTS.md`, opencode
`~/.config/opencode/AGENTS.md`), applies scope and harness filtering with
per-harness section overrides, and emits the mount-eligibility manifest ward
reads. V2 is this repo: the personality engine - bundles, projection,
launch, inspection.

They coexist by ownership, not by flag-day. V1 keeps global doctrine and the
mount manifest. V2 owns personality. Binary names differ, nothing breaks by
installing both, and absorbing v1 into v2 is explicitly out of v0.1 scope -
whoever absorbs it later inherits the mount-eligibility obligation to ward.

## The seam rule

On a host, v1 owns the harness load points and v2 feeds it sources. In a
container, v2 owns the whole home and v1 does not exist. No path is ever
written by both.

## Host tier: personality rides the v1 cascade

V2 renders a roster artifact into `~/.config/agent-compose/sources/` - a
directory v1 already walks as a source root - containing an
`AGENTS.COMPOSE.md` entry plus the personality files it references. The entry
carries the seat dispatch table: "if you are codex running the engineer
role: your name is terran engineer (he/him), your personality is grounded,
its definition lives at <path>, your favorite color is #5fa87a." Each agent
self-selects by facts it already knows; no launcher cooperation, no
environment variable, no blessed entrypoint.

The next v1 compose run carries the table into every harness's global load
point with zero v1 code changes. Global context loads at session start
unconditionally, which is what makes personality-at-launch mechanical rather
than hopeful. Two concurrent agents sharing a (harness, role) pair share a
seat by design; containers are the disambiguator when that is wrong.

Host caveat: a host whose v1 install is a dormant snapshot (this
workstation, per the note in its `agent-compose.yaml`) needs either a
composer run from the agentic-os checkout or a one-time hand append. Fleet
hosts converge through infrastructure.

## Container tier: v2 owns the home

Ward mounts a bundle read-only and container start projects it into
container-HOME load points - home-scope variants of the projection registry.
The container carries exactly one member's identity; the rest of the roster
is absent from the filesystem, which is the hard isolation the host tier
deliberately does not promise. Issue #17 owns this slice, including
verifying each harness's global load-point path before the layouts land.

## What migrates in v0.1

Nothing moves out of v1. The v0.1 migration is additive: publish the roster
artifact into the cascade, release the v2 binary, and leave retirement
questions to a future contract. The old #9 framing - moving AOS composition
itself - is superseded by this document.

## See also

* [projection.md](projection.md) - the load-point layer both tiers drive.
* [launch.md](launch.md) - refresh-then-exec and the recursion guard.
* [person-contract.md](person-contract.md) - roles, seats, and colors.
* [architecture.md](architecture.md) - composition inputs and ownership.
