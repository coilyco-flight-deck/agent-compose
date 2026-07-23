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
    source "aos-public" root="agentic-os" required=#true
}
```

The role activates its complete ordered personality set from the embedded
person source, every ordinary provider skill, and its composed-skill allowlist.
A request cannot narrow those sets with a selector.
`delivery` is `native-skills` or `compiled`. `density` is `brief` or `full`
and only changes how much personality prose the bundle carries. A caller
usually derives it from model class. Nothing else about the agent - model,
harness, reasoning effort, interactivity - appears in a request.

Sources are evaluated in request order. A `root` or `declaration` path is
locator data that says where files live. It never becomes part of the composed
content's identity.

## Personality sources

The public AOS provider needs only its root. Agent-compose discovers the shared
`personality-invariant` and every ordinary skill under `.agents/skills` in
lexical order. It also reads `.agents/roles.kdl`:

```kdl
roles {
    role "engineer" {
        composed-skill "coding-shape-cli"
        intent "autonomous-coding" {
            harness "openhands"
        }
    }
}
```

Each `composed-skill` binding admits `.agents/composed/<name>/COMPOSED.md` only
for that role. Each `intent` records one model-opaque default harness route for
that role. Agent-compose validates and preserves those routes as composition
policy. Ward does not parse them.
Materialization renames the admitted entry point to `SKILL.md`. A `SKILL.md`
anywhere under `.agents/composed` and ordinary/composed name collisions fail.
The same root form works in requests, roster arguments, and `roster_sources`.

An overlay or another provider can instead carry an explicit declaration:

```kdl
source "aos-public" {
    instruction "foundation" path="content/foundation.md"
    skill "personality-curious" path="skills/personality-curious"
}
```

The request admits that file with
`source "aos-public" declaration="source-public.kdl"`. Paths inside the
declaration are relative to the declaration and must stay beneath its source
root. Request locator paths are relative to the request. Symlinks and escaping
paths fail validation. A required missing source fails composition. An optional
missing source is skipped with a note in the trace.

## See also

* [architecture.md](architecture.md) - composition inputs and ownership.
* [bundle-protocol.md](bundle-protocol.md) - machine-readable output.
* [person-contract.md](person-contract.md) - embedded personal policy.
* [contract-review.md](contract-review.md) - review decisions of record.
