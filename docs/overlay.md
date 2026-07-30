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
  --role engineer \
  --seat codex \
  --expression acting
```

Omit `--person-source` to use the embedded `roster:core` default.

The default text width is 40 columns. `--width 200` collapses the same fields
onto one line. Output contains no control sequences, so pipes and CI receive
stable plain text.

## JSON

Add `--json` for `agent-compose.overlay.v1`. The document contains:

* person, role, purpose, and selected seat
* the caller-supplied expression
* the role's derived favorite color
* every component personality's color and identity primitives

The JSON is a projection of the selected person model, not a second policy
source.

## State boundary

The caller supplies one expression from the fixed person vocabulary. Unknown
roles, seats, and expressions fail closed. Agent-compose never inspects a
process, trace, log, queue, agent runtime, or launcher state to infer an expression.

This keeps the overlay suitable for terminals, browser shells, streams, and
future mobile surfaces without turning identity data into observability.

## See also

* [identity-primitives.md](identity-primitives.md) - renderer semantics.
* [person-packages.md](person-packages.md) - external package selection.
* [person-snapshot.md](person-snapshot.md) - complete person export.
