# Catalogues and bundle export

Agent Compose exposes the effective selected profile through deterministic text
and JSON catalogues. Inspection is read-only. It does not select a role,
activate a personality, change authority, or fetch a source.

## Catalogue commands

Every person-dependent command accepts `--person-source` and repeatable
`--personality-library` roots:

```text
agent-compose catalog personalities [--query <cue>] [--json]
agent-compose catalog roles [--json]
agent-compose catalog seats [--role <slug>] [--json]
agent-compose catalog expressions [--json]
```

Text output is unpaged and follows effective catalogue order. Exact normalized
personality slugs win over aliases. An ambiguous alias returns every candidate
in catalogue order.

Every JSON command emits:

```json
{
  "format": "agent-compose.catalog.v1",
  "items": []
}
```

Personality items contain `slug`, `skill`, the one-sentence skill
`description`, `aliases`, `color`, `motif`, `emblem`, `form`, `sound_mark`,
`source_library`, `digest`, and complete role `affinities`. Role items contain
`slug`, `purpose`, `skill`, role-skill provenance, `seats`, ordered
`personalities`, and `favorite_color`. Seat items contain `role` plus the full
stable seat object. Expression items are stable strings.

## Deterministic export

`agent-compose bundle export <bundle-dir> --out <file>.tar.gz` verifies the
bundle before opening the output. The exporter sorts safe slash-separated
relative paths, rejects links and non-regular entries through verification,
normalizes gzip and tar metadata, and includes `manifest.json`. Identical
verified trees produce byte-identical archives.

## Content-aware diff

`agent-compose diff <left-bundle> <right-bundle>` reports resolver-decision
changes, logical content changes, and changed bundle artifacts. Logical content
uses stable IDs and SHA-256 digests from the manifest. The effective role
skill, invariant, personality definitions, evaluation assets, copy contract,
and compact role identity metadata therefore remain visible even when a change
does not alter a resolver decision.

## See also

* [bundle-protocol.md](bundle-protocol.md) - verified tree and archive contract.
* [manifest-schema.md](manifest-schema.md) - logical content entries.
* [personality-libraries.md](personality-libraries.md) - effective graph and cue lookup.
