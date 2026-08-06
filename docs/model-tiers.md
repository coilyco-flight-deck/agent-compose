# Core model-tier matrix

The role KDL fragments under `internal/person/roles/` own this policy. This
page is its human-readable reference inventory. Model tier controls role
compatibility, not context selection, permissions, or executable authority.
Every tier supported by a role receives the same complete selected context.

Frontier covers Claude and Codex, commodity covers DeepSeek, and OSS covers
local or open models such as Ornith and Mistral.

## Category 1: complex roles

These roles require frontier reasoning for complex evidence or consequential
decisions.

* Director - frontier
* Portfolio Strategist - frontier
* DevOps - frontier
* AI Engineer - frontier

## Category 2: foundational roles

These roles produce bounded artifacts or verdicts that fit frontier and
commodity models.

* Engineer - frontier, commodity
* QA - frontier, commodity
* Designer - frontier, commodity
* Content Creator - frontier, commodity, OSS

## Category 3: high-security roles

Content Creator includes an OSS-classified Discord seat so callers can keep
sensitive community context on a local or open-model route. Role compatibility
does not choose a route or grant access. The launch consumer still selects the
appropriate tier and controls the supplied context.

## Evaluation state

Evaluation packs retain all three lanes so matrix changes remain visible.
Role-incompatible lanes are disabled. Commodity and OSS execution also stays
disabled until independently reviewed evidence admits those lanes. AI Engineer
remains frontier-only until a complete lower-tier evidence lane passes.

## See also

* [person-contract.md](person-contract.md) - owning roster schema.
* [evaluation.md](evaluation.md) - tiered behavior evidence.
* [architecture.md](architecture.md) - model identity ownership.
