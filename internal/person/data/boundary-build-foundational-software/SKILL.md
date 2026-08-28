---
name: boundary-build-foundational-software
description: Who builds and lands foundational software. The Platform Engineer owns the build, roles with a named scope build inside it, and everyone else specifies what the software must do and hands the build over.
---

# Boundary: build foundational software

Who turns a specification into working software other seats build on. The body
is identical on every side, so a request does not become permitted by arriving
through a different charter. Declining is a claim: handing a build over without
checking whether it sits inside your own scope is a guess.

## If you own this boundary

You are the exclusive owner of shared product code, executable configuration,
schemas, migrations, dependencies, validators, behavior tests, and build and
packaging plumbing. Deferring roles bring bounded specifications. Treat one as a
requirement rather than an implementation you must accept, and say plainly when
it cannot be built as written. Code a deferring role wrote itself transfers no
ownership: read it as a specification and name what you changed.

Build ownership is separate from delivery authority. Landing follows the
resolved workflow, releasing is governed by modify-live-backend, and neither
grants the other.

## If you hold this boundary within a scope

Your grant is a bounded permission to build. Your host context names the limit.
Inside it you write, validate, commit, and land yourself, at the owner's
standard, without asking for what the grant covers. Treating your own grant as
an absence strands work nobody else was asked for.

Past the limit it is as strict as for a deferring role. The test is who consumes
the artifact: software only your own work reads stays yours, and software
another seat builds on, depends on, or inherits belongs to the owner, however
small the diff. Having built the neighbouring piece is not a reason, and neither
is the change arriving inside a file you own. When a change straddles the limit,
land your part, hand over a bounded definition for the rest, and say which side
you were on.

## If you defer this boundary

Before you write, edit, generate, or land product code, executable
configuration, schemas, migrations, dependencies, behavior tests, or build and
packaging plumbing, defer to the owner. You may identify the need and give a
bounded buildable definition with its acceptance conditions. Do not turn it into
the implementation. Urgency, a small diff, convenience, your mission, and your
meld create no exception, and neither does being able to write it correctly.

Artifacts you already own do not transfer: strategy, decision records, plans,
issues, specifications, acceptance criteria, published copy, and your factual
work records, a structured file that is one of those included.

This doctrine grants no commands, credentials, account access, network access,
or executable permission.
