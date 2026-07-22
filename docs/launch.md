# Launch-time refresh

The launch verb freshens context, then hands the process to the real command:

```
agent-compose launch --request <kdl> --layout <name> --target <dir> -- <command> [args...]
```

Refresh is compose plus project. Both halves are already idempotent - the
bundle cache reuses identical inputs and projection replaces only its own
files - so a warm launch is a no-op that validates and execs. The warm path
runs in single-digit milliseconds on the reference host, well inside the
250 ms budget the test suite enforces.

## Recursion guard

Launch sets `AGENT_COMPOSE_LAUNCH` in the child environment before exec. A
nested launch sees the sentinel, skips refresh, and execs straight through,
so a shadowed binary that wraps a harness can never recurse into itself.

## Failure behavior

A refresh failure never blocks a launch that has context to run with. When
compose or project fails and the target holds a validated last-known-good
projection - every file the sidecar records still present - launch warns
loudly on stderr and proceeds with it. Without a usable previous projection
the launch aborts. A refresh failure touches only the bundle cache and
projection-owned files; credentials and mutable harness configuration are
never in its write path.

## Concurrency

Concurrent identical launches converge on one cache entry: the materializer
stages beside the target and the rename loser reuses the winner. Concurrent
launches for different requests or targets stay isolated by construction -
distinct cache keys, distinct target directories, and a per-target lock file
(`.agent-compose/lock`) serializing projection writes.

## Wrapper installation requirements

Binary shadowing rollout belongs to the infrastructure repo. A wrapper that
fronts a harness must exec `agent-compose launch` with its fixed request,
layout, and target, forward the original argv after `--`, and resolve the
real harness binary through normal PATH lookup - the sentinel, not PATH
surgery, is what prevents recursion. No rollout code lives here.

## See also

* [projection.md](projection.md) - the load-point layer launch drives.
* [bundle-protocol.md](bundle-protocol.md) - cache identity and atomicity.
* [architecture.md](architecture.md) - composition inputs and ownership.
