# Bundle protocol

Every successful composition produces one immutable, content-addressed tree.
Consumers treat the tree as opaque and enter it only through `manifest.json`.

## Tree contract

The v0.1 tree contains:

* `manifest.json` - protocol version, identity, provenance, files, and delivery.
* `trace.json` - the deterministic decision trace.
* `content/instructions.md` - canonical selected instructions.
* `content/skills/<source-id>/<content-id>/...` - canonical selected skill trees.
* `delivery/compiled.md` - present only when the adapter compiles selected skill
  bodies into one context document.

Native delivery points the manifest at `content/instructions.md` and
`content/skills`. Compiled delivery points at `delivery/compiled.md` and exposes
no skill-root entry point, but the bundle retains the canonical selected skill
trees for inspection and semantic diffing. Harness load-point paths never
appear inside the generic tree.

Every path uses slash-separated relative form. Bundle trees contain regular
files and directories only. Symlinks and paths that escape the root are invalid.

## Manifest

The stable JSON fields, source provenance, file inventory, and delivery entry
points are defined in [manifest-schema.md](manifest-schema.md).

## Identity and atomicity

The bundle id is SHA-256 over canonical JSON containing the protocol version,
normalized request facts, ordered source records, delivery metadata, and the
sorted `files` records. The identity record excludes `manifest.json` to avoid a
self-reference. The manifest carries the resulting id.

The identity excludes absolute paths, cache location, timestamps, duration,
TTY state, and cache-hit status. The CLI may report those values as runtime
telemetry, but the materializer never stores them under the bundle root.

The materializer creates a private staging directory beside the final bundle,
writes and verifies every file, then renames the completed directory atomically.
The materializer never rewrites an existing bundle id. Concurrent writers that
lose the rename race verify and reuse the winner, then discard their staging
directory. A consumer rejects an unsupported version, digest mismatch, or
missing declared entry point before agent launch.

## Compatibility

Consumers must match protocol name and v0.1 version exactly. Producers may add
new files only when old consumers can ignore them through the manifest. Any
field removal, meaning change, or entry-point change requires a new protocol
version and compatibility fixtures in each consumer.

## See also

* [decision-trace.md](decision-trace.md) - the retained explanation schema.
* [manifest-schema.md](manifest-schema.md) - manifest fields and provenance.
* [architecture.md](architecture.md) - integration boundaries.
* [contract-review.md](contract-review.md) - consumer decisions under review.
