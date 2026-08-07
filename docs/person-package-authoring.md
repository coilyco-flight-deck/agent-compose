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
roles/01-builder.kdl
[melds/01-shared-boundary.kdl]
personalities/01-curious.kdl
[inspirations/01-example.kdl]
```

The two-digit prefix controls order. The remaining filename must match the
node slug. Every bound personality needs one definition directory, and the
definitions directory may not contain extra skills. Each optional role method
is declared in the role fragment and stored below that role's `skills/`
directory. Method directories contain only `SKILL.md`. Each optional meld is
declared in its own fragment, stores its body beside the personality
definitions, and must be referenced by at least one role.
Symlinks are invalid anywhere in the package.

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
* [role-melds.md](role-melds.md) - shared doctrine binding and delivery.
