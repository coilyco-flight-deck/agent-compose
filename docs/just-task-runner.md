# just, the task runner

Every development verb is a recipe in the repo-root [justfile](../justfile).
`just` alone lists them.

Retiring per-repo `ward exec` is
[coilysiren/inbox#366](https://forgejo.coilysiren.me/coilysiren/inbox/issues/366),
under the principle in
[#365](https://forgejo.coilysiren.me/coilysiren/inbox/issues/365): ward is
out-of-band flight control, so a repo should mention it in passing rather than
route its whole build through it. The pattern is
[agentic-os#1048](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/pulls/1048).

## What changed

All 30 verbs moved with **identical names and identical command lines**. The
`commands:` block is gone from `.ward/ward.yaml`.

Arguments pass straight through, so the `--` separator is retired:

```
just evalkit-export evaluations/pilot/ops-board-2026-08-12-regraded
just test
```

## Why `.ward/ward.yaml` still exists

It carries the `catalog:` block and nothing else. `check_catalog_block` pins
that exact path, and `catalog-trifecta` requires README, AGENTS, and FEATURES
to each link it. Both are authored upstream in agentic-os, so this repo cannot
remove the file. Tracked at
[agentic-os#1081](https://forgejo.coilysiren.me/coilyco-flight-deck/agentic-os/issues/1081).

## One line of comment per recipe

`just` reads only the **last** comment line above a recipe, so a wrapped
description silently truncates to its tail. agentic-os#1048 found this the
hard way. Keep descriptions on one line.

## What was lost, stated rather than hidden

**The clean-tree gate.** `ward exec` refused a repo verb while the working tree
was dirty, so an audit row could be reconstructed from git history. `just` has
no equivalent and this repo now has none.

Nothing here depended on it: no non-doc reference to `--audit-override-dirty`
or the gate exists, which matches what agentic-os#1048 found in its own
corpus. Recorded because it is a real reduction, not because it broke
anything.

## Related

* [Release](release.md) - the release verbs, now recipes.
