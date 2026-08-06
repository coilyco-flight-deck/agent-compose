# Claude Code native UI, design role

A hand-built reference bundle for one role, used to check that the surfaces in
[docs/claude-native-ui-surfaces.md](../../docs/claude-native-ui-surfaces.md) read
as distinct before generating all eight. Nothing here is generated yet.

## Files

* `themes/aos-design.json` - custom theme for the design role. Every override
  key was validated against the `dark` base token list shipped in Claude Code
  2.1.221, so no key is silently dropped.
* `settings.design.json` - the settings fragment that selects the theme and
  carries the spinner verbs and tips for the same role.

## Trying it

1. Copy `themes/aos-design.json` to `~/.claude/themes/aos-design.json`.
2. Merge `settings.design.json` into `~/.claude/settings.json`, or select the
   theme interactively with `/theme`.

The theme directory is watched, so an edit to the theme file applies without
restarting. Settings changes need a fresh session.

## Palette derivation

* `#ac8fd7` is the design role color. It carries the prompt border, the Clawd
  mascot body, and the assistant accent, so the role reads before any text does.
* `#b682ed` imaginative drives the skill and auto-accept accents.
* `#e882e1` playful drives the permission and bash-input accents, which is where
  the interface is asking rather than telling.
* `#4f9eb8` editorial drives suggestions and memory, the two places the
  interface is annotating the user rather than acting.

Text, background, and every diff token stay at base values. Role identity is
worth a border, never a contrast regression.

## Spinner verbs

The fragment uses `mode: "replace"` so the role voice is unambiguous while
testing. `append` is the gentler production choice, since it keeps the default
whimsy and adds the role vocabulary on top. Verbs are grouped by the meld:
refracting and conjuring from imaginative, whirling and doodling from playful,
kerning and redlining from editorial.
