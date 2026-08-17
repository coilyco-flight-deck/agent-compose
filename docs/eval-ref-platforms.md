# Eval references: the platforms

## Reference: Inspect

The platform version of the execution half of this system, and the closest
structural match to how the pipeline is organised.
<https://inspect.aisi.org.uk/>

### Source

Inspect, from the UK AI Security Institute. Open-source Python, installed with
`pip install inspect-ai`, developed in the open at
<https://github.com/UKGovernmentBEIS/inspect_ai> and actively released.

### What it says

Evaluations decompose into four primitives:

* **Dataset**, test cases each carrying an input and a target.
* **Task**, the unit that binds a dataset to how it runs.
* **Solver**, what produces an answer, from a single generate call up to a
  full tool-using agent.
* **Scorer**, what turns an answer into a result.

It ships sandboxed execution, a web-based viewer, a VS Code extension, and a
library of prebuilt evaluations.

### How it relates to this eval

The dataset, solver, scorer split is the same seam this repository draws
between samples, subject, and annotator. Citing Inspect is the shortest way to
show that the structure here is conventional rather than invented, and that it
is the shape frontier evaluation work uses. Where it diverges: Inspect's scorer
is normally programmatic or model-graded, and this eval puts a human there.
That difference is the point rather than a gap, so the comparison is worth
drawing explicitly rather than eliding.

**Adopted for the run leg.** `evalkit/task.py` is an Inspect task, `--epochs`
replaced the hand-rolled n=5 fan-out, and the `.eval` log replaced a jsonl
sink. `inspect view` comes along as a browser surface.

Annotation stays outside it, which Inspect supports directly: `--no-score`
produces a log of unscored samples, and its score-editing API exists for
"applying manual review adjustments". Adopting a review UI remains separate and
open in
<https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/issues/213>.

## Reference: Phoenix annotation

The platform version of the grading half of this system. Referenced as a
comparison point rather than adopted.
<https://arize.com/docs/phoenix/tracing/how-to-tracing/feedback-and-annotations/annotating-in-the-ui>

### Source and what it says and how it relates to this eval and source and what it says and how it relates to this eval

Arize Phoenix documentation, "Annotating in the UI". Phoenix is an open-source
LLM observability and evaluation platform that self-hosts in one command and
builds on OpenTelemetry.

Human annotations attach labels to any span or trace in the UI, in three
shapes:

* **Categorical**, predefined labels chosen from a list.
* **Continuous**, a score across a range.
* **Freeform**, open-ended text.

Annotated traces can be filtered to a sample and exported to a dataset, which
the docs describe as an input to experimentation, fine-tuning, or building a
human-aligned eval.

This is `aos-eval annotate` with a vendor's name on it. Categorical annotation
is the pass-or-fail and three-way tiers, freeform is the critique on a
deduction, and the export-to-dataset path is what a future judge calibration
would use if human labels ever become a gold set for an automated evaluator.
The page is cited for orientation, not adoption. Whether to adopt any review
surface remains the open decision in
<https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/issues/213>,
whose gate is explicit that case-study value is not an adoption criterion.

**Do not cite the Phoenix `evaluation/llm-evals` page for this work.** It
covers code-based and LLM-as-judge evaluators, and does not cover human
annotation, benchmarking evaluators against human labels, or golden datasets.
It describes the approach this eval replaced.

## Reference: RULERS

Prescriptive framework for making rubric-based scoring reproducible. The newest
of the six references, and the one whose recommendations this repository had
already implemented. <https://arxiv.org/abs/2601.08654>

Hong et al. *From Rubrics to Reliable Scores: Evidence-Grounded Text Evaluation
with LLM Judges*, January 2026. The framework is named RULERS, for Rubric
Unification, Locking, and Evidence-anchored Robust Scoring.

A compiler-executor design that turns natural-language rubrics into executable
specifications, targeting three recurrent failure modes:

* **Rubric instability from prompt sensitivity.** Fixed by compiling criteria
  into versioned, immutable bundles.
* **Unverifiable reasoning with no auditable evidence.** Fixed by requiring
  verbatim quotes from the input for every high score, with deterministic
  verification to catch hallucinated justifications.
* **Scale misalignment with human grading boundaries.** Fixed by post-hoc
  calibration against human annotations.

It reports stronger agreement with human judgment than conventional prompting,
holding up under adversarial rubric perturbation.

Each of the three failure modes maps onto a decision already made here, which
makes this the reference worth leading with:

* Versioned immutable criteria bundles are the pack digest contract. Editing a
  role, a meld, or the policy retires every record bound to the old digest.
* Evidence anchoring is the note-on-deduction rule.
* Calibrating the scale to human grading boundaries is why personality moved
  off a 1-to-5 scale to fit, undecided, or does not fit.

The convergence is independent rather than derived. That is worth stating
plainly wherever this is cited, since a reader will otherwise assume the design
followed the paper.
