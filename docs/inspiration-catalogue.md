# Inspiration catalogue

Agent-compose embeds a credited inspiration catalogue beside Kai's roles and
personalities. The catalogue explains which public figures informed each
assignment, why the relationship fits, and which public work supports that
reading.

## What ships

The canonical `internal/person/person.kdl` source contains three connected
layers:

* Every role and personality owns one `inspiration` reference plus its own fit
  rationale.
* Every normalized inspiration owns one name, representative achievement,
  impact mode, impact rationale, and profile citation key.
* Every inspiration owns one selected public speaking appearance with title,
  event, year, format, summary, and citation keys.

The fit stays on the relationship because one person can inform a company role
and a personality for different reasons. Shared biographical and impact data
stays on the normalized inspiration so the source never maintains parallel
copies.

## Validation

The embedded source fails to load when a role or personality omits its
inspiration, a reference does not resolve, two entries name the same person, or
an entry is unused. Each inspiration also needs its complete achievement,
impact, profile, and speaking-appearance record. An appearance needs a curated
summary and at least one unique citation key.

These rules protect relationships and completeness without hard-coding the
current catalogue size in tests. Kai can revise the policy source without
teaching a test to repeat its content.

## Identity boundary

An inspiration is an acknowledgement, not an agent identity. Agent-compose
does not name a seat after the credited person or ask a model to imitate her.
Bundles include only compact credit, impact-mode, and appearance metadata.
Long biography, fit, summary, and citation records stay out of runtime prompts.

The raw person source participates in bundle hashing, so policy changes remain
traceable even though long inspiration prose does not enter the prompt.

## Provenance

Citation keys are stable, meaningful identifiers for the public profiles,
talks, transcripts, and editorial coverage selected during curation. The
source links and longer editorial notes live on
[agent-compose #60](https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/issues/60).
The binary performs no network fetch, media download, or transcript ingestion.

This split keeps public evidence reviewable without turning agent-compose into
a media cache or a second knowledge-provider system.

## Consumption

The in-memory person model exposes the catalogue to agent-compose renderers.
Normal host convergence writes the complete versioned person snapshot at
`~/.agent-compose/sources/personality/person.json`. It is the machine-readable
boundary for pickers, visual projections, and other consumers. Consumers
should join on stable role, personality, inspiration, and appearance slugs
rather than copying curated prose.
