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

## Meld extraction

Extracting the shared communication and live-operations boundaries into
[role melds](role-melds.md) freed 602 words of role body prose across the eight
Core Roster roles, from 2,759 to 2,157, a 22% reduction. Per role: engineer 295
to 180, director 318 to 195, qa 298 to 196, ops 394 to 264, design 395 to 364,
exec 288 to 226, ai 369 to 330, and creator unchanged at 402 because it owns
the communication boundary rather than deferring to it.

The melded bodies are additive rather than deducted. Each is bounded by its own
400-word ceiling and never enters `Role.Briefing`, so the freed budget is
available to role-specific charter prose.

## Comms rescope

Dropping the no-invention paragraph, which is a roster-wide honesty rule rather
than a communication boundary, took `comms` from 259 to 227 words. The meld now
binds design, exec, and ops instead of seven roles. Role bodies are unchanged,
since the four unbound roles carried no comms prose to restore.

## Evidence meld

The `evidence` meld binds engineer, exec, and ops, and spends 357 words of its
own 400-word ceiling. It charges no role body. Role prose is unchanged except exec,
which grew 223 to 244 words when the shared acquisition obligation replaced its
`Prefer primary evidence` clause and absorbed the strategist residue about
measuring rather than assuming portfolio numbers. Every role body stays under
its 400-word ceiling, and Core Roster role prose totals 2,159 words. These
counts come from the loader's own body counter, so they differ by a word or two
from the extraction figures above, which were measured with a separate tool.

## See also

* [role melds](role-melds.md) - shared doctrine and its separate budget.
* [role skills](role-briefings.md) - source and loading contract.
* [features](FEATURES.md) - shipped capability inventory.
* [issue-suite run journal](run-115.md) - current one-shot execution state.
