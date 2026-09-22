# Claude launch identity

A [Claude launch](native-role-launch.md) receives its native identity as launch
arguments, so no file lands in the host's `~/.claude` and nothing has to be
converged before the session starts.

## The two flags

* `--name <annotation>` carries the resolved seat annotation into the prompt
  box, the `/resume` picker, and the terminal title. The selected seat's own
  name wins, then the role-owned agent identity, then the role.

  ```text
  Angie (Platform Engineer)
  ```

  The annotation is the seat name and the role's authored `display-name` in
  parentheses. A person package that omits the display name drops that part
  rather than rendering an empty pair of parens, so an external package still
  launches.
* `--settings <bundle>/claude-settings.json` carries the role's
  [native UI](claude-native-ui-surfaces.md) fragment: the theme selection, the
  spinner verbs, and the spinner tips. Refresh writes that fragment beside the
  bundle it just composed, and only for the `claude` harness.

## Why arguments and not files

Claude Code resolves settings through an ordered set of tiers:

```
userSettings  projectSettings  localSettings  flagSettings  policySettings
```

`--settings` loads into `flagSettings`, which outranks the user, project, and
local settings files and loses only to policy. So one argument delivers the
whole settings half of a role's native UI, and the launch path never writes into
host state to dress a session.

Selection through that tier is observed, not inferred. A session launched with
the platform fragment renders the platform theme even when the user settings file
selects a different custom theme.

## Caller precedence

A caller-supplied `--name` or `--settings` always wins, in either the separate
value or the inline `--flag=value` spelling. The launcher puts its own flags
ahead of the caller's arguments, and treats everything after a `--` terminator
as harness input rather than as a flag it may match.

Other harnesses receive no flags. They still resolve a seat name, because a name
is identity rather than a file, but only Claude Code reads a settings fragment as
an argument.

## What still needs the host

The theme document the fragment names has to exist under
`<home>/.claude/themes/`, which convergence installs. Claude Code drops an
unresolvable `custom:` reference silently, so a host without the theme files
loses the colors and keeps the name, the verbs, and the tips.

`--safe-mode` disables custom themes along with plugins, output styles, and
keybindings. The session name survives it.

## Token metrics

A Claude seat exports Claude Code's `claude_code.token.usage` metrics, labelled
by seat, once the host configuration names a collector
(`teable:coilyco-flight-deck/agent-compose#7834`). The fleet rollout renders the
endpoint, and Agent Compose ships no default.

```yaml
telemetry:
  otlp_metrics_endpoint: http://collector.example:4318/v1/metrics
  protocol: http/protobuf   # optional, also http/json or grpc
```

* It sets `CLAUDE_CODE_ENABLE_TELEMETRY=1`, `OTEL_METRICS_EXPORTER=otlp`, and
  only the per-signal metrics endpoint and protocol, so no other signal is routed.
* Logs and traces are forced to `none`, and the prompt, tool, and raw-body
  logging flags are cleared, because events can carry prompt text.
* `OTEL_RESOURCE_ATTRIBUTES=seat=<role slug>,shadow=<AOS_NATIVE_SESSION>`. It
  never uses the display name, which concurrent sessions in one shadow share,
  and `session.id` separates those sessions.
* `AGENT_COMPOSE_TELEMETRY=off`, or no `telemetry` block, clears every managed
  variable. The session then emits nothing even when a parent seat exported,
  and the switch carries on to seats that session launches.

## See also

* [Native role launch](native-role-launch.md) - selection and the launch flow.
* [Native UI surfaces](claude-native-ui-surfaces.md) - what each role emits.
* [Build order](claude-native-ui-plan.md) - which surfaces ship and in what order.
