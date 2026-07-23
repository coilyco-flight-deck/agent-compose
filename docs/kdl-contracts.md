# KDL contracts

Agent-compose uses KDL for human-authored requests and policy. Parsers fail on
unknown nodes, duplicate scalar facts, missing required sources, or an empty
selection.

## Compose request

A request names a role, a delivery mode, a density, and the sources personality
files come from:

```kdl
compose {
    role "engineer"
    delivery "native-skills"
    density "full"
    source "aos-public" declaration="source-public.kdl" required=#true
}
```

The role activates its complete ordered personality set from the embedded
person source. A request cannot narrow that set with a personality selector.
`delivery` is `native-skills` or `compiled`. `density` is `brief` or `full`
and only changes how much personality prose the bundle carries. A caller
usually derives it from model class. Nothing else about the agent - model,
harness, reasoning effort, interactivity - appears in a request.

Sources are evaluated in request order. A declaration path is locator data
that says where files live; it never becomes part of the composed content's
identity.

## Personality source declaration

A source is a place personality files live - an AOS checkout, a private
overlay directory, or coincidentally a repository:

```kdl
source "aos-public" {
    instruction "foundation" path="content/foundation.md"
    skill "personality-curious" path="skills/personality-curious"
}
```

Paths are relative to the declaration and must stay beneath the source root.
Symlinks and escaping paths fail validation. A required missing source fails
composition. An optional missing source is skipped with a note in the trace.

## See also

* [architecture.md](architecture.md) - composition inputs and ownership.
* [bundle-protocol.md](bundle-protocol.md) - machine-readable output.
* [person-contract.md](person-contract.md) - embedded personal policy.
* [contract-review.md](contract-review.md) - review decisions of record.
