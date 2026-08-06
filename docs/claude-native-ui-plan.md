# Claude Code native UI build order

How the surfaces inventoried in
[claude-native-ui-surfaces.md](claude-native-ui-surfaces.md) reach a host, and in
what order they are worth building.

## Layer ownership

Two of these surfaces are pull-based and two are push-based, and that split
decides who writes them.

* **Pull, no host mutation** - the status line and the subagent status line are
  commands. Claude Code invokes `agent-compose`, which reads the nearest
  projection and prints. Agent Compose owns these outright and nothing has to be
  installed into `~/.claude`.
* **Push, host mutation** - themes and spinner settings are files that must
  exist in `~/.claude` before Claude Code starts. By the authoring-versus-rollout
  law, Agent Compose authors the documents and convergence installs them. Agent
  Compose must not write into `~/.claude` at compose time.
* **Launch argument** - the session display name is neither. It is an argument
  the native launch path already controls.

## Build order

Cheapest and least invasive first.

1. **Session display name** - pass the resolved seat name as `--name` from the
   native launch path. No new file format, no host mutation, immediately visible
   in the prompt box and terminal title.
2. **Subagent status line** - add a `--subagent` mode to the existing
   `statusline` command that reads row context from stdin and renders the role
   mark and color per subagent row. Reuses the renderer that already exists.
3. **Theme emission** - a `theme` command that projects the selected person into
   a validated theme document, in the same shape as the existing palette
   projection. Validation against the base token list is part of the command,
   not the consumer.
4. **Spinner emission** - extend the same command family to emit the
   `spinnerVerbs` and `spinnerTipsOverride` fragments per role. Verbs come from
   the personality meld, not from the role alone.
5. **Plugin packaging** - wrap theme, output style, and syntax highlighting into
   one generated plugin per role so convergence installs a single unit rather
   than patching several settings keys.

## Open questions

* Whether spinner verbs ship in `append` or `replace` mode by default. Replace
  makes the role unmistakable and discards Claude Code's own vocabulary.
* Whether the eight fixed subagent color tokens should be remapped per role or
  left alone, given they are shared across every subagent in the session.
* Whether theme emission tracks the light bases as well, or commits to `dark`
  until someone asks.
