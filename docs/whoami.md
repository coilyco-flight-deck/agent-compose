# What this session calls itself

`acompose whoami` prints the composed name for the projection at `--target`:

```text
Angie [she] uz86
```

Nothing else on stdout, so a shell hook can use it without parsing. Silence
means no projection applies.

## Why it exists

A SessionStart hook has to tell an agent what it is called. Before this, the
caller computed a name of its own - harness, OS, host, a tag sliced out of the
raw session UUID - which produced a second naming scheme that agreed with
nothing. It could not know the composed seat, so an agent introduced itself as
one thing and its status line said another.

`whoami` makes the composition the single answer. It is the same seat label the
[status line](statusline.md) renders, from the same bundle manifest, so the two
surfaces cannot disagree.

## Silence rather than a guess

Outside a projection there is no composed name, and `whoami` prints nothing.

That is deliberate. A session with no composition genuinely has no composed
identity, and synthesising one is exactly what the retired local computation did
wrong - it always produced a name, so a caller could never tell a real identity
from a fabricated one. A hook that wants a fallback can supply its own, knowing
it is a fallback.

## What it carries

* The seat name and subject pronoun from the selected bundle.
* The session [short id](short-id.md) when one is in scope.

It does **not** carry the role. The status-line row already names the role as
`role@harness`, and the bundle manifest stores the role slug rather than a
display name, so a paren here would either restate the slug or need a bundle
format change for cosmetics.

## See also

* [Projected composition status line](statusline.md) - the row that shares this name.
* [The dictatable short id](short-id.md) - the trailing four characters.
