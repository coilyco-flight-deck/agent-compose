# Complete person snapshot

Normal bare `acompose` convergence and the hidden roster compatibility command
write `person.json` into the roster artifact directory. The default location is
`~/.agent-compose/sources/personality/person.json`.

## Contract

The JSON format marker is `agent-compose.person-snapshot.v2`, with numeric
schema version `2`. The artifact exports:

* the person name, `person:kai` source provenance, and canonical role order
* every role's purpose, verbatim long-form briefing, ordered personality meld,
  derived favorite color, inspiration relationship, and named seats
* every personality's skill binding, color, emblem, motif, form language,
  sound mark, and inspiration relationship
* the fixed renderer expression vocabulary
* the normalized inspiration and speaking-appearance catalogue

Roles and personalities are keyed by their stable slugs. `role_order` is the
canonical presentation order. Consumers should use the explicit order rather
than relying on JSON object order.

Schema v2 adds identity primitives and expressions. Consumers pinned to v1
must upgrade before treating the format marker as compatible.

## Convergence

The snapshot is generated from the loaded person model in the same owned,
transactional roster projection as the human-readable files. A failed
projection restores the prior owned artifact. A second convergence leaves
identical bytes unchanged.

The `compose` terminal transcript and composed agent instructions render the
selected-role slice from this same model. Each inspiration record stays with
the role or personality relationship it describes. Both surfaces include
identity primitives, fits, achievements, impacts, profile citations, and the
expression vocabulary. The appearance catalogue remains snapshot-only.

The generated file remains outside repositories. Consumers can read it but do
not edit it or treat it as a second policy source.

## Authority boundary

The artifact describes public identity and orientation only. It contains no
model choice, reasoning effort, permission, credential, endpoint, routing
decision, or runtime authority. Ward and deployment-specific systems keep
those fields.

## See also

* [person-contract.md](person-contract.md) - canonical embedded KDL model.
* [inspiration-catalogue.md](inspiration-catalogue.md) - credited evidence model.
* [role-briefings.md](role-briefings.md) - unconditional role charter.
