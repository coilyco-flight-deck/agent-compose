# Bundle protocol

Every successful composition produces one immutable tree. Consumers enter it
through `manifest.json` and otherwise treat the tree as opaque.

## Tree contract

The v0.1 tree contains:

* `manifest.json` - what was composed and the delivery entry points.
* `trace.json` - the plain-language decision trace.
* `content/instructions.md` - canonical selected instructions.
* `content/skills/<source-id>/<skill>/...` - canonical selected skill trees.
* `delivery/compiled.md` - present only when the adapter compiles selected
  skill bodies into one context document. The canonical skill trees stay
  beside it for inspection and diffing.

Every path uses slash-separated relative form. Bundle trees contain regular
files and directories only. Symlinks and paths that escape the root are
invalid. Harness load-point paths never appear inside the generic tree.

## Immutability and atomicity

The materializer writes into a private staging directory beside the final
location, verifies the tree is complete, then renames it into place
atomically. A finished bundle is never rewritten in place - refresh produces
a new tree and swaps it in. A failed refresh never partially replaces a
known-good bundle; the previous bundle stays live until the replacement is
complete.

`agent-compose verify <bundle-dir>` exposes that same read-only consumer
check. It rejects links and special files, unsafe or missing entry points,
unknown delivery modes, invalid traces, and any bundle whose identity trees do
not exactly match every skill selected by its trace. Cache hits pass
verification again before reuse, so the presence of `manifest.json` alone
never blesses a tree.

Runtime telemetry - durations, cache location, cache-hit status, terminal
state - never lands under the bundle root.

## See also

* [manifest-schema.md](manifest-schema.md) - manifest fields.
* [projection.md](projection.md) - placing bundle content at harness load points.
* [decision-trace.md](decision-trace.md) - the retained explanation data.
* [architecture.md](architecture.md) - integration boundaries.
* [contract-review.md](contract-review.md) - review decisions of record.
