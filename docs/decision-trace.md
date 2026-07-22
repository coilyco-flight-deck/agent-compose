# Decision trace

The resolver builds the trace while evaluating facts, candidates, and rules.
The renderer never reconstructs reasons after composition.

## Shape

`trace.json` uses this stable top-level structure:

* `protocol` - `agent-compose.trace`.
* `version` - `0.1`.
* `facts` - normalized inputs with owner and provenance.
* `candidates` - every person, instruction, capability, skill, and delivery
  candidate considered by the resolver.
* `decisions` - ordered resolution nodes with parent references.

A fact contains `id`, `name`, `value`, `owner`, and `provenance`. A candidate
contains `id`, `kind`, `source`, `delivery_key`, `digest`, decision-relevant
`attributes`, and `provenance`. Attributes retain normalized policy metadata
such as tier, capability, rank, and scopes, but never copy instruction bodies.
A decision contains `id`, `parents`, `phase`, `subject`, `rule`, `outcome`,
`reason_code`, `reason`, and `inputs`.

Decision ids are stable within a trace and follow evaluation order. Parent ids
form an acyclic graph without duplicating nodes. Renderers may group the flat
ordered nodes for progressive disclosure. Semantic diff keys decisions by
phase, subject, rule, and inputs rather than by their sequence ids.

## Outcomes

The v0.1 outcomes are:

* `selected` - policy admitted the candidate.
* `excluded` - policy considered and rejected the candidate.
* `shadowed` - byte-identical content at a higher precedence deduplicated the
  candidate for the same delivery key.
* `fallback` - the preferred candidate was unavailable and policy selected an
  allowed alternative.
* `delivered` - the adapter assigned selected content to a bundle entry point.

Invalid input fails composition and produces diagnostics from the in-progress
trace, but the materializer emits no bundle.

## Provenance and redaction

Provenance is tagged as `request`, `source`, or `decision`. Request provenance
contains the normalized request field path. Source provenance contains a stable
source id, source kind, relative document path, logical KDL node path, and source
digest. Decision provenance contains the producing decision id. No form contains
an absolute path, credential, auth-store path, opaque host identifier, or private
source body.

The trace may record a generic private-overlay source id and its digest. The
trace records why private content was selected without copying that content
into reasons. A source author keeps reason strings public-safe.

## Determinism

Facts follow request order, candidates follow source precedence and source
order, and decisions follow evaluation order. Maps serialize with canonical
key ordering for bundle identity. Runtime duration, cache location, cache-hit
status, terminal width, and color decisions remain outside the trace.

The stored trace is sufficient for later describe, item-level why, and semantic
diff commands without reopening source files. Human output and TTY styling are
views over this data and never enter model instructions.

## See also

* [bundle-protocol.md](bundle-protocol.md) - trace retention and digesting.
* [architecture.md](architecture.md) - resolver flow and ownership.
