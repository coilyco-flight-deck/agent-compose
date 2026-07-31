# Staged-home handoff

A composition adapter can turn one immutable bundle into a launcher-neutral
home tree without teaching agent-compose the launcher's manifest, container,
permission, or lifecycle model.

The adapter starts with an empty private directory and runs:

```text
agent-compose verify <bundle>
agent-compose project <bundle> \
  --layout <agent> \
  --scope home \
  --target <empty-home>
```

The selected agent controls the only accepted load points:

* `claude` - `.claude/CLAUDE.md` and optional `.claude/skills/`
* `codex` - `.codex/AGENTS.md` and optional `.agents/skills/`
* `goose` - `.config/goose/.goosehints` and optional `.agents/skills/`
* `opencode` - `.config/opencode/AGENTS.md` and optional `.agents/skills/`

Agent-compose also writes `.agent-compose/` projection state. That directory
can contain `projection.json` and a platform lock file. It is not agent
context. After `project` returns and no other process can use the private
staging directory, an adapter removes `.agent-compose/`, validates that only
the selected load points remain, and wraps those files in its own handoff
schema.

The adapter owns that new schema. Agent-compose never parses or emits it.
Agent-compose also never starts a container, selects runtime authority, mounts
a tool, or invokes a launch consumer.

The role in the source bundle selects context. A matching role slug in another
system does not transfer permissions or merge authority into the projected
home.
When the bundle was composed from
[role-scoped providers](role-scoped-providers.md), home projection preserves
the same selected skill inventory as native projection. Projection does not
re-resolve providers or mutate the immutable input bundle.

The generic projection remains useful on its own through `agent-compose` and
the `acompose` host entrypoint. No composition root is required for native use.

Cross-repository orchestration is tracked in
[inbox#267](https://forgejo.coilysiren.me/coilysiren/inbox/issues/267). This
producer-side boundary is tracked in
[agent-compose#103](https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/issues/103).

## See also

* [projection.md](projection.md) - transactional ownership and load points.
* [bundle-protocol.md](bundle-protocol.md) - immutable source bundle.
* [integration.md](integration.md) - native and isolated delivery tiers.
* [role-scoped-providers.md](role-scoped-providers.md) - shared native and staged selection.
