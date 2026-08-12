# Boundary owners

Every boundary names the role holding the other side with a required `owner`.
A boundary without one is a roster-wide rule that some roles care about more
than others, which belongs in the layer that owns rules rather than here.

The owner is a relationship, not authority. It grants no permission and no
executable capability. It does two mechanical things: the owner receives the
body without declaring it, and loading fails when an owner declares its own
boundary, since that would place one role on both sides.

## Two-sided bodies

One body carries both halves under conditional headings, so the reader
self-selects before reading a word of prose:

```markdown
# Boundary: modify live system

Who changes running systems, and who hands that change to the role that owns it.

## If you own this boundary

You own live system modification...

## If you defer this boundary

Your clone is sealed against live mutation...
```

The owner section comes first, so the deferral reads as the consequence of the
allocation rather than as a bare prohibition. Both sections are required, each
is bounded separately at 400 words, and the whole file goes to both sides so
each role can read what the other was told.

Nothing parses these headings. The roster already records who owns and who
declares, so delivery, the identity card, and evaluation coverage all key off
that. Headings exist for the glance.

Sections identify by relationship, never by role name. A role list in prose
beside the same list in KDL is the drift this design removes.

## See also

* [Role boundaries](role-boundaries.md) - the primitive and its budget.
* [Evaluation](evaluation.md) - the coverage each side owes.
* [Person package authoring](person-package-authoring.md) - complete layout.
