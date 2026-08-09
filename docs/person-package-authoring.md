# Person package authoring

An external package uses the same validated layout as the embedded default:

```text
person.kdl
roles/
roles/<role>/skills/<method>/SKILL.md
personalities/
[inspirations/]
definitions/INVARIANT.md
definitions/skills/<skill>/SKILL.md
```

`person.kdl` contains only `person "<name>"`. Each policy node lives in one
ordered KDL fragment:

```text
data/role-builder/role.kdl
data/role-builder/SKILL.md
[data/role-builder/evals.yaml]
data/personality-curious/personality.kdl
data/personality-curious/SKILL.md
[data/boundary-shared-thing/boundary.kdl]
[data/invariant/INVARIANT.md]
```

Every first-class entity owns one flat directory named `<kind>-<slug>`, where
kind is `role`, `personality`, `boundary`, or `inspiration`. Its KDL fragment is
named for the kind, its body is `SKILL.md`, and a role may add `evals.yaml`. The
directory slug must match the node slug.

Each entity declares an `order`, which sequences the roster in place of the
filename prefixes the layout used to carry. Order is data on the entity, so
moving a directory never reorders anything. The loader strips it before parsing,
so it never reaches the node model.

The invariant lives at `data/invariant/INVARIANT.md`. Every bound personality
needs its own directory, every boundary must be referenced by at least one role,
and a boundary's [owner](boundary-owners.md) names a defined role that must not
declare it. Symlinks are invalid anywhere in the package.

The role, personality, identity, color, and model-tier validation applies
unchanged. Inspiration relationships and their catalogue remain optional
compatibility data for independently authored packages. When supplied, every
reference must resolve, every entry must be used, and every record must remain
complete. A missing, malformed, or internally inconsistent package fails
before bundle materialization or host projection. Person packages never
transport credentials or launcher authority.

## See also

* [person-packages.md](person-packages.md) - selection and machine use.
* [person-contract.md](person-contract.md) - KDL policy schema.
* [role-methods.md](role-methods.md) - method binding and delivery.
* [role-boundaries.md](role-boundaries.md) - shared doctrine binding and delivery.
