# Eval export

`evalkit.export` projects a committed run into a display payload. **One way,
and nothing returns.** Committed records stay canonical, so the surface reading
this payload is a rebuildable projection rather than a second home for
evidence.

Deployment target and its reasoning:
[coilyco-bridge/deploy#572](https://forgejo.coilysiren.me/coilyco-bridge/deploy/issues/572).

## Why one way

The display target is presentation, not review. `evalkit.annotate` remains the
grading surface, and nothing is authored in the projection, so nothing has to
come back. That removes the round trip
[agent-compose#213](https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/issues/213)
step 4 names as the hard requirement for a review UI, because this is not one.

## What it reads

A run directory holding the two committed files, joined on sample id:

* `dataset.yaml` - authored samples plus the subject output
* `annotations.yaml` - one human decision per sample

## What it emits

`format: agent-compose.eval-export.v1`, carrying counts, the pair structure,
and one record per case.

**Pairs travel as their own structure.** The pair is the scoring unit, never
the half, so a renderer receives `complete` and `passed` rather than
reconstructing them from case rows and getting a half-graded pair wrong.

## The grader's own notes are withheld

`critique` and `evidence` are free text written by the grader, for the grader,
in a keystroke-driven TUI. They are **not exported unless `--include-private`
is passed**, and the payload records which way it went in
`includes_private_fields`.

The default is not squeamishness. Labels, prompts, targets, outputs, and the
pair structure are what make a board legible. Margin notes are not, and the
display target is a permanent public recording.

## It refuses rather than redacts

A scrubber that misses a pattern ships the secret. This stops instead, names
every reason in one pass, and exits non-zero.

Recognized: AWS key ids, bearer and API tokens, JWTs, private key blocks, SSM
parameter paths, Discord snowflakes, tailnet hosts, and email addresses.

**Withheld text is not scanned**, because text that never leaves cannot leak,
and refusing on it would block an export that is safe.

## Commands

```
ward exec evalkit-export -- <run-dir>
ward exec evalkit-export -- <run-dir> --format yaml --out run.yaml
ward exec evalkit-export -- <run-dir> --include-private
```

## Related

* [Eval orchestration](eval-orchestration.md) - the pipeline that produces a run.
* [Reference: Phoenix annotation](eval-ref-phoenix.md) - why the platform's own
  annotation surface is a reference here rather than an adoption.
