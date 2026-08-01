# agent-compose features

Inventory of what ships today. Planned behavior lives in the issue tracker.

## Composition engine

* `agent-compose compose` turns a KDL request into an immutable offline bundle.
* `roster:core` carries eight roles, sixteen referenced personalities, display
  names, melds, and seats grounded in Kai's real projects.
* [External person packages](person-packages.md) and
  [local personality libraries](personality-libraries.md) replace the default
  across every person-dependent command.
* AOS roots expose ordinary and [role-scoped skills](role-scoped-providers.md).
  `.agents/roles.kdl` owns wildcard admission.
* Role compatibility fails closed. Portfolio Strategist supports only
  `frontier` until the v2 OSS evaluation admits it.
* Materialization promotes admitted `COMPOSED.md` files to native `SKILL.md`.
* Resolver traces provider and content outcomes with context budgets.
* Atomic materialization verifies staged and reused bundles.
* Canonical skills use compact identity cards and compiled fallback. Designer
  may land bounded page-level web experiences. Content may land content-only
  code changes.
* Historical scorecard mode renders preserved v1 results without rebinding.

## Load-point projection

* `agent-compose project` places verified bundles transactionally at repo or
  container-home load points for all four harnesses.
* Sidecar ownership protects foreign files and restores prior owned state.

## Launch-time refresh

* [Launch](native-role-launch.md) adds color, an Enter gate, and a Codex intro.
* Refresh uses validated fallback unless `external-only` forbids it.

## Inspection

* `agent-compose describe` renders a collapsible decision tree. `--why`
  follows one item from consideration to outcome.
* `agent-compose diff` reports semantic changes. `verify` checks entry points,
  delivery, trace integrity, and selected identities.
* [Catalogues and export](catalogues-and-export.md) provide rich inspection,
  reproducible archives, and logical content diff.
* `compose` renders complete role metadata. `--explain` adds decisions.
* [Evaluation](evaluation.md) has paired Core Roster scenario matrices,
  independent-review evidence, and [scorecards](evaluation-scorecard.md).
* [V2 migration](v2-migration.md) maps v1 roles without aliases.
* TTY colors use canonical identity. Redirects and `NO_COLOR` stay plain.
* Colors pass an OKLab legibility gate. Each role carries the chroma-restored centroid.

## Identity surfaces

* The local [palette](personality-palette.md) shows melds and identity primitives.
  `agent-compose overlay` emits text or JSON for one selected member. See [overlay.md](overlay.md).

## Roster artifact and cascade

* `agent-compose roster --out <dir>` renders lazy-loaded role and personality
  skills plus [native adaptation](native-adaptation.md).
* Bare convergence emits deterministic [`person.json`](person-snapshot.md).
* `agent-compose cascade` is the absorbed v1 composer: doctrine sources into
  per-harness files, symlinks, filtering, overrides, a mount manifest,
  and dry-run/check behavior compatible with the Python outputs.
* Bare `acompose` converges hosts. `--reapply` forces the layout, `--verbose`
  traces `source => destination`, and `-- <command>` refreshes then execs.
  Ward smoke proves idempotence and its test verb runs full validation.
* [Local skill catalogues](local-skill-catalogues.md) consume AOS roots.

## Release

* [Release](release.md) publishes product-impacting main pushes. Docs and
  results only validate. A tracked hold reserves major versions for dispatch.

## See also

* [../README.md](../README.md) - product boundary and current status.
* [../AGENTS.md](../AGENTS.md) - repo-specific operating rules.
* [../.ward/ward.yaml](../.ward/ward.yaml) - allowlisted development commands.
* [Catalog trifecta convention](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/src/branch/main/docs/features-release-tooling.md) - shared structure.
