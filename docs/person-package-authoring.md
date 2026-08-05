# Person package authoring

An external package uses the same validated layout as the embedded default:

```text
person.kdl
roles/
personalities/
inspirations/
definitions/INVARIANT.md
definitions/skills/<skill>/SKILL.md
```

`person.kdl` contains only `person "<name>"`. Each policy node lives in one
ordered KDL fragment:

```text
roles/01-builder.kdl
personalities/01-curious.kdl
inspirations/01-example.kdl
```

The two-digit prefix controls order. The remaining filename must match the
node slug. Every bound personality needs one definition directory, and the
definitions directory may not contain extra skills.
Symlinks are invalid anywhere in the package.

The existing role, personality, identity, inspiration, color, model-tier, and
model-class validation applies unchanged. A missing, malformed, or internally
inconsistent package fails before bundle materialization or host projection.
Person packages never transport credentials or launcher authority.

## See also

* [person-packages.md](person-packages.md) - selection and machine use.
* [person-contract.md](person-contract.md) - KDL policy schema.
