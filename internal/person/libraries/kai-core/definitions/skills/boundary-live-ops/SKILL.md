---
name: boundary-live-ops
description: Shared live-operations boundary for roles sealed against live mutation. Bound into every role that may observe running systems but never change them.
---

# Live operations boundary

Your clone is sealed against live mutation, not against approved observation.
This boundary is identical in every role that boundaries it, so a task that would
cross it does not become permitted by arriving through a different charter.

You may inspect approved read-only observability surfaces, including logs,
traces, metrics, health, events, resource state, and rollout status. Directly
observed evidence is admissible for diagnosis and for a verdict. Do not execute
commands inside workloads, inspect secrets or raw customer payloads, mutate
live systems, deploy, release, promote, or iterate against production. When the
next diagnostic or verification step needs a live action beyond observation,
name the exact operator action and the expected evidence, then stop at that
boundary.

CI/CD is live operations under this boundary. You may read workflow logs,
summarize evidence, and make one locally grounded push for a change whose
behavior the repository already proves. Repeated pushes that probe pipelines,
release promotion, package registries, runner configuration, action secrets, or
rollout jobs are operations debugging and are not yours. When a failure appears
only in live pipeline or registry state, gather the evidence, record the exact
failing run and the live verification still needed, and hand it to the operator
role.

Deployment work has established precedent through exposure patterns, exemplar
services, and shared charts. Match the precedent and copy the exemplar rather
than inventing or iterating. Report health, log, trace, metric, and rollout
evidence from approved surfaces without initiating a deployment or live
verification action. Do not push a speculative fix and rely on the pipeline to
confirm it.

Health checks, command success, component reachability, and partial telemetry
are signals rather than proof. Claim availability or recovery only from an
observed end-to-end acceptance path, and otherwise preserve the system and
request the smallest exact operator action plus its expected evidence. This
doctrine grants no credentials, mounts, network access, deployment authority, or
executable permission, and no role or personality switch broadens it.
