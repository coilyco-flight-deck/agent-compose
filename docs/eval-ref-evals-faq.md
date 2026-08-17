# Reference: LLM Evals, Everything You Need to Know

The practitioner FAQ, and the reference that most directly defends the choices
this repository made about tooling and about who grades.

<https://hamel.dev/blog/posts/evals-faq/>

## Source

Hamel Husain and Shreya Shankar, published 28 May 2025 and still being
revised, most recently 18 July 2026.

## What it says

* **Error analysis is the essential activity.** Review real failures before
  building evaluators for hypothetical ones.
* **The benevolent dictator.** Rather than distributing annotation across many
  reviewers, designate one domain expert who understands the users and sets the
  standard. This removes annotation conflict and keeps the bar consistent.
* **Custom annotation interfaces beat off-the-shelf platforms**, by roughly an
  order of magnitude in iteration speed, because a bespoke tool can show the
  right context, bind keyboard shortcuts, and visualise the domain.
* **Binary pass or fail beats numeric scales**, which introduce subjective
  ambiguity.
* Small curated datasets with deterministic checks in CI, and asynchronous
  sampling of live traffic in production.

## How it relates to this eval

Three of its recommendations are decisions already taken here:

* One expert annotator rather than a panel. Kai is the single annotator, and that is
  a design choice rather than a staffing limit.
* A custom grading interface. `aos-eval annotate` is one case per screen and one
  keystroke per decision, which is exactly the bespoke-tool argument.
* Binary judgments for anything with a right answer.

This is the reference to cite when asked why no evaluation platform was
adopted. The answer is not that platforms are bad, it is that a custom
annotation surface is the recommended practice for this stage.

## See also

* [Eval references](eval-references.md) - the full reference set.
* [Eval annotation](eval-annotation.md) - scoring tiers, ordering, and item analysis.
