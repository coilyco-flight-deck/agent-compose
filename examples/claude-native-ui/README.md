# Claude Code native UI, generated

Generated output, not hand-authored. Regenerate with:

```
agent-compose native-ui --out examples/claude-native-ui
```

## Files

* `themes/aos-<role>.json` - one custom theme per role, ready to copy into
  `~/.claude/themes/`.
* `settings.<role>.json` - the settings fragment that selects that theme and
  carries the role's spinner verbs.

## Trying one

1. Copy `themes/aos-design.json` to `~/.claude/themes/aos-design.json`.
2. Select it with `/theme`, or merge `settings.design.json` into
   `~/.claude/settings.json`.

The theme directory is watched, so an edit to a theme file applies without
restarting. Settings changes need a fresh session.

## How a role becomes a theme

The role color is the OKLab meld of its personality colors, the same math the
palette explorer and status line already use. It carries the frame: the prompt
border, the Clawd mascot body, and the assistant accent.

Each personality in the meld then carries one interaction the frame contains.

* The first carries what the agent offers, so it colors skills and auto-accept.
* The middle carries what the agent asks for, so it colors permissions and the
  bash input border.
* The last carries what the agent annotates back, so it colors suggestions and
  memory.

Text, background, and every diff and severity token stay at base values. Role
identity is worth a border, never a contrast regression, and a test enforces
that.

Each role also claims the nearest of Claude Code's eight fixed subagent color
slots. Two roles can land on the same slot, which is harmless because only one
role is active per session, but it does mean the slot is not a role identifier.

## Spinner verbs

Verbs live on the personality in the person source, six each, so a role's
spinner vocabulary is the concatenation of its meld in role order. Design gets
refracting and conjuring from imaginative, whirling and doodling from playful,
kerning and redlining from editorial.

`replace` is the default, which discards Claude Code's own vocabulary and makes
the role unmistakable. Pass `--spinner-mode append` to keep both. No authored
verb repeats one of the harness's own 184, so replace mode always reads as the
role, and a test holds that line.

## Spinner tips

Three per role: the purpose, the charter lock naming the seat, and the meld. A
tip lands while the reader is waiting rather than reading, so it carries
doctrine and not voice. `excludeDefault` stays false, because Claude Code's own
tips teach the harness and a role has no business hiding them.
