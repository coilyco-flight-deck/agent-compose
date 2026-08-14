# Naming the seat

A compose request may rename the seat it composes:

```kdl
compose {
    role "ops"
    identity name="Echo" pronouns="it"
    delivery "native-skills"
}
```

Both properties are required and the node takes no arguments. Omitting it keeps
the role's own seat, which is what every existing request does.

## Why it exists

A caller that already has an identity would otherwise carry two. Sirens Echo
composes `ops` for its operator doctrine and answers as Sirens Echo, so without
this its prompt introduced Olaf as well — a name belonging to a different
context, in a lane whose policy forbids describing itself at all.

The alternative was a whole person package, which replaces the embedded roster
as one unit. Copying eight roles to rename one seat makes every future roster
change something the copy has to chase.

## What it does not do

It renames. Skills, methods, personalities, boundaries, and model tiers are
untouched, and no equivalent exists for any of them.

The personality meld is the one people ask for next, and it is a different
shape. `role.Personalities` also drives the melded favorite color, the nativeui
theme tokens, and a validator requiring exactly three personalities on a core
role. Filtering a meld is several seams; a name is one.

## The invariant this sits beside

[The person contract](person-contract.md) says an overlay may not redefine
selected roles, personalities, definitions, or role personality sets. This is
the single exception, and the line it holds is between **who is speaking** and
**what the role is**. Naming a seat is identity. A caller that wants different
role content still brings its own package, as one unit.

## What moves together

The override rewrites the role identity and every seat, because different
consumers read different ones: the identity card and roster read the role, while
the overlay, statusline, native launcher, and bundle manifest read the seats.
Rewriting one alone yields a bundle whose card and statusline disagree.

The bundle key moves with it, because `cacheKey` hashes the rendered
instructions and the card is in them. A renamed seat cannot reuse an unrenamed
bundle.
