---
name: eval-role-comms
description: Curate cross-role behavioral evaluations for human communication ownership, factual handoffs, drafting and advice boundaries, and separate send or publication authority. Use when the active AI Engineer creates, updates, or diagnoses role communication evaluations, especially after roster, policy, model, or observed-regression changes.
---

# Role Communication Evaluation

Treat communication ownership as one executable cross-role concern. Evaluate
the selected roster's contract without deciding policy, copying its role
matrix, or granting delivery authority.

## Establish the contract

1. Locate the selected person package, role briefings, communication policy,
   and generated evaluation pack. Prefer the selected package over remembered
   defaults because another package may assign different owners.
2. Record source revisions, pack digest, active model tiers, harness, and seat.
   Reject a result that cannot identify the exact contract under test.
3. Enumerate roles from the selected package instead of maintaining another
   role list. Extract each role's allowed action, prohibited action, required
   handoff, and delivery gate from owning sources.
4. Stop on contradictory or missing policy. Report the source gap instead of
   inventing an expected answer or weakening the case.

## Curate coverage

Cover every role with the smallest cases that distinguish these boundaries:

* Owner behavior - accept a sufficient factual handoff and produce only the
  communication work the canonical contract assigns.
* Non-owner behavior - stop before drafting, rewriting, evaluating wording, or
  recommending tone, framing, timing, channel, or reply strategy.
* Factual handoff - preserve verified facts, constraints, risks, audience, and
  authorized action without smuggling suggested language into the handoff.
* Indirect pressure - test urgency, convenience, quoted drafts, requests to
  "just polish," and mission or personality framing without teaching the answer.
* Delivery separation - distinguish preparing or executing an approved artifact
  from authorization to send, post, publish, interview, or speak.

Keep prompts separate from expected outcomes. Add a historical regression as a
case, never as a hint embedded in the prompt. Require one positive, prohibited,
ambiguous, and adjacent-owner case whenever the contract supports them.

## Run and classify

Use the evaluation pack's driver and reviewer policy. Hold role bundle, prompt,
model, runtime, tools, retry policy, and run count constant across comparisons.
Start a fresh isolated session per case and preserve raw responses, finish
reasons, retries, and failures before review.

Classify failures as policy-source, case-specification, context-construction,
model-behavior, runner, or reviewer defects. Do not repair a policy-source gap
with scorer leniency. Keep the prompt or rubric author from serving as the sole
reviewer, and retain QA's independent acceptance.

## Deliver evidence

Report coverage by role, tier, harness, and scenario boundary. Name missing
coverage, behavior changes from the prior baseline, hard failures, uncertainty,
and the smallest owning source to revise. Regenerate derived packs and
scorecards after an accepted source change while retaining failed evidence.
