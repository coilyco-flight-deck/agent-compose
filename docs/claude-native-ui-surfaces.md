# Claude Code native UI surfaces

Which parts of the Claude Code terminal UI a composed identity can drive.
Verified against the installed Claude Code binary at version 2.1.221 by reading
the shipped settings schema and theme loader, not from documentation. The build
order lives in [claude-native-ui-plan.md](claude-native-ui-plan.md).

## What is available

* **Custom themes** - `~/.claude/themes/<slug>.json`, selected as
  `"theme": "custom:<slug>"`. Carries the role color into the prompt border, the
  skill and permission accents, and the Clawd mascot body.
* **Status line** - `statusLine` runs a command that receives session JSON on
  stdin. Already implemented as `agent-compose statusline`.
* **Subagent status line** - `subagentStatusLine.command` renders a per-subagent
  row in the agent panel and receives row context as JSON on stdin. Not yet
  implemented here.
* **Spinner verbs** - `spinnerVerbs` with `mode` of `append` or `replace`, plus
  a `verbs` array. The single cheapest per-role voice signal.
* **Spinner tips** - `spinnerTipsOverride` with `excludeDefault` and `tips`,
  gated by `spinnerTipsEnabled`.
* **Session display name** - the `--name` launch flag, shown in the prompt box,
  the `/resume` picker, and the terminal title. Pure launch argument, no host
  file is touched.
* **Output styles and syntax highlighting** - both are plugin components, so a
  generated plugin can carry them alongside a theme.

## What is not available

* The ASCII logo and the welcome banner. No override string exists in the
  binary. The mascot is recolorable through the `clawd_body` and `claude`
  tokens, never redrawn.
* Borders, box drawing, and component layout. Themes carry color tokens only.
* Free per-subagent color. The subagent palette is a fixed set of eight tokens
  suffixed `_FOR_SUBAGENTS_ONLY`. A role can recolor a slot but cannot add one.

## Theme file contract

The loader accepts three keys and silently drops anything else.

* `name` - display string, falls back to the file slug.
* `base` - one of `dark`, `light`, `light-daltonized`, `dark-daltonized`,
  `light-ansi`, `dark-ansi`. Anything else falls back to `dark`.
* `overrides` - a flat map. A key is kept only when the chosen base theme
  already defines that token, and a value is kept only when it matches
  `#rgb`, `#rrggbb`, `rgb(r,g,b)`, `ansi256(n)`, or `ansi:<name>`.

Silent dropping is the trap worth designing around. A misspelled token does not
error, it just does nothing, so a generator needs to validate token names
against the base theme rather than trusting its own output.

Files over 256KB are skipped with a warning. The whole directory is watched, so
a rewritten theme file is picked up without a restart.

## Safe mode caveat

`--safe-mode` disables custom themes, keybindings, output styles, and plugins
together. Every surface above except the status line command and the session
name vanishes in a troubleshooting session. Identity that must survive safe mode
belongs in the composed context, never in the theme.

## Worked example

[`examples/claude-native-ui/`](../examples/claude-native-ui/README.md) carries a
validated theme and settings fragment for the design role, built from the role
color `#ac8fd7` and its imaginative, playful, and editorial meld.
