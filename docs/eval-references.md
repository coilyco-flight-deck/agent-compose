# Eval references

Six external reference points for the evaluation work here, two of each kind,
chosen so a reader can place this system from whichever direction they arrive.

A paper reaches someone who wants the method named. A platform reaches someone
who wants a tool to open. A blog reaches the working engineer who reads neither.

## Papers

* [CheckList](eval-ref-papers.md) - behavioral testing as a capability by
  test-type matrix, modelled on unit tests. ACL 2020.
  <https://arxiv.org/abs/2005.04118>
* [RULERS](eval-ref-platforms.md) - versioned immutable criteria, evidence-anchored
  scoring, calibration to human grading boundaries. January 2026.
  <https://arxiv.org/abs/2601.08654>

One old and one new on purpose. CheckList names the board's shape. RULERS
prescribes what to do rather than warning what to avoid, and its three
recommendations were independently already in place here.

## Platforms

* [Phoenix annotation](eval-ref-platforms.md) - the human annotation surface, so
  the platform equivalent of `evalkit.annotate`.
  <https://arize.com/docs/phoenix/tracing/how-to-tracing/feedback-and-annotations/annotating-in-the-ui>
* [Inspect](eval-ref-platforms.md) - dataset, task, solver, scorer, so the
  platform equivalent of `evalkit.run` and `evalkit.filter`.
  <https://inspect.aisi.org.uk/>

The two cover different halves rather than competing. Inspect is adopted for
the run leg. Phoenix stays a reference, since annotation is local. A review UI
remains the open decision in
<https://forgejo.coilysiren.me/coilyco-flight-deck/agent-compose/issues/213>.

## Blogs

* [A pragmatic guide to LLM evals for devs](eval-ref-papers.md) -
  Orosz and Husain, December 2025. Partly paywalled, free through section 3.
  <https://newsletter.pragmaticengineer.com/p/evals>
* [LLM Evals, Everything You Need to Know](eval-ref-papers.md) - Husain and
  Shankar, May 2025, revised through July 2026.
  <https://hamel.dev/blog/posts/evals-faq/>

## What the set agrees on

The two blogs converge, independently of the papers, on three things this
system does:

* Binary pass or fail beats a point scale wherever a case has a right answer.
* A single expert annotator beats a distributed panel.
* A custom annotation interface beats an off-the-shelf platform at this stage.

That convergence is the reason no platform was adopted. It is a recommended
practice rather than a shortcut, and the references say so directly.

## See also

* [Eval annotation](evaluation.md) - scoring tiers, ordering, and item analysis.
* [Eval orchestration](eval-pipeline.md) - the pipeline and its seam.
