# Integration and delivery tiers

Agent-compose is the shared context substrate for AOS sources, Ward, and native
harnesses. AOS authors reusable knowledge. Agent-compose materializes the
selected context surface. Ward supplies authority and consumes that surface
without parsing AOS policy.

The v1 Python composer (`generate-agent-compose` in AOS) has been
absorbed: [`agent-compose cascade`](cascade.md) now composes
`AGENTS.COMPOSE.md` doctrine sources into
`~/.agent-compose/COMPOSED.<harness>.md`, symlinks each harness's
global load point at the result, applies scope and harness filtering with
per-harness section overrides, and emits the mount-eligibility manifest ward
reads - byte-compatible with the Python outputs. The Python generator
remains in AOS only until fleet hosts converge on the binary. Retirement steps
(hook repointing and script removal) are AOS work.

## The seam rule

On a host, the cascade owns the harness global load points and everything
else (roster, overlays) feeds it sources. In a container, projection owns
the whole home and no cascade runs. No path is ever written by both.

Host convergence may also mount skills from repositories already admitted by
`mount-eligibility.json` into configured harness-native skill directories.
Agent-compose owns only the links recorded in its sidecar. Infrastructure still
owns the load points a host declares.

## Host tier: context rides the native cascade

Agent-compose renders a roster artifact into `~/.agent-compose/sources/`, a
directory the cascade walks as a source root, containing an
`AGENTS.COMPOSE.md` entry plus the personality files it references. The entry
carries admitted provider instructions plus the seat dispatch table: "if you
are codex running the engineer role, your name is terran engineer (he/him),
your personalities are curious, grounded, and meticulous, their definitions
live at these paths, and their melded favorite is this derived color." Each
agent self-selects by facts it already knows and loads the linked definitions
for that role before acting. Definitions on other roles stay inactive. No
launcher cooperation, environment variable, or blessed entrypoint
participates.

Running `agent-compose cascade` then carries the table into every harness's
global load point - one binary, no Python. Global context loads at session
start unconditionally, which is what makes personality-at-launch mechanical
rather than hopeful. Two concurrent agents sharing a (harness, role) pair
share a seat by design; containers are the disambiguator when that is wrong.

## Container tier: v2 owns the home

Agent-compose now supplies the stable provider half: `verify` checks a
read-only bundle, then `project --scope home` transactionally fills the
claude, codex, goose, or opencode global load points. Black-box fixtures prove
that each native home contains every personality identity activated by the
role, compiled homes contain all of their prose, and neither path changes the
input bundle. Ward still needs to mount and invoke this contract at container
start under issue #17.

## Migration state

The cascade is native as of v0.2.0. What remains is fleet cutover: hosts
switch from the Python entry points to the binary, the agentic-os drift hook
repoints at `agent-compose cascade --check`, and the Python generator
retires. That work belongs to AOS and infrastructure and is tracked in
[agentic-os#618](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/issues/618).
A host with
hand-edited COMPOSED snapshots (this workstation) must reconcile edits back
into sources before its first native cascade run, or they will be
regenerated away.

## See also

* [projection.md](projection.md) - the load-point layer both tiers drive.
* [launch.md](launch.md) - refresh-then-exec and the recursion guard.
* [person-contract.md](person-contract.md) - roles, seats, and colors.
* [architecture.md](architecture.md) - composition inputs and ownership.
