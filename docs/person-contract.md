# Embedded person contract

The binary embeds exactly one `person` source for v0.1. No provider interface
or trait-slider abstraction sits in front of it:

```kdl
person "kai" {
    role "engineer" {
        purpose "Write code, merge code, stay focused on your goal."
        personality "curious" "grounded" "meticulous"
        agent "claude" name="opal engineer" pronouns="she"
        agent "codex" name="terran engineer" pronouns="he"
    }
    personality "curious" skill="personality-curious"
}
```

A role names its purpose, the personalities compatible with it, and its named
agent seats. A personality entry binds the name to the skill that defines it.
The definition itself - however presence, attention, tempo, and voice end up
being expressed - lives inside that skill, not in this contract. The format
inside the skill (KDL, YAML, or a metadata block in SKILL.md) is deliberately
undecided in v0.1.

## Agent seats

An `agent` node under a role is a named seat: the harness lineage is the key,
and `name` plus `pronouns` are the identity that seat wears. This shape is
adapted from the seat sketch in ward's `roles.kdl`, and the shared role slug
remains the join key between the two: ward keeps guardfiles, models, and
reasoning effort on its side, and nothing here grants authority. Names are
opaque strings to the engine - that a name's flavor happens to track which
model the launcher assigns to a seat is naming taste, not schema.

The embedded source carries the six active roles (engineer, director, qa,
advisor, ops, pm), their twelve named seats, and the social and sales stubs.
A role may omit `personality` until the full catalog assigns its compatible
set; issue #10 owns completing personalities and per-seat defaults.

A private overlay may add scoped instructions or selection rules. It may not
redefine canonical roles, personalities, seats, or compatibility. AOS owns no
copy of this person source.

## See also

* [kdl-contracts.md](kdl-contracts.md) - request and source grammar.
* [architecture.md](architecture.md) - policy ownership.
