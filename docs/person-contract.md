# Embedded person contract

The binary embeds one ordered `person` source set for v0.1, without a provider interface:

```kdl
person "kai" {
    role "engineer" {
        purpose "Write code, merge code, stay focused on your goal."
        briefing "..."
        personality "curious" "grounded" "meticulous"
        agent "claude" name="opal engineer" pronouns="she"
        agent "codex" name="terran engineer" pronouns="he"
    }
    personality "curious" skill="personality-curious" color="#d98e48" motif="map-paper"
}
```

A role names its concise purpose, required long-form briefing, the two or
three personalities it wears together, and its named agent seats. `purpose`
is the short label used in headings and summaries. `briefing` is the
unconditional role charter described in [role-briefings.md](role-briefings.md),
with at least three substantial paragraphs. The loader also rejects any other
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
* `community` - `nurturing`, `diplomatic`, `playful`
* `customer-success` - `nurturing`, `diplomatic`, `optimistic`

A personality entry is a catalog binding from its canonical name to a stable
skill id (`personality-<name>`). Every role reference needs one or the loader
rejects it. The same embedded `person:kai` source supplies one complete
`SKILL.md` tree for every binding plus the personality invariant. A missing,
empty, extra, or mismatched canonical definition fails source validation.
Roster output therefore carries every canonical definition without an
external provider.

The [identity primitives](identity-primitives.md) define renderer semantics. The [inspiration catalogue](inspiration-catalogue.md) defines credits.

## Favorite colors

Every canonical personality declares one hex `color` in the embedded source,
which owns the exact palette. Bundles tell the agent every active personality's
name, skill, and color plus the melded favorite. The parse gate requires OKLab
lightness 0.60-0.80 and chroma of at least 0.05. Each role derives its favorite
as the OKLab centroid of every component, restores chroma to their minimum, and
clamps it into the legible band - the perceptual middle, never gray.

## Agent seats

An `agent` node is a named seat. The harness is its join key, while `name` and
`pronouns` are the identity it wears. Ward keeps guardfiles, models, and
reasoning effort on its side. Nothing here grants authority.

All eleven roles keep Claude she/her and Codex he/him seats. Fourteen AOSH-selected
public seats use grep-friendly `placeholder ...` names and they/them pronouns. Discord adds one she/her community host, for 37 canonical seats total.

Seats are personality-neutral. A compose request selects a role, and that role
activates its ordered personality set. Roster briefing delivery is documented
in [role-briefings.md](role-briefings.md).

A private overlay may add scoped instructions or selection rules. It may not
redefine canonical roles, personalities, seats, definitions, or role
personality sets. AOS owns no copy of this person source or its personality
definitions.

## See also

* [person-snapshot.md](person-snapshot.md) - complete machine-readable export.
* [role-briefings.md](role-briefings.md) - role charter and delivery contract.
* [kdl-contracts.md](kdl-contracts.md) - request and source grammar.
* [architecture.md](architecture.md) - policy ownership.
