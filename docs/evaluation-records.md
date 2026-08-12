# Evaluation review records

[`evaluations/latest/`](../evaluations/latest/) contains the independently
reviewed baseline. Historical mode renders archived v2 records without
rebinding their pack digests.

## V3 compact format

V3 is the compact human review format. Each case preserves its question and
answer with typographic apostrophes normalized to ASCII for Forgejo review. It
also carries a criterion-to-score mapping, total, verdict, model, and optional
finish reason. Notes appear only for deductions and receive the same
normalization.

Record provenance retains the source revision, pack digest, independent
reviewer, and any retries.

## Release gating

The `commodity` subject lane gates release. It is the only lane the board runs,
so no second lane waits to be enabled and none merely informs. An archived
record keeps whatever tier it was earned on and renders in historical mode.

## Unreviewed driver output

[`evaluations/baseline/`](../evaluations/baseline/) holds preserved driver
output that no reviewer has scored yet. It carries the pack digest, resolved
model, finish reason, retries, and the context-isolation state of the run, so
a later review binds to the exact contract that produced it. Unreviewed output
never gates a release and never enters a scorecard.

## See also

* [evaluation.md](evaluation.md) - pack generation and the Core matrix.
* [evaluation-policy.md](evaluation-policy.md) - driver and reviewer contract.
* [evaluation-driver.md](evaluation-driver.md) - executable driver and reviewer.
* [evaluation-scorecard.md](evaluation-scorecard.md) - validated aggregate view.
