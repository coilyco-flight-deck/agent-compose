# agent-compose features

Inventory of what ships today.

## Composition engine

* `agent-compose compose` turns a KDL request into an immutable bundle.
* `roster:core` has 8 roles, 16 personalities, balanced triads, and stable
  identities. [Role briefings](role-briefings.md) own each charter.
* [External person packages](person-packages.md) and
  [local personality libraries](personality-libraries.md) replace the default.
* `.agents/roles.kdl` owns [skill-provider repos](role-scoped-providers.md), skills, and [repository policy](repository-policy.md).
* [Three model tiers](model-tiers.md) declare Core role deployment
  compatibility without changing selected context.
* Materialization promotes admitted `COMPOSED.md` to native `SKILL.md`.
* Resolver traces provider and content outcomes with budgets.
* Atomic materialization verifies staged and reused bundles.
* Canonical skills use identity cards and compiled fallback. Role bodies cap at
  400 words, [role boundaries](role-boundaries.md) under a separate one.
* [Role adjacency](role-adjacency.md) names each role's two likeliest absorptions.
* [evalkit](eval-orchestration.md) runs the board, `aos-eval` grades it.

## Load-point projection

* `agent-compose project` places verified bundles transactionally at repo or
  container-home load points for four harnesses.
* Sidecar ownership protects foreign files and restores prior owned state.

## Launch-time refresh

* [Launch](native-role-launch.md) adds color, an Enter gate, and a Codex intro.
* A Claude launch passes [identity flags](claude-launch-identity.md), not files.
* Refresh uses validated fallback unless `external-only` forbids it.

## Inspection

* `config validate` strictly checks staged host configuration without writes.
* `agent-compose describe` renders a collapsible decision tree. `--why`
  follows one item from consideration to outcome.
* `agent-compose diff` reports semantic changes. `verify` checks entry points,
  delivery, traces, and selected identities.
* [Catalogues and export](catalogues-and-export.md) provide inspection,
  reproducible archives, and logical content diff.
* `compose` renders complete role metadata. `--explain` adds decisions.
* [Evaluation](evaluation.md) derives the board from the roster and grades it
  by hand, with no mechanical scorer in the loop.
* [V2 migration](v2-migration.md) maps v1 roles without aliases.
* TTY colors use canonical identity and pass an OKLab legibility gate.
  Redirects and `NO_COLOR` stay plain.

## Identity surfaces

* [Identity renderers](statusline.md) cover the palette, overlays, the
  `acompose statusline` row, `--subagent` rows, and the
  [short id](short-id.md). [`whoami`](whoami.md) prints it.
* [`native-ui`](claude-native-ui-surfaces.md) emits per-role Claude Code themes.

## Roster artifact and cascade

* `agent-compose roster --out <dir>` renders lazy-loaded role and personality
  skills plus [native adaptation](native-adaptation.md).
* Bare convergence emits deterministic [`person.json`](person-snapshot.md).
* `cascade` emits harness doctrine and the role/residency
  `repository-plan.yaml` with sealed provenance.
* `bundle materialize` returns a verified role/harness bundle with provenance.
* Bare `acompose` converges hosts. `--reapply` forces the layout, `--verbose`
  traces `source => destination`, and `-- <command>` refreshes then execs.
  Ward smoke proves idempotence.
* [Local skill catalogues](local-skill-catalogues.md) consume AOS roots.
* [Release](release.md) publishes unreleased product deltas, including
  roll-forward recovery, under a hold.

## See also

* [../README.md](../README.md) - product boundary and current status.
* [../AGENTS.md](../AGENTS.md) - repo-specific operating rules.
* [../justfile](../justfile) - development recipes.
* [../.ward/ward.yaml](../.ward/ward.yaml) - catalog metadata.
* [Catalog trifecta](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/src/branch/main/docs/features-release-tooling.md).
