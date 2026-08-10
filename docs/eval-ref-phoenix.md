# Reference: Phoenix annotation

The platform version of the grading half of this system. Referenced as a
comparison point rather than adopted.

<https://arize.com/docs/phoenix/tracing/how-to-tracing/feedback-and-annotations/annotating-in-the-ui>

## Source

Arize Phoenix documentation, "Annotating in the UI". Phoenix is an open-source
LLM observability and evaluation platform that self-hosts in one command and
builds on OpenTelemetry.

## What it says

Human annotations attach labels to any span or trace in the UI, in three
shapes:

* **Categorical**, predefined labels chosen from a list.
* **Continuous**, a score across a range.
* **Freeform**, open-ended text.

Annotated traces can be filtered to a sample and exported to a dataset, which
the docs describe as an input to experimentation, fine-tuning, or building a
human-aligned eval.

## How it relates to this eval

This is `evalkit.grade` with a vendor's name on it. Categorical annotation is
the pass-or-fail and three-way tiers, freeform is the note on a deduction, and
the export-to-dataset path is what a future judge calibration would use if
human labels ever become a gold set for an automated evaluator.

The page is cited for orientation, not adoption. Whether to adopt any review
surface remains the open decision in
<https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/issues/213>,
whose gate is explicit that case-study value is not an adoption criterion.

**Do not cite the Phoenix `evaluation/llm-evals` page for this work.** It
covers code-based and LLM-as-judge evaluators, and does not cover human
annotation, benchmarking evaluators against human labels, or golden datasets.
It describes the approach this eval replaced.

## See also

* [Eval references](eval-references.md) - the full reference set.
* [Eval orchestration](eval-orchestration.md) - the pipeline it compares to.
* [Reference: Inspect](eval-ref-inspect.md) - the platform for the other half.
