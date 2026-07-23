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
    personality "curious" skill="personality-curious" color="#d98e48"
}
```

A role names its purpose, the personalities compatible with it, and its named
agent seats. The embedded roster contains this approved compatibility matrix:

* `director` - `bold`, `grounded`, `diplomatic`
* `advisor` - `reflective`, `curious`, `candid`
* `pm` - `warm`, `meticulous`, `curious`
* `designer` - `imaginative`, `playful`, `warm`
* `engineer` - `curious`, `grounded`, `meticulous`
* `qa` - `meticulous`, `candid`, `playful`
* `ops` - `protective`, `grounded`, `reflective`
* `sales` - `charming`, `energetic`, `warm`
* `social` - `quirky`, `playful`, `optimistic`
* `customer-success` - `nurturing`, `diplomatic`, `optimistic`

A personality entry is a catalog binding from its canonical name to a stable
AOS skill id (`personality-<name>`); every role reference needs one or the
loader rejects it. Bindings are not supplied bodies: the definition itself -
however presence, attention, tempo, and voice end up being expressed - lives
in the externally admitted AOS skill. A bound body may remain pending in roster
output until its source is supplied. Its format (KDL, YAML, or a metadata block
in SKILL.md) is deliberately undecided in v0.1.

## Favorite colors

A personality may declare one `color` as hex. Legibility is a parse-time
gate, not a hope: the value must sit in the terminal-legible OKLab band
(lightness 0.60-0.80, chroma at least 0.05) so it reads on dark and light
terminals alike. A single-personality composition's favorite is that color,
carried in the bundle manifest. An identity composed of several personalities
derives its favorite as the OKLab centroid of the component colors with
chroma restored to the components' minimum, clamped back into the band -
the perceptual middle, never gray.

## Agent seats

An `agent` node under a role is a named seat: the harness lineage is the key,
and `name` plus `pronouns` are the identity that seat wears. This shape is
adapted from the seat sketch in ward's `roles.kdl`, and the shared role slug
remains the join key between the two: ward keeps guardfiles, models, and
reasoning effort on its side, and nothing here grants authority. Names are
opaque strings to the engine - that a name's flavor happens to track which
model the launcher assigns to a seat is naming taste, not schema.

The embedded source carries all ten roles, including Designer for product
shaping and Customer Success for onboarding, support, retention, customer
research, and feeding recurring customer pain back into product work. Its six
named roles (engineer, director, qa, advisor, ops, pm) retain twelve seats.
Designer, Customer Success, Social, and Sales have no approved harness names.

Seats are personality-neutral: an `agent` node declares only its harness,
name, and optional pronouns. It neither selects nor defaults a personality; a
compose request selects a role and a compatible personality explicitly. The
catalog binding is embedded policy, while its externally supplied skill body
can remain visibly pending in roster output.

A private overlay may add scoped instructions or selection rules. It may not
redefine canonical roles, personalities, seats, or compatibility. AOS owns no
copy of this person source.

## See also

* [kdl-contracts.md](kdl-contracts.md) - request and source grammar.
* [architecture.md](architecture.md) - policy ownership.
