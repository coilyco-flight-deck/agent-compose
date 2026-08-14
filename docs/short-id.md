# The dictatable short id

Terminal surfaces append the running session's short id to the rendered name:

```text
Angie [she] (Engineer) uz86
```

Four characters, two letters then two digits, over an alphabet that drops the
visually and phonetically confusable ones (`i l n o`, `0 1 2 3`).

## Where the contract lives

The shape and alphabet come from the archived o2r channel protocol. agentic-os
holds the canonical definition in `agentic_os/agent_id.py`, with a
cross-language vector file pinning it, and `aos` mints session ids from the same
contract in `aos-cli/native_shadow.go`.

Agent Compose duplicates two constants rather than taking a dependency in the
wrong direction, because it only ever needs to **recognise** an id. If the
alphabet changes upstream, `internal/agentid` is the second place to edit.

## Read, never minted

Agent Compose reads `AOS_NATIVE_SESSION` and never generates an id.

The status line re-renders on every tick, so a freshly minted id would differ
each time and label nothing, which is worse than showing none. A seeded id would
need a stable seed this process does not have. The session id is already unique
per agent, already dictatable, and already names the shadow directory the agent
works in, so reading it keeps every surface agreeing with `aos`.

Outside a native session there is no id, and every surface renders exactly what
it rendered before. A value that is set but malformed is dropped rather than
shown: displaying a non-dictatable id breaks the alphabet's only promise.

## Ephemeral surfaces only

The id reaches the status-line session row and the launch `--name` flag. Both
are computed per tick or per launch.

It is deliberately kept out of overlay documents and nativeui settings. Those
are written once and read by later sessions, and the bundle cache key hashes the
rendered instructions, so an id baked into either would name the wrong agent on
reuse and fork the cache per session. `TestShortIDNeverReachesPersistedBundleArtifacts`
walks the whole bundle directory to hold that line.

## Why the subagent rows omit it

The id names the session. Every row in one agent panel would repeat the same
four characters and disambiguate nothing. Where it earns its place is telling
two concurrently running sessions apart, and that is the session row.

## See also

* [Naming the seat](seat-identity.md) - renaming the composed seat.
* [Projected composition status line](statusline.md) - the row it appears on.
