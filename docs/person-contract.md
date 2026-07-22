# Embedded person contract

The binary embeds exactly one `person` source for v0.1. No provider interface
or trait-slider abstraction sits in front of it:

```kdl
person "kai" {
    role "engineer" {
        purpose "Build and land independently verifiable product changes."
        personality "curious" "grounded" "meticulous"
    }
    personality "curious" skill="personality-curious"
}
```

A role names its purpose and the personalities compatible with it. A
personality entry binds the name to the skill that defines it. The definition
itself - however presence, attention, tempo, and voice end up being expressed -
lives inside that skill, not in this contract. The format inside the skill
(KDL, YAML, or a metadata block in SKILL.md) is deliberately undecided in
v0.1.

The person source owns organizational purpose, role-neutral personalities, and
curated role compatibility. Issue #10 owns the complete roster and prose.

A private overlay may add scoped instructions or selection rules. It may not
redefine canonical roles, personalities, or compatibility. AOS owns no copy of
this person source.

## See also

* [kdl-contracts.md](kdl-contracts.md) - request and source grammar.
* [architecture.md](architecture.md) - policy ownership.
