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

The trace also retains one provider report for every selected provider and
every configured provider excluded from the active role. Reports classify the
selected person package, ordinary catalogues, and role-only skill-provider repositories, then
record their configuration scope, outcome, and reason.

When a role provider carries an ordinary-skill selector, its report also
records the configured selector and the admitted catalogue fraction. Skills
outside that slice remain explicit excluded decisions, so `describe --why`
can distinguish selector filtering from provider or role exclusion.

Each provider report carries a context-budget contribution. `skills` counts
canonical selected skill trees attributed to that provider. `context_bytes`
is their exact retained byte count, and `approximate_tokens` uses
`ceil(context_bytes / 4)`. Shadowed copies do not contribute twice. Excluded
providers record explicit zero values for all three fields. Native and staged
projection preserve the same trace, so the budget does not depend on the
consumer layout. Selector-backed provider budgets count only the admitted
slice.

Invalid input fails composition with diagnostics from the in-progress trace,
and no bundle is produced.

Reasons are plain sentences, safe to show in a terminal and safe to keep in a
public bundle. A private overlay is referenced by its source id; its content
never appears in a reason. Runtime noise - durations, cache hits, terminal
styling - stays out of the trace.

`agent-compose describe` renders provider outcomes, context budgets, and the
stored decisions in scannable sections,
`describe --why <item>` follows one item to its outcome, and `diff` compares
two bundles by decision subject plus manifest logical content ID and digest.
Artifact-level changes remain visible beside logical changes. These commands
do not reopen authoring roots. `trace.json` itself is the decision
machine-readable surface;
there is no second explanation format. Human output and TTY styling are views
over the trace and never enter model instructions; redirected output is plain
and deterministic.

## See also

* [bundle-protocol.md](bundle-protocol.md) - where the trace lives.
* [architecture.md](architecture.md) - resolver flow and ownership.
* [catalogues-and-export.md](catalogues-and-export.md) - logical content diff semantics.
