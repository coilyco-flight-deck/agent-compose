---
doc_goal: Define Agent Compose consumption of AOS-hydrated local skill catalogue manifests.
---
# Local skill catalogues

Agent Compose accepts an AOS-emitted local catalogue manifest:

```yaml
skill_catalog_manifest: ~/.config/aos/catalogues.json
```

This is the Agent Compose v2 ownership boundary. AOS owns remote selection,
Git access, locking, cache freshness, offline fallback, and host paths. Agent
Compose opens only the resulting local JSON and never fetches its sources.

## Manifest contract

The document must use `aos.catalogues.v1`:

```json
{
  "format": "aos.catalogues.v1",
  "catalogues": [
    {
      "source": "owner/repo/.agents/skills@main",
      "path": "/absolute/local/catalogue",
      "commit": "0123456789abcdef0123456789abcdef01234567"
    }
  ]
}
```

Every entry needs a nonempty source, an absolute existing directory, and a
full 40- or 64-character Git object ID. Unknown fields, trailing JSON,
unsupported formats, missing paths, relative paths, and regular files fail
before roster or load-point writes.

Entries retain declaration order. Later catalogue entries win duplicate skill
names. Existing unowned files at a native load point still win over every
managed catalogue.

## Compatibility window

The v1 `remote_skill_sources` and `mcp_inventory` fields remain temporarily
available while AOS and infrastructure land their side of the ordered
migration. New convergence config uses `skill_catalog_manifest`. Agent Compose
v2 removes the network and native MCP mutation surfaces after consumers move.

## See also

* [Cascade](cascade.md) - native skill projection and ownership sidecar.
* [Remote skills](remote-skills.md) - deprecated v1 hydration behavior.
* [Native MCP](native-mcp.md) - deprecated v1 projection behavior.
