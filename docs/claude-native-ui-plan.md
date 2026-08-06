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

1. **Theme emission** - done. `agent-compose native-ui` projects the selected
   person into a validated theme document per role.
2. **Spinner emission** - done. The same command emits `spinnerVerbs` from the
   personality meld, sourced from `verb` nodes in the person source.
3. **Session display name** - pass the resolved seat name as `--name` from the
   native launch path. No new file format, no host mutation, immediately visible
   in the prompt box and terminal title.
4. **Subagent status line** - add a `--subagent` mode to the existing
   `statusline` command that reads row context from stdin and renders the role
   mark and color per subagent row. Reuses the renderer that already exists.
5. **Spinner tips** - emit `spinnerTipsOverride` per role. Held back from the
   first pass because tips are prose, and role prose has a review path that
   colors and verbs do not.
6. **Plugin packaging** - wrap theme, output style, and syntax highlighting into
   one generated plugin per role so convergence installs a single unit rather
   than patching several settings keys.

## Upstream coupling

The emitted token names are Claude Code's, and the harness drops an unknown
token silently rather than erroring. `internal/nativeui` therefore emits a fixed
slot map rather than arbitrary tokens, and its test asserts every emitted token
against the dark base set shipped in Claude Code 2.1.221. An upstream rename
fails that test instead of quietly blanking a role's identity on a user's
terminal. Refreshing the list is a deliberate act when the harness moves.

## Open questions

* Whether spinner verbs ship in `append` or `replace` mode by default. Replace
  makes the role unmistakable and discards Claude Code's own vocabulary.
* Whether the eight fixed subagent color tokens should be remapped per role or
  left alone, given they are shared across every subagent in the session.
* Whether theme emission tracks the light bases as well, or commits to `dark`
  until someone asks.
