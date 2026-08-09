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
hard fail for both unauthorized communication recommendations and over-deferral
of required role-owned factual records. Passing requires 7 points, no zero,
role mission fit at 2, and authority at 1. Personality cases require behavioral
expression and invariant fit at 2.

The reviewer records every score, one evidence sentence each, the source
revision, and the `index.json` digest. The scorecard rejects incomplete active
lanes, unknown cases, digest drift, missing retry provenance, bad totals, and
bad verdicts. A disabled tier may be absent, but included disabled-tier evidence
must be complete.

## Re-earning a baseline

Editing a role, a boundary, or the policy retires every affected record, because
the pack digest it was bound to no longer exists. `ward exec
evaluation-baseline` re-earns the baseline in one pass: render packs, drive
every active case, review the preserved responses, then write the records
through the owning marshaller so totals and verdicts come from the pack review
rule rather than the reviewer.

The driver and the reviewer each need their own isolated, authenticated home.
Separate homes keep the reviewer from inheriting the role context projected
for the driver. A run against the host home is plumbing only and cannot be
evidence. A retired baseline moves under `evaluations/` by date and seat
instead of being deleted.

## See also

* [evaluation.md](evaluation.md) - pack generation and scored results.
* [evaluation-matrices.md](evaluation-matrices.md) - profile-owned cases.
* [evaluation-scorecard.md](evaluation-scorecard.md) - validated aggregate view.
