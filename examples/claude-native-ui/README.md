# Claude Code native UI, generated

Generated output, not hand-authored. Regenerate with:

```
agent-compose native-ui --out examples/claude-native-ui
```

## Files

* `themes/aos-<role>.json` - one custom theme per role, ready to copy into
  `~/.claude/themes/`.
* `settings.<role>.json` - the settings fragment that selects that theme and
  carries the role's spinner verbs, tips, and subagent status line.

A Claude launch passes the fragment as `--settings`, so a composed session needs
only the theme file installed. See
[claude-launch-identity.md](../../docs/claude-launch-identity.md).

## Trying one

1. Copy `themes/aos-frontend.json` to `~/.claude/themes/aos-frontend.json`.
2. Select it with `/theme`, or merge `settings.frontend.json` into
   `~/.claude/settings.json`.

The theme directory is watched, so an edit to a theme file applies without
restarting. Settings changes need a fresh session.

## How a role becomes a theme

The role color is the OKLab boundary of its personality colors, the same math the
palette explorer and status line already use. It carries the frame: the prompt
border, the Clawd mascot body, and the assistant accent.

Each personality in the boundary then carries one interaction the frame contains.

* The signature carries what the agent offers, so it colors skills and
  auto-accept.
* The bond carries what the agent asks for and annotates back, so it colors
  permissions, the bash input border, suggestions, and memory.

Text, background, and every diff and severity token stay at base values. Role
identity is worth a border, never a contrast regression, and a test enforces
that.

Each role also claims the nearest of Claude Code's eight fixed subagent color
slots. Two roles can land on the same slot, which is harmless because only one
role is active per session, but it does mean the slot is not a role identifier.
The `subagentStatusLine` row therefore carries the identity as text.

## Spinner verbs

Verbs live on the personality in the person source, six each, so a role's
spinner vocabulary is the concatenation of its meld in role order. Frontend gets
whirling and doodling from playful, then refracting and conjuring from
imaginative.

`replace` is the default, which discards Claude Code's own vocabulary and makes
the role unmistakable. Pass `--spinner-mode append` to keep both. No authored
verb repeats one of the harness's own defaults, so replace mode always reads as
the role, and a test against the
[vendored list](../../docs/harness-vendoring.md) holds that line.

## Spinner tips

Three per role: the purpose, the charter lock naming the seat, and the boundary. A
tip lands while the reader is waiting rather than reading, so it carries
doctrine and not voice. `excludeDefault` stays false, because Claude Code's own
tips teach the harness and a role has no business hiding them.
