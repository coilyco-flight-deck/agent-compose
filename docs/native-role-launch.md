# Native role launch

`acompose <role> <harness>` starts a native host harness with one
caller-assigned role bundle:

```sh
acompose designer codex
acompose engineer claude --model opus
acompose qa goose run
acompose social opencode
```

Arguments after the harness pass through unchanged. The long form is
`agent-compose launch <role> <harness>`.

## Selection

The launcher converges the host first, then reads the generated
`mount-eligibility.json`. It admits the selected harness's eligible local
providers and requires at least one provider with role-composed bindings.
When a launch consumer has reproduced the configured repositories in an
isolated workspace, Agent Compose remaps the provider paths to that workspace.

Model class is a launch-consumer runtime fact. Agent Compose defaults to the
full `frontier` bundle. A launch consumer may set
`AGENT_COMPOSE_MODEL_CLASS=frontier` or `low-context` before exec. AOS supplies
that value from its own layout registry. All four harnesses receive
native-skills delivery through their existing projection layout.

The resulting bundle contains only the assigned role skill, its complete
ordered personality meld, ordinary admitted skills, and role-composed skills
bound to that role. Startup instructions require the harness to read the role
and meld skills before acting. Another role requires another launch.

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
