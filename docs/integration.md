# Integration and delivery tiers

The v1 Python composer (`generate-agent-compose` in agentic-os) has been
absorbed: [`agent-compose cascade`](cascade.md) now composes
`AGENTS.COMPOSE.md` doctrine sources into
`~/.config/agent-compose/COMPOSED.<harness>.md`, symlinks each harness's
global load point at the result, applies scope and harness filtering with
per-harness section overrides, and emits the mount-eligibility manifest ward
reads - byte-compatible with the Python outputs. The Python generator
remains in agentic-os only until fleet hosts converge on the binary;
retirement steps (hook repointing, script removal) are agentic-os work.

## The seam rule

On a host, the cascade owns the harness global load points and everything
else (roster, overlays) feeds it sources. In a container, projection owns
the whole home and no cascade runs. No path is ever written by both.

## Host tier: personality rides the v1 cascade

V2 renders a roster artifact into `~/.config/agent-compose/sources/` - a
directory v1 already walks as a source root - containing an
`AGENTS.COMPOSE.md` entry plus the personality files it references. The entry
carries the seat dispatch table: "if you are codex running the engineer
role: your name is terran engineer (he/him), your personality is grounded,
its definition lives at <path>, your favorite color is #5fa87a." Each agent
self-selects by facts it already knows; no launcher cooperation, no
environment variable, no blessed entrypoint.

Running `agent-compose cascade` then carries the table into every harness's
global load point - one binary, no Python. Global context loads at session
start unconditionally, which is what makes personality-at-launch mechanical
rather than hopeful. Two concurrent agents sharing a (harness, role) pair
share a seat by design; containers are the disambiguator when that is wrong.

## Container tier: v2 owns the home

Ward mounts a bundle read-only and container start projects it into
container-HOME load points - home-scope variants of the projection registry.
The container carries exactly one member's identity; the rest of the roster
is absent from the filesystem, which is the hard isolation the host tier
deliberately does not promise. Issue #17 owns this slice, including
verifying each harness's global load-point path before the layouts land.

## Migration state

The cascade is native as of v0.2.0. What remains is fleet cutover: hosts
switch from the Python entry points to the binary, the agentic-os drift hook
repoints at `agent-compose cascade --check`, and the Python generator
retires - all agentic-os and infrastructure work, filed there. A host with
hand-edited COMPOSED snapshots (this workstation) must reconcile edits back
into sources before its first native cascade run, or they will be
regenerated away.

## See also

* [projection.md](projection.md) - the load-point layer both tiers drive.
* [launch.md](launch.md) - refresh-then-exec and the recursion guard.
* [person-contract.md](person-contract.md) - roles, seats, and colors.
* [architecture.md](architecture.md) - composition inputs and ownership.
