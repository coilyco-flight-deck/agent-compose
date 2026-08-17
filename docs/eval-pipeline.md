# The eval pipeline

## Eval orchestration

`evalkit` is the Python half of the evaluation system. Go owns what a pack is
and what a valid record is, `aos-eval` owns grading, shared with sirens-echo so
the pairing rule has one implementation, and what stays here is the runner plus
the roster-derived case list. Orchestration changes often and
is throwaway, contracts change rarely and are load-bearing, and the rule keeping
that seam honest is that **Python consumes what Go emits and never restates
it.** No second pack schema, coverage rule, or record writer: two parsers is the
failure this split avoids.

### Pipeline

```text
generator         ->  samples.yaml
inspect eval      ->  .eval log      (five epochs per sample, unscored)
evalkit.filter      ->  dataset.yaml   (epoch 1 attached to every sample)
aos-eval annotate   ->  annotations.yaml
aos-eval taxonomy   ->  failure modes, ranked
```

Go turns annotations into the canonical record, so totals and verdicts come
from the pack rule rather than from the annotator.

### The dataset is derived

`evalkit.matrix` reads the roster Go exports and prints the cases it implies.
Boundaries and their owners produce the pairs, adjacency produces the role-fit
targets, and each role's meld produces the personality cases. Adding a
boundary, flipping an adjacency edge, or swapping a personality moves the
sample list on its own, so the dataset cannot drift from the roster. Adjacency
reasons become the role-fit descriptors directly, which is the same text a
generator needs to construct the right confusion.

### Execution and transport

Execution is one case per session. Batching a role's cases into one request
would save background machine time and no human time, while manufacturing the
reflexive deferral the in-out pair exists to catch, breaking per-case
independence, and adding order effects. Whether doctrine survives accumulated
context is a real question, but a second arm rather than a cheaper version of
this one. `evalkit.run` records the transport on every response, and Agent Proxy
is the only one producing a measured result: `--direct` exists for incident
isolation, marks its output, and warns.

### There is no mechanical scorer

Samples carried a `discriminator`, a list of regexes for the failing behaviour,
and item analysis used the match count to pick which samples reached the
annotator. Across nine pass-or-fail cases the patterns and the grader agreed on
nothing that mattered: one false positive, two false negatives, and six
agreements on cases where nothing happened. A follow-up probe repeated it, three
detectors disagreeing with each other and with a reading of the same twenty
responses. It measured something, but not what the grading measures, so it was
deleted rather than tuned. One authored case per slot now, since the match count
was also the only rule for choosing between candidates. Epochs stay at five,
with the annotator seeing epoch 1 and the rest left in the log as evidence.

### Commands

The justfile carries `evalkit-sync`, `evalkit-run`, `evalkit-filter`, `evalkit-annotate`,
`evalkit-matrix`, and `evalkit-check`. The last runs ruff, format, mypy strict,
and pytest from `scripts/evalkit-check.sh` rather than pre-commit, since that
file is managed by agentic-os and a hand-added hook is lost on sync.

## Eval export

`aos-eval export` projects a committed run into a display payload. **One way,
and nothing returns.** Committed records stay canonical, so the surface reading
this payload is a rebuildable projection rather than a second home for
evidence. Deployment target and its reasoning:
[coilyco-bridge/deploy#572](https://forgejo.coilysiren.me/coilyco-bridge/deploy/issues/572).

### Why one way and what it reads and what it emits and the grader's own notes are withheld and it refuses rather than redacts and commands

The display target is presentation, not review. `aos-eval annotate` remains the
grading surface, and nothing is authored in the projection, so nothing has to
come back. That removes the round trip
[agent-compose#213](https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/issues/213)
step 4 names as the hard requirement for a review UI, because this is not one.

A run directory holding the two committed files, joined on sample id:

* `dataset.yaml` - authored samples plus the subject output
* `annotations.yaml` - one human decision per sample

`format: agent-compose.eval-export.v1`, carrying counts, the pair structure,
and one record per case.

**Pairs travel as their own structure.** The pair is the scoring unit, never
the half, so a renderer receives `complete` and `passed` rather than
reconstructing them from case rows and getting a half-graded pair wrong.

`critique` and `evidence` are free text written by the grader, for the grader,
in a keystroke-driven TUI. They are **not exported unless `--include-private`
is passed**, and the payload records which way it went in
`includes_private_fields`. The default is not squeamishness. Labels, prompts,
targets, outputs, and the pair structure are what make a board legible. Margin
notes are not, and the display target is a permanent public recording.

A scrubber that misses a pattern ships the secret. This stops instead, names
every reason in one pass, and exits non-zero. Recognized: AWS key ids, bearer
and API tokens, JWTs, private key blocks, SSM parameter paths, Discord
snowflakes, tailnet hosts, and email addresses.

**Withheld text is not scanned**, because text that never leaves cannot leak,
and refusing on it would block an export that is safe.

```
just evalkit-export <run-dir>
just evalkit-export <run-dir> --format yaml --out run.yaml
just evalkit-export <run-dir> --include-private
```

## See also

* [eval-ref-platforms.md](eval-ref-platforms.md) - why Phoenix's own annotation
  surface is a reference here rather than an adoption.
