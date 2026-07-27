# Integration and delivery tiers

Agent-compose is the shared context substrate for knowledge sources, native
harnesses, and isolated launch consumers. It materializes the selected context
surface while launchers keep authority outside that surface.

The v1 Python composer was absorbed into
[`agent-compose cascade`](cascade.md), which composes doctrine sources into
`~/.agent-compose/COMPOSED.<harness>.md`, symlinks each harness's
global load point at the result, applies scope and harness filtering with
per-harness section overrides, and emits a generic mount-eligibility manifest.

## The seam rule

On a host, the cascade owns the harness global load points and everything
else (roster, overlays) feeds it sources. In a container, projection owns
the whole home and no cascade runs. No path is ever written by both.

Host convergence may also mount skills from repositories already admitted by
`mount-eligibility.json` into configured harness-native skill directories.
Agent-compose owns only the links recorded in its sidecar. Infrastructure still
owns the load points a host declares.

## Host tier: context rides the native cascade

Agent-compose renders the selected person package into a roster artifact under
`~/.agent-compose/sources/`, a directory the cascade walks as a source root,
containing an
`AGENTS.COMPOSE.md` entry plus the personality files it references. The entry
carries the selected personality invariant, admitted overlay instructions, and
the seat dispatch table: "if you
are codex running the engineer role, this is your operating briefing, name,
pronouns, personality meld, definition paths, and derived favorite color."
Host config may select one
[external person package](person-packages.md), which replaces the embedded
default before this artifact is rendered.
Under the [role-selection contract](role-selection.md), an explicit role stays
fixed. An unassigned native agent selects from the initial request and loads
that role's definitions. The embedded default needs no external source.

Running `agent-compose cascade` then carries the table into every harness's
global load point - one binary, no Python. Global context loads at session
start unconditionally, which is what makes personality-at-launch mechanical
rather than hopeful. Two concurrent agents sharing a (harness, role) pair
share a seat by design. Containers disambiguate when that is wrong.

## Container tier: v2 owns the home

`verify` checks a read-only bundle, then `project --scope home`
transactionally fills the claude, codex, goose, or opencode global load
points. Black-box fixtures prove that each native home contains ordinary and
role-composed skills plus every active personality, compiled homes contain all
selected prose, and neither path changes the input bundle. Projected
instructions fix the caller-selected role under the
[same contract](role-selection.md). A composition adapter can use an empty
private target under the [staged-home contract](staged-home.md).

## Migration state

The cascade is native as of v0.2.0. Fleet cutover belongs to AOS and
infrastructure and is tracked in
[agentic-os#618](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/issues/618).
Hosts must reconcile hand-edited snapshots into sources before cutover.

## See also

* [projection.md](projection.md) - the load-point layer both tiers drive.
* [staged-home.md](staged-home.md) - isolated adapter handoff.
* [launch.md](launch.md) - refresh-then-exec and the recursion guard.
* [person-contract.md](person-contract.md) - roles, seats, and colors.
* [person-packages.md](person-packages.md) - external package selection.
* [role-briefings.md](role-briefings.md) - role charter delivery.
* [architecture.md](architecture.md) - composition inputs and ownership.
