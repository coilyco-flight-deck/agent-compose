# Role-skill context budget

This comparison measures the role-skill migration against released `v1.32.0`.
Both sides use the embedded `roster:core` profile. The assigned measurement uses
`testdata/contracts/native.kdl` with Engineer selected. Approximate tokens use
the deliberately simple `ceil(bytes / 4)` heuristic.

* Unassigned native Codex - `AGENTS.COMPOSE.md` fell from 31,578 bytes
  (about 7,895 tokens) to 14,434 bytes (about 3,609), a 54% reduction.
* Unassigned native Claude - `AGENTS.COMPOSE.md` plus `AGENTS.claude.md`
  fell from 32,961 bytes (about 8,241 tokens) to 14,616 bytes (about
  3,654), a 56% reduction.
* Assigned Engineer bundle - startup `content/instructions.md` fell from
  7,215 bytes (about 1,804 tokens) to 2,145 bytes (about 537), a 70%
  reduction. The role and personality bodies remain present as selected
  skills, and compiled delivery still embeds them.

`ward exec test` runs `scripts/context-budget.sh` and reports the current
measurements. The released baseline is retained here because rerunning a newer
binary cannot reconstruct an older renderer.

## See also

* [role skills](role-briefings.md) - source and loading contract.
* [features](FEATURES.md) - shipped capability inventory.
* [issue-suite run journal](run-115.md) - current one-shot execution state.
