# Decision trace

Composition is chatty. The resolver records what it picked and why as each
decision occurs, into `trace.json` inside the bundle. Reasons are never
reconstructed after composition.

A trace is an ordered list of decisions. Each decision names the thing being
decided - a personality skill, an instruction, a delivery entry point - the
source it came from, the outcome, and a human-readable reason. The v0.1
outcomes are:

* `selected` - policy admitted the candidate.
* `excluded` - policy considered and rejected the candidate.
* `shadowed` - an identical higher-precedence copy already filled the slot.
* `delivered` - the adapter placed selected content at a bundle entry point.

Invalid input fails composition with diagnostics from the in-progress trace,
and no bundle is produced.

Reasons are plain sentences, safe to show in a terminal and safe to keep in a
public bundle. A private overlay is referenced by its source id; its content
never appears in a reason. Runtime noise - durations, cache hits, terminal
styling - stays out of the trace.

`agent-compose describe` renders the stored decisions in scannable sections,
`describe --why <item>` follows one item to its outcome, and `diff` compares
two bundles by decision subject rather than file bytes - all without
reopening source files. `trace.json` itself is the machine-readable surface;
there is no second explanation format. Human output and TTY styling are views
over the trace and never enter model instructions; redirected output is plain
and deterministic.

## See also

* [bundle-protocol.md](bundle-protocol.md) - where the trace lives.
* [architecture.md](architecture.md) - resolver flow and ownership.
