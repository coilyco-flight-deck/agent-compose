# Reference: CheckList

Behavioral testing for models, built on the analogy to unit testing. The oldest
of the six references and the one that names the board's shape.

<https://arxiv.org/abs/2005.04118>

## Source

Ribeiro, Wu, Guestrin, and Singh. *Beyond Accuracy: Behavioral Testing of NLP
Models with CheckList*. ACL 2020. Tooling at
<https://github.com/marcotcr/checklist>.

## What it says

Held-out accuracy overestimates model performance, because a single aggregate
number hides which behaviors work. CheckList replaces it with a matrix:
linguistic capabilities as rows, test types as columns, and a failure rate in
each cell.

Three test types:

* **Minimum Functionality**, simple cases checking one behavior inside a
  capability. Modelled directly on unit tests.
* **Invariance**, a label-preserving perturbation where the prediction must not
  change.
* **Directional Expectation**, a perturbation where the label is expected to
  change in a specific way.

## How it relates to this eval

The dataset is a CheckList matrix. Roles are the capabilities, test types are
the columns, and each cell carries a pass rate rather than one aggregate score. That
is why `evalkit.matrix` derives cases from the roster instead of holding a list:
the matrix is the artifact, and the cases are what fills it.

The boundary in-out pair is a Directional Expectation test. The perturbation is
moving a situation across the boundary, and the expected verdict flips with it.

The framing is also the most useful one for explaining this work to software
engineers, since "unit tests for behavior" needs no further translation.

## See also

* [Eval references](eval-references.md) - the full reference set.
* [Eval annotation](eval-annotation.md) - scoring tiers and the pair rule.
* Contrast sets, the narrower precedent for the pair specifically:
  <https://aclanthology.org/2020.findings-emnlp.117/>
