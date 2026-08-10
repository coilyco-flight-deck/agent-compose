# Reference: A pragmatic guide to LLM evals for devs

The reference with the widest working-engineer reach, and the one that argues
for binary judgments in front of an audience that does not read papers.

<https://newsletter.pragmaticengineer.com/p/evals>

## Source

Gergely Orosz and Hamel Husain, The Pragmatic Engineer, 2 December 2025.

**Partly paywalled.** The article is free through section 3 and truncated after
it. The free portion carries the parts cited here, so the link stays usable
without a subscription. Say so when citing it, rather than sending a reader
into a wall unwarned.

## What it says

In the freely readable portion:

* The **vibe check** problem, shipping on informal testing rather than
  systematic evaluation.
* Three gaps in LLM development: comprehension, specification, and
  generalization.
* **Error analysis** as a structured method borrowed from qualitative research.
  Review traces, open-code them descriptively, axial-code them into categories,
  then prioritise by what the data shows.
* Code-based evaluators for deterministic checks, LLM-as-judge for subjective
  ones.
* **Binary pass or fail beats point scales**, because a scale invites
  ambiguity that a boundary does not.

## How it relates to this eval

The binary-judgment argument is the boundary and role-fit tiers. Those cases
have a right answer, so scoring them across four graded dimensions would
manufacture grader discretion where none is warranted.

Error analysis is what the notes on deductions accumulate toward. The board
produces failure categories rather than a single score, which is the same
output the guide's open-then-axial coding aims at.

## See also

* [Eval references](eval-references.md) - the full reference set.
* [Eval grading](eval-grading.md) - scoring tiers and the pair rule.
