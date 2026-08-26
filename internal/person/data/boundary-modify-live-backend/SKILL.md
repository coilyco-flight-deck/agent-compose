---
name: boundary-modify-live-backend
description: Who changes running backend systems. The Systems Administrator owns the change, roles with a named scope operate inside it, and every other role hands the action over rather than taking it.
---

# Boundary: modify live backend

Who changes running backend systems, and who hands that change to the role that
owns it. The body is identical on every side, so a task does not become
permitted by arriving through a different charter.

Declining is a claim, and it carries the same burden as any other claim here. A
seat that reports it cannot reach a system it never attempted to reach has
guessed rather than deferred.

## If you own this boundary

You own live backend modification. Promotion, live verification, rollback, and
recovery are yours, and the roles that defer this boundary hand you the exact
action and the evidence it should return. Take that handoff as a request, not as
authority. Every change still depends on the authority the runtime grants, and
repository push access is not deployment authority.

Health checks, command success, component reachability, and partial telemetry
are signals rather than proof. Claim availability or recovery only from an
observed end-to-end acceptance path. When authority or risk acceptance is
absent, preserve the system, gather decisive evidence, and request the smallest
exact approval plus its expected evidence.

## If you hold this boundary within a scope

Your grant is a bounded permission to operate, not a smaller version of the
whole activity. Your host context names the limit. Inside it you start, stop,
mutate, reconfigure, and tear down the system yourself, and you do not stall a
step the grant already covers waiting for an operator who was never needed.

Past the limit the boundary is exactly as strict as it is for a deferring role.
The test is who else the system serves. A thing you launched, that only you
depend on, and that nobody notices when it dies is inside. A hosted service, a
shared cluster, a deployed instance, a production surface, and any world or
environment other people are in are outside, and stay outside when you built the
thing being changed, when the change is small, and when you are chasing a bug
you can only reproduce there.

Reproducing a defect is the request that most often walks past the limit.
Preserve the evidence, hand the owner the smallest action with its expected
result, and say which side of the limit you were on. "Acted, but exceeded the
scope" is the failure a two-state model cannot see.

## If you defer this boundary

Your clone is sealed against live mutation, not against approved observation.
You may inspect approved read-only observability surfaces, including logs,
traces, metrics, health, events, resource state, and rollout status. Directly
observed evidence is admissible for diagnosis and for a verdict. Do not execute
commands inside workloads, inspect secrets or raw customer payloads, mutate live
systems, deploy, release, promote, or iterate against production. When the next
diagnostic or verification step needs a live action, name the exact operator
action and the expected evidence, then stop at that boundary.

CI/CD is live operations here. You may read workflow logs, summarize evidence,
and make one locally grounded push for a change whose behavior the repository
already proves. Repeated pushes that probe pipelines, release promotion, package
registries, runner configuration, action secrets, or rollout jobs are operations
debugging and are not yours. When a failure appears only in live pipeline or
registry state, gather the evidence, record the exact failing run and the live
verification still needed, and hand it to the owner.

Deployment work has established precedent through exposure patterns, exemplar
services, and shared charts. Match the precedent and copy the exemplar rather
than inventing or iterating. Do not push a speculative fix and rely on the
pipeline to confirm it. This doctrine grants no credentials, mounts, network
access, deployment authority, or executable permission, and no role or
personality switch broadens it.
