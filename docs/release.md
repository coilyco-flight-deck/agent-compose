# Release

Agent Compose releases are Forgejo-canonical. Every push to `main` enters a
no-cancel queue and validates the exact commit. The owning
`scripts/release-impact.sh` classifier then decides whether publication runs.

Automatic publication occurs when the unreleased diff from the latest reachable
`v*` release tag changes shipped product inputs:

* the Go command or internal engine and embedded Core Roster
* Go module dependencies
* release binary construction
* Homebrew or Scoop rendering

Documentation, scored evaluation results, examples, tests, and development
workflow changes still validate but do not create a product version. The
event base is the fallback before the first release tag. This roll-forward
window keeps a failed product release eligible when a later main push contains
only recovery evidence. The classifier fails closed to publication when its
base revision is unavailable. Its fixture suite covers documentation, results,
product code, roll-forward recovery, initial pushes, the major hold, and manual
dispatch.

An automatic release bumps the minor version, cross-compiles macOS, Linux, and
Windows binaries, creates the Forgejo release, uploads checksums and package
files, and updates Homebrew and Scoop when their write tokens are present.

## Major release hold

A tracked `.release-major` pauses automatic publication while a breaking stack
lands. Main validation continues. Manual workflow dispatch ignores the hold and
remains the only major-version and explicit-tag path.

For v2.0, keep the hold until the Core Roster stack, complete frontier and OSS
evidence, independent QA verdict, generated scorecard, and clean-main smoke are
all present. Then an operator dispatches the exact landed revision with a major
bump. Remove the hold in a follow-up commit only after v2.0.0 and both package
channels are verified. Removing the hold alone does not publish another
version.

## See also

* [../README.md](../README.md) - installation and product status.
* [FEATURES.md](FEATURES.md) - shipped capability inventory.
* [evaluation.md](evaluation.md) - behavioral release gates.
* [../justfile](../justfile) - local release recipes.
