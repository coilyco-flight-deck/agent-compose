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
  the native launch path already controls. The spinner and theme settings turned
  out to travel the same way: `--settings` loads a fragment into Claude Code's
  `flagSettings` tier, which outranks every settings file but policy. Only the
  theme document itself still needs to be pushed to the host.

## Build order

Cheapest and least invasive first.

1. **Theme emission** - done. `agent-compose native-ui` projects the selected
   person into a validated theme document per role.
2. **Spinner emission** - done. The same command emits `spinnerVerbs` from the
   personality meld, sourced from `verb` nodes in the person source.
3. **Spinner tips** - done. Each role emits its purpose, the charter lock, and
   its meld, with `excludeDefault` false so the harness keeps teaching its own
   features.
4. **Launch identity** - done. A Claude launch receives `--name <seat>` and
   `--settings <bundle>/claude-settings.json`, so the seat name, theme
   selection, verbs, and tips all arrive as arguments. No host mutation, and
   both survive a caller-supplied flag of the same name. See
   [native role launch](native-role-launch.md).
5. **Subagent status line** - done. `statusline --subagent` reads one tick of
   rows from stdin and emits `{id, content}` per row. See
   [status line](statusline.md).
6. **Plugin packaging** - wrap theme, output style, and syntax highlighting into
   one generated plugin per role so convergence installs a single unit rather
   than patching several settings keys.
7. **Host rollout** - install the theme documents onto hosts. Reduced to eight
   additive files under `<home>/.claude/themes/`, since the settings half now
   travels as a launch argument. Authored here, applied by convergence, and out
   of scope for this repo by the authoring-versus-rollout law.

## Upstream coupling

The emitted token names are Claude Code's, and the harness drops an unknown token
silently rather than erroring. `internal/nativeui` therefore emits a fixed slot
map rather than arbitrary tokens, and its test asserts every emitted token
against a vendored list of the names the harness knows. An upstream rename fails
that test instead of quietly blanking a role's identity on a user's terminal.

Drift is detected by content, not by version. The test re-extracts both vendored
lists from the installed binary on every run and skips when Claude Code is
absent, so an upgrade that changes neither list stays quiet. `ward exec
harness-refresh` rewrites them. The method and its limits live in
[harness vendoring](harness-vendoring.md).

## Open questions

* Whether spinner verbs ship in `append` or `replace` mode by default. Replace
  makes the role unmistakable and discards Claude Code's own vocabulary.
* Whether the eight fixed subagent color tokens should be remapped per role or
  left alone, given they are shared across every subagent in the session.
* Whether theme emission tracks the light bases as well, or commits to `dark`
  until someone asks.
