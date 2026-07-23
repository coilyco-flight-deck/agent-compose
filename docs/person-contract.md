# Embedded person contract

The binary embeds exactly one `person` source for v0.1. No provider interface
or trait-slider abstraction sits in front of it:

```kdl
person "kai" {
    role "engineer" {
        purpose "Write code, merge code, stay focused on your goal."
        briefing "..."
        personality "curious" "grounded" "meticulous"
        agent "claude" name="opal engineer" pronouns="she"
        agent "codex" name="terran engineer" pronouns="he"
    }
    personality "curious" skill="personality-curious" color="#d98e48"
}
```

A role names its concise purpose, required long-form briefing, the two or
three personalities it wears together, and its named agent seats. `purpose`
is the short label used in headings and summaries. `briefing` is the
unconditional role charter described in
[role-briefings.md](role-briefings.md). The loader also rejects any other
personality cardinality or a repeated personality. The embedded roster
contains this approved meld matrix:

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
AOS skill id (`personality-<name>`). Every role reference needs one or the
loader rejects it. Bindings are not supplied bodies. The definition lives in
the externally admitted AOS skill and may remain pending in roster output.

## Favorite colors

Every canonical personality declares one hex `color` in the embedded person
source, which owns the exact palette. The parse-time gate requires OKLab
lightness 0.60-0.80 and chroma of at least 0.05, keeping colors readable on dark
and light terminals. Every role composition derives its favorite from all
active personalities as the OKLab centroid of the component colors, restores
chroma to the components' minimum, and clamps it into the band - the perceptual
middle, never gray.

## Agent seats

An `agent` node is a named seat. The harness is its join key, while `name` and
`pronouns` are the identity it wears. Ward keeps guardfiles, models, and
reasoning effort on its side. Nothing here grants authority.

The six named roles (engineer, director, qa, advisor, ops, pm) retain twelve
seats. Designer, Customer Success, Social, and Sales remain canonical but have
no approved harness names.

Seats are personality-neutral. A compose request selects a role, and that role
activates its ordered personality set. Roster briefing delivery is documented
in [role-briefings.md](role-briefings.md).

A private overlay may add scoped instructions or selection rules. It may not
redefine canonical roles, personalities, seats, or role personality sets. AOS owns no
copy of this person source.

## See also

* [role-briefings.md](role-briefings.md) - role charter and delivery contract.
* [kdl-contracts.md](kdl-contracts.md) - request and source grammar.
* [architecture.md](architecture.md) - policy ownership.
