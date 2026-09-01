# Eval references: the papers

The papers and practitioner writing behind this system's method. Platforms are on
[eval-ref-platforms.md](eval-ref-platforms.md).

## Reference: CheckList

Behavioral testing for models, built on the analogy to unit testing. The oldest of the six
references and the one that names the board's shape. <https://arxiv.org/abs/2005.04118>

### Source

Ribeiro, Wu, Guestrin, and Singh. *Beyond Accuracy: Behavioral Testing of NLP Models with
CheckList*. ACL 2020. Tooling at <https://github.com/marcotcr/checklist>.

### What it says

Held-out accuracy overestimates model performance, because a single aggregate number hides
which behaviors work. CheckList replaces it with a matrix: linguistic capabilities as rows,
test types as columns, and a failure rate in each cell. Three test types:

* **Minimum Functionality**, simple cases checking one behavior inside a capability.
  Modelled directly on unit tests.
* **Invariance**, a label-preserving perturbation where the prediction must not change.
* **Directional Expectation**, a perturbation where the label is expected to change in a
  specific way.

### How it relates to this eval

The dataset is a CheckList matrix. Roles are the capabilities, test types are the columns,
and each cell carries a pass rate rather than one aggregate score. That is why
`evalkit.matrix` derives cases from the roster instead of holding a list: the matrix is the
artifact, and the cases are what fills it. The boundary in-out pair is a Directional
Expectation test. The perturbation is moving a situation across the boundary, and the
expected verdict flips with it. The framing is also the most useful one for explaining this
work to software engineers, since "unit tests for behavior" needs no further translation.

## Reference: LLM Evals, Everything You Need to Know

The practitioner FAQ, and the reference that most directly defends the choices this
repository made about tooling and about who grades.
<https://hamel.dev/blog/posts/evals-faq/>

### Source

Hamel Husain and Shreya Shankar, published 28 May 2025 and still being revised, most
recently 18 July 2026.

### What it says

* **Error analysis is the essential activity.** Review real failures before building
  evaluators for hypothetical ones.
* **The benevolent dictator.** Rather than distributing annotation across many reviewers,
  designate one domain expert who understands the users and sets the standard. This removes
  annotation conflict and keeps the bar consistent.
* **Custom annotation interfaces beat off-the-shelf platforms**, by roughly an order of
  magnitude in iteration speed, because a bespoke tool can show the right context, bind
  keyboard shortcuts, and visualise the domain.
* **Binary pass or fail beats numeric scales**, which introduce subjective ambiguity.
* Small curated datasets with deterministic checks in CI, and asynchronous sampling of live
  traffic in production.

### How it relates to this eval

Three of its recommendations are decisions already taken here:

* One expert annotator rather than a panel. Kai is the single annotator, and that is a
  design choice rather than a staffing limit.
* A custom grading interface. `housecast grade annotate` is one case per screen and one keystroke
  per decision, which is exactly the bespoke-tool argument.
* Binary judgments for anything with a right answer.

This is the reference to cite when asked why no evaluation platform was adopted. The answer
is not that platforms are bad, it is that a custom annotation surface is the recommended
practice for this stage.

## Reference: A pragmatic guide to LLM evals for devs

The reference with the widest working-engineer reach, and the one that argues for binary
judgments in front of an audience that does not read papers.
<https://newsletter.pragmaticengineer.com/p/evals>

### Source

Gergely Orosz and Hamel Husain, The Pragmatic Engineer, 2 December 2025.

**Partly paywalled.** The article is free through section 3 and truncated after it. The free
portion carries the parts cited here, so the link stays usable without a subscription. Say
so when citing it, rather than sending a reader into a wall unwarned.

### What it says

In the freely readable portion:

* The **vibe check** problem, shipping on informal testing rather than systematic
  evaluation.
* Three gaps in LLM development: comprehension, specification, and generalization.
* **Error analysis** as a structured method borrowed from qualitative research. Review
  traces, open-code them descriptively, axial-code them into categories, then prioritise by
  what the data shows.
* Code-based evaluators for deterministic checks, LLM-as-judge for subjective ones.
* **Binary pass or fail beats point scales**, because a scale invites ambiguity that a
  boundary does not.

### How it relates to this eval

The binary-judgment argument is the boundary and role-fit tiers. Those cases have a right
answer, so scoring them across four graded dimensions would manufacture annotator discretion
where none is warranted. Error analysis is what the critiques on deductions accumulate
toward. The board produces failure categories rather than a single score, which is the same
output the guide's open-then-axial coding aims at.

## See also

* [Evaluation](evaluation.md) - the eval this places, and what it does with these.
