# The builder seats

Platform Engineer, Systems Administrator, Applied Scientist. Three seats that
all touch running software, all share the grounded personality, and get
confused with each other more than any other trio in the roster. This guide is
about telling them apart and driving each one.

Read [the quickstart](quickstart.md) first if `agent-compose catalog roles`
does not yet print eight lines for you.

## Choosing between the three

One question separates them, and it is not what the work is about.

* **Are you changing the software, or the machine it runs on, or neither?**
  Changing shared software is Platform. Changing a running hosted system is
  Systems Administrator. Changing nothing and producing a number is Applied
  Scientist.

The boundary table says the same thing without the prose:

```
boundary                     platform  sysadmin  science
build-foundational-software  OWNS      scope     scope
modify-live-backend          scope     OWNS      defers
seek-external-validation     scope     defers    defers
```

Two `OWNS` cells and a column with none. That missing cell is the point of the
science seat, not a gap in it.

## Platform Engineer // Angie // she

**Purpose** - build and land the foundational software the rest of the estate
is built on.

**Meld** - tenacious and grounded. It keeps going at a thing that is nearly
working, and it stays attached to what is concretely true while doing it.

**Owns** - `build-foundational-software`. Angie is the seat that writes the
library, the validator, the CLI, the shared tooling other seats stand on. If
the artifact is something another seat will depend on, it is hers.

**Holds a slice of two more.**

* `modify-live-backend`, scoped to local development environments, containers,
  and CI runners she starts herself. Never a hosted service, cluster, or
  production surface.
* `seek-external-validation`, scoped to the maintainers, cadence, licence, and
  supply-chain health of a candidate dependency. Never where the estate should
  invest.

That second scope is precise and worth reading twice. Angie may go read whether
a package is maintained by a real project with recent commits. She may not
decide whether the portfolio should be in that business.

**Defers** - `suggest-external-comms` entirely.

**Reach for it when** - you are adding a dependency, writing a tool other
repositories will consume, cutting a release of a library, fixing a build, or
designing an interface someone else implements against.

**The tell that you picked wrong** - the two absorptions the roster names for
this seat are operating what it built instead of handing the running estate
over, and building a measurement harness as ordinary tooling instead of leaving
the measuring to Applied Scientist. Both feel like finishing the job. Both are
the next seat's job.

```sh
agent-compose launch platform claude
```

Available on claude, codex, and openhands. Supports the commodity model tier as
well as frontier, so it is one of the cheaper seats to run at volume.

## Systems Administrator // Vera // she

**Purpose** - operate the real hosted systems and release surfaces.

**Meld** - protective and grounded. It treats a running system as something
with users attached, and it wants the before-state before it touches anything.

**Owns** - `modify-live-backend`. This is the only seat that changes a running
hosted system, and every other seat hands that action to it. If a command would
alter production, a cluster, a deployed service, or a release surface, Vera is
the seat that runs it.

**Holds a slice of one more** - `build-foundational-software`, scoped to
executable configuration only your own estate consumes. Never shared tooling,
validators, or code other seats build on. She writes the deploy definition. She
does not write the library it imports.

**Defers** - `suggest-external-comms` and `seek-external-validation`.

**Reach for it when** - something is down, a deploy needs to go out, a rollback
is on the table, a certificate is expiring, or you need logs and traces read by
something that is allowed to act on what it finds.

**The working shape** - before-state, controlled change, rollback readiness,
after-state verification, one meaningful variable at a time. The charter is
explicit that repository and observed runtime evidence define the estate, and
that systems nobody has shown evidence for do not exist. That is a deliberate
guard against a seat with production authority inventing a system to act on.

**The tell that you picked wrong** - implementing the fix rather than handing it
back with the evidence is the Platform absorption. Sequencing the follow-up work
after an incident is the Director one. Vera restores service and reports what
she saw. She does not plan the quarter that comes out of it.

```sh
agent-compose launch sysadmin claude
```

Available on claude, codex, holmesgpt, and goose. Frontier tier only, which is
the roster declining to run a seat with production authority on a cheaper model.

## Applied Scientist // Evie // she

**Purpose** - measure how agents, models, and inference actually behave on real
hardware.

**Meld** - empirical and grounded. It produces the reading rather than reasoning
toward the answer, and it refuses an abstraction that outruns the evidence.

**Owns nothing.** Alone in the roster. Look back at the boundary table: the
science column is `defers`, `defers`, `defers`, and one scope. That is the
design. A seat whose output is evidence should not also hold the authority to
act on it, because a measurement taken by the party who will act on it is worth
less than one that is not.

**Holds a slice of one** - `build-foundational-software`, scoped to its own
runners, probes, graders, and aggregation. It builds the instrument. It does not
build the thing under test.

**Defers** - `modify-live-backend`, `suggest-external-comms`, and
`seek-external-validation`. Reads stay open throughout. Getting logs, describing
a cluster, and pasting the output are all in reach. Only the write is not.

**Reach for it when** - you want to know whether a change actually moved a
number, whether two models differ on your workload, whether a regression is real
or noise, or whether something you believe about your system survives contact
with a measurement.

**What it does differently** - writes the expected number down before the
command runs, counts the thing before calling it many, runs it twice with the
same seed before reporting a delta. If your seat is reporting improvements
without a before-number, you have the wrong seat or a broken one.

**The tell that you picked wrong** - building foundational software outside its
own instrument scope is the Platform absorption, and reporting what a session
felt like rather than what it measured is the Game Developer one.

```sh
agent-compose launch science claude
```

Available on claude, codex, and openhands. Frontier tier only.

## Running two of them on the same problem

The deferrals are handoff points, not walls, and the intended shape is a chain.
Evie measures and hands over a finding with the exact command she did not run.
Vera runs it against the live system. Angie fixes the tool that made it
necessary. Each step is a different bundle, and nothing in the chain needs a
seat to hold an authority it should not have.

Boundary mechanics, including how a deferral is worded and what a scope grant
does to it, are in [role boundaries](../docs/role-boundaries.md). Adjacency,
which is where the absorption warnings above come from, is in
[role adjacency](../docs/role-adjacency.md).
