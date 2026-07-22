# Bundle manifest schema

`manifest.json` uses stable JSON field names:

```json
{
  "protocol": "agent-compose.bundle",
  "version": "0.1",
  "bundle_id": "sha256:<digest>",
  "subject": {
    "schema_version": 1,
    "agent": "codex",
    "roles": [{"domain": "context", "name": "engineer"}],
    "model": "gpt-5.6-terra",
    "model_class": "frontier",
    "harness": "codex",
    "reasoning_effort": "high"
  },
  "profile": {"personality": "curious", "mode": "interactive"},
  "scopes": ["public"],
  "targets": ["coilyco-flight-deck/agent-compose"],
  "sources": [
    {"id": "person:kai", "kind": "person", "schema_version": 1,
     "declaration_sha256": "<digest>", "content_sha256": "<digest>"},
    {"id": "repo:coilyco-flight-deck/agent-compose", "kind": "repo",
     "schema_version": 1, "declaration_sha256": "<digest>",
     "content_sha256": "<digest>"},
    {"id": "aos-public", "kind": "aos", "schema_version": 1,
     "declaration_sha256": "<digest>", "content_sha256": "<digest>"}
  ],
  "files": [
    {"path": "content/instructions.md", "sha256": "<digest>",
     "media_type": "text/markdown"},
    {"path": "content/skills/aos-public/fixture-review/SKILL.md",
     "sha256": "<digest>", "media_type": "text/markdown"},
    {"path": "trace.json", "sha256": "<digest>",
     "media_type": "application/json"}
  ],
  "delivery": {"adapter": "codex", "mode": "native-skills",
               "instructions": "content/instructions.md",
               "skills_root": "content/skills"},
  "trace": {"path": "trace.json", "sha256": "<digest>"}
}
```

The subject contains only a context role. Ward keeps an authority role in its
independent mount, where the matching name cannot become a grant through this
record.

## Sources and files

Sources appear as the embedded person, target repos, admitted overlays, then AOS
sources in deterministic precedence order. Every target repo receives a source
record.

`declaration_sha256` hashes declaration bytes. `content_sha256` hashes canonical
JSON containing the declaration and each declared or discovered source file the
resolver considers, represented by relative path and digest. Directory
references expand to sorted regular files. Locators and absolute paths never
enter either digest.

`files` lists every regular file under the bundle root except `manifest.json`,
sorted by path. Each record carries a digest and media type. Delivery names only
file entry points present in `files`. A directory entry point must prefix at
least one listed file and is omitted when empty. A compiled adapter uses
`compiled_context` instead of `skills_root` and names `delivery/compiled.md`.

## See also

* [bundle-protocol.md](bundle-protocol.md) - tree identity and atomicity.
* [decision-trace.md](decision-trace.md) - retained decision evidence.
* [contract-review.md](contract-review.md) - consumer decisions under review.
