# agent-compose features

Inventory of what ships today.

## Composition engine

* `agent-compose compose` turns a KDL request into an immutable offline bundle.
* `roster:core` carries eight roles, sixteen personalities, melds, and seats.
  Content exclusively owns human communication recommendations. Every other
  role stops at a factual handoff, and external delivery remains separately
  authorized.
* [External person packages](person-packages.md) and
  [local personality libraries](personality-libraries.md) replace the default
  across every person-dependent command.
* `.agents/roles.kdl` owns [skill-provider repos](role-scoped-providers.md), skills, and [repository policy](repository-policy.md).
* [Three model tiers](model-tiers.md) fail closed through the Core role matrix. Tier stays separate from bundle density.
* Materialization promotes admitted `COMPOSED.md` files to native `SKILL.md`.
* Resolver traces provider and content outcomes with context budgets.
* Atomic materialization verifies staged and reused bundles.
* Canonical skills use compact identity cards and compiled fallback. Designer
  may land bounded pages with supplied static copy and semantic markup. Content
  may land content-only code changes.

## Load-point projection

* `agent-compose project` places verified bundles transactionally at repo or
  container-home load points for all four harnesses.
* Sidecar ownership protects foreign files and restores prior owned state.

## Launch-time refresh

* [Launch](native-role-launch.md) adds color, an Enter gate, and a Codex intro.
* Refresh uses validated fallback unless `external-only` forbids it.

## Inspection

* `config validate` strictly checks staged host configuration without writes.
* `agent-compose describe` renders a collapsible decision tree. `--why`
  follows one item from consideration to outcome.
* `agent-compose diff` reports semantic changes. `verify` checks entry points,
  delivery, trace integrity, and selected identities.
* [Catalogues and export](catalogues-and-export.md) provide rich inspection,
  reproducible archives, and logical content diff.
* `compose` renders complete role metadata. `--explain` adds decisions.
* [Evaluation](evaluation.md) has three-lane Core Roster matrices, explicit
  communication hard fails, independent-review evidence, and
  [scorecards](evaluation-scorecard.md). `disabled_model_tiers` pauses a lane
  without deleting its matrix.
* [V2 migration](v2-migration.md) maps v1 roles without aliases.
* TTY colors use canonical identity. Redirects and `NO_COLOR` stay plain.
* Colors pass an OKLab legibility gate.

## Identity surfaces

* [Identity renderers](statusline.md) include the local palette, text or JSON
  overlays, and an `acompose statusline` row for active bundle identity and health.

## Roster artifact and cascade

* `agent-compose roster --out <dir>` renders lazy-loaded role and personality
  skills plus [native adaptation](native-adaptation.md).
* Bare convergence emits deterministic [`person.json`](person-snapshot.md).
* `cascade` emits harness doctrine and the role/residency `repository-plan.yaml`
  with sealed policy-source provenance.
* `bundle materialize` returns a verified role/harness bundle with provenance.
* Bare `acompose` converges hosts. `--reapply` forces the layout, `--verbose`
  traces `source => destination`, and `-- <command>` refreshes then execs.
  Ward smoke proves idempotence and its test verb runs full validation.
* [Local skill catalogues](local-skill-catalogues.md) consume AOS roots.

* [Release](release.md) publishes product-impacting main pushes under a major-version hold.

## See also

* [../README.md](../README.md) - product boundary and current status.
* [../AGENTS.md](../AGENTS.md) - repo-specific operating rules.
* [../.ward/ward.yaml](../.ward/ward.yaml) - allowlisted development commands.
* [Catalog trifecta](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/src/branch/main/docs/features-release-tooling.md) - shared structure.
