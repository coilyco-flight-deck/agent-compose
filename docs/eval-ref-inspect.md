# Reference: Inspect

The platform version of the execution half of this system, and the closest
structural match to how the pipeline is organised.

<https://inspect.aisi.org.uk/>

## Source

Inspect, from the UK AI Security Institute. Open-source Python, installed with
`pip install inspect-ai`, developed in the open at
<https://github.com/UKGovernmentBEIS/inspect_ai> and actively released.

## What it says

Evaluations decompose into four primitives:

* **Dataset**, test cases each carrying an input and a target.
* **Task**, the unit that binds a dataset to how it runs.
* **Solver**, what produces an answer, from a single generate call up to a
  full tool-using agent.
* **Scorer**, what turns an answer into a result.

It ships sandboxed execution, a web-based viewer, a VS Code extension, and a
library of prebuilt evaluations.

## How it relates to this eval

The dataset, solver, scorer split is the same seam this repository draws
between samples, subject, and annotator. Citing Inspect is the shortest way to
show that the structure here is conventional rather than invented, and that it
is the shape frontier evaluation work uses.

Where it diverges: Inspect's scorer is normally programmatic or model-graded,
and this eval puts a human there. That difference is the point rather than a
gap, so the comparison is worth drawing explicitly rather than eliding.

Referenced for orientation, not adopted. Adoption remains the open decision in
<https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/issues/213>.

## See also

* [Eval references](eval-references.md) - the full reference set.
* [Eval orchestration](eval-orchestration.md) - the pipeline it compares to.
* [Reference: Phoenix annotation](eval-ref-phoenix.md) - the platform for the
  grading half.
