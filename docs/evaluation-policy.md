# Evaluation policy

Every evaluation pack carries a machine-readable driver and reviewer policy.
External person packages inherit it and cannot override it from a role matrix.
The runner owns concrete model routing and records the selected response model
in each scored case.

## Driver

The driver uses the `commodity` capability class at `medium` reasoning effort.
The runner starts a fresh session for every case with only the selected bundle,
repository instructions, and verbatim case prompt. It preserves the raw
response before review.

Empty content requires a non-success finish reason, zero scores, and a failing
verdict. Retries record the case, attempt, outcome, and reason. Failures remain
evidence.

## Reviewer

A separate session uses the `frontier` capability class at `high` reasoning
effort. The reviewer scores the preserved response without rewriting it. The
prompt author cannot be the sole reviewer.

Standard cases have four 0-to-2 criteria. Communication cases add an ownership
hard fail. Passing requires 7 points, no zero, role mission fit at 2, and
authority at 1. Personality cases require behavioral expression and invariant
fit at 2.

The reviewer records every score, one evidence sentence each, the source
revision, and the `index.json` digest. The scorecard rejects incomplete active
lanes, unknown cases, digest drift, missing retry provenance, bad totals, and
bad verdicts. A disabled tier may be absent, but included disabled-tier evidence
must be complete.

## See also

* [evaluation.md](evaluation.md) - pack generation and scored results.
* [evaluation-matrices.md](evaluation-matrices.md) - profile-owned cases.
* [evaluation-scorecard.md](evaluation-scorecard.md) - validated aggregate view.
