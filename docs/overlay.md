# Identity overlay

`agent-compose overlay` projects one selected person-package member into a
compact surface.
It is deliberately noninteractive. Terminal panes can use the plain card, and
other renderers can consume the same versioned JSON document.

## Text

Supply a role, harness seat, and renderer state:

```sh
agent-compose overlay \
  --person-source ./person \
  --role platform \
  --seat codex \
  --expression acting
```

Omit `--person-source` to use the embedded `roster:core` default.

The default text width is 40 columns. `--width 200` collapses the same fields
onto one line. Output contains no control sequences, so pipes and CI receive
stable plain text.

## JSON

Add `--json` for `agent-compose.overlay.v1`. The document contains:

* person, role, `role_display_name`, purpose, and selected seat
* `annotation`, the composed identity string a renderer shows verbatim
* the caller-supplied expression
* the role's derived favorite color
* every component personality's color and identity primitives

The JSON is a projection of the selected person model, not a second policy
source.

## Annotation

`annotation` is the one identity string every terminal surface shows for a
session, so a window title, a status row, and a launch flag never drift apart:

```text
Angie [she] (Agentic Platform Engineer)
```

Agent Compose owns the shape. A renderer that has the document shows the field
rather than reassembling it from the seat name, pronouns, and role. Both fields
are additive to `agent-compose.overlay.v1`, so a consumer built before them
keeps parsing the document unchanged.

The plain text card stops at `Angie [she]`, because it already prints the role
on its own line.

## State boundary

The caller supplies one expression from the fixed person vocabulary. Unknown
roles, seats, and expressions fail closed. Agent-compose never inspects a
process, trace, log, queue, agent runtime, or launcher state to infer an expression.

This keeps the overlay suitable for terminals, browser shells, streams, and
future mobile surfaces without turning identity data into observability.

## See also

* [identity-primitives.md](identity.md) - renderer semantics.
* [person-packages.md](person-packages.md) - external package selection.
* [person-snapshot.md](person-contract.md) - complete person export.
