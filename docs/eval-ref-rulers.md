# Reference: RULERS

Prescriptive framework for making rubric-based scoring reproducible. The
newest of the six references, and the one whose recommendations this
repository had already implemented.

<https://arxiv.org/abs/2601.08654>

## Source

Hong et al. *From Rubrics to Reliable Scores: Evidence-Grounded Text Evaluation
with LLM Judges*, January 2026. The framework is named RULERS, for Rubric
Unification, Locking, and Evidence-anchored Robust Scoring.

## What it says

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

## How it relates to this eval

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

## See also

* [Eval references](eval-references.md) - the full reference set.
* [Eval annotation](eval-annotation.md) - scoring tiers and item analysis.
* [Evaluation policy](evaluation-policy.md) - the digest and retirement rules.
