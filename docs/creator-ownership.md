# Content Creator ownership

Agent Compose keeps executable parsing, validation, selection, deterministic
rendering, and diagnostics in Go. Product prose and profile policy belong to
reviewable local data assets.

## Boundaries

* Engine assets - generic evaluation behavior and the native adaptation
  policy. `internal/roster/definitions/NATIVE-ADAPTATION.txt` is embedded data,
  not a profile override.
* Profile assets - role skills, role-bound methods, structured role metadata,
  role identity, invariant, copy contracts, and optional complete
  `evaluations/<role>.yaml` matrices.
* Personality-library assets - personality bindings, aliases, identity
  primitives, and definition skills.
* Consumer configuration - local profile and library roots only. Agent
  Compose does not fetch URLs, clone repositories, resolve releases, or read
  git references.

## Evaluation matrices

The engine supplies its complete generic matrix when a selected role has no
profile asset. A profile matrix replaces that matrix as one complete unit. The
loader does not merge fields, and a role cannot silently opt out.

Content Creator owns one connected audience loop: reusable source artifacts,
provenance, claim discipline, editorial recommendations, audience and contact
research within accepted strategy, channel adaptation, community continuity,
durable feedback, qualification, discovery support, evidence selection, and
decision records. These responsibilities no longer cross artificial role
handoffs.

Content Creator owns recommendations about human communication, not every
human-readable artifact. The `comms` [meld](role-melds.md) is the single source
for which records a deferring role retains and for the separate authorization
that publishing, sending, and other external action needs. This page does not
restate those lists, because three hand-maintained copies had already drifted
apart before the meld existed.

## Reviewed production locations

* `internal/evaluation/evaluation.go` - typed evaluation pack rendering,
  generic fallback, profile matrix parsing, and validation remain executable.
  Role-specific profile matrices load from data assets.
* `internal/evaluation/result.go` - result decoding, score validation, and
  canonical v2 pack digesting remain executable behavior.
* `internal/roster/roster.go` - rendering remains executable. The long-form
  native-adaptation policy lives in `definitions/NATIVE-ADAPTATION.txt`.
* `internal/person/person.go` - KDL parsing, local library merge, conflict
  detection, and copy-contract validation remain executable.
* `cmd/agent-compose/main.go` - CLI help and command wiring remain executable
  glue. No profile, library, or customer-specific prose belongs there.

Generated artifacts retain logical source IDs and content digests. They do not
publish local filesystem paths.
