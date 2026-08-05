# Native role launch

`acompose <role> <harness>` starts a native host harness with one
caller-assigned role bundle:

```sh
acompose design codex
acompose engineer claude --model opus
acompose qa goose run
acompose content opencode
```

Arguments after the harness pass through unchanged. The long form is
`agent-compose launch <role> <harness>`.

## Selection

The launcher converges the host first, then reads the generated
`mount-eligibility.json`. It admits ordered default, harness, and
[role-scoped providers](role-scoped-providers.md). Required providers fail when
missing. Optional exclusions stay traced. Isolated workspaces remap the same set.

Model tier and bundle model class are separate launch-consumer runtime facts.
Agent Compose defaults both to `frontier`. A launch consumer may set
`AGENT_COMPOSE_MODEL_TIER=frontier`, `commodity`, or `oss` to select the role
compatibility lane, and `AGENT_COMPOSE_MODEL_CLASS=frontier` or `low-context`
for context density. AOS owns the runtime registry for both. All harnesses
receive native skills through their existing projection layout.

The resulting bundle contains only the assigned role skill, its complete
ordered personality meld, ordinary admitted skills, and role-composed skills
bound to that role. Startup instructions require the harness to read the role
and meld skills before acting. Another role requires another launch.

Before the harness starts, the launcher prints routine composition status at
normal speed, then renders the canonical role transcript as the final
substantive block. When both input and output belong to an interactive terminal,
`Press Enter to continue` keeps that identity visible until acknowledgement.
Enter starts the harness. Ctrl-C cancels before launch.

Piped, redirected, and headless launches retain the non-interactive flow. They
never read stdin or wait for acknowledgement. TTY output uses the melded role
color and each personality's own color. Redirected output and `NO_COLOR` remain
plain.

Bare interactive Codex launches also supply an initial prompt asking the active
Codex seat to introduce itself from the loaded identity card and personality
meld, then invite the user's task. Codex options such as AOS's workspace trust
override or an explicit model selection may precede that prompt. An explicit
positional prompt, subcommand, or unknown option passes through without an
added prompt.

## Native workspace integration

Agent Compose owns selection and projection. A launch consumer owns workspace
isolation and process lifecycle. AOS's shared shell wrapper places the explicit
role launch inside its leased native workspace and supplies a shadow home
before invoking `agent-compose launch`. The shadow preserves native host state
but omits the host user-skill mount, so inactive role and personality skills do
not re-enter the session. System and plugin skills remain harness-owned. No
container is involved. When the shadow links an existing Codex state directory,
Agent Compose resolves that link before setting `CODEX_HOME`. Codex therefore
keeps one canonical identity for persisted hook trust while `HOME` remains
isolated to the session.

The direct long form projects into the current directory using Agent Compose's
transactional sidecar rules. A consumer that permits concurrent sessions
should provide a distinct current directory for each launch. Projection fails
before starting the harness when foreign load-point files occupy that target.
Without a launch-consumer shadow home, existing user-scoped skills remain
visible according to the harness's normal discovery rules.

Bare `acompose` still converges the host. `acompose -- <command>` retains the
inferred-role native path for compatibility.

## See also

* [Integration](integration.md) - host and isolated delivery tiers.
* [Role selection](role-selection.md) - inferred and caller-assigned roles.
* [Projection](projection.md) - harness load points and ownership.
