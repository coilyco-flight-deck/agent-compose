# Native MCP projection

Agent-compose projects one canonical mcporter inventory into the native Claude
Code and Codex MCP registries. Mcporter remains the inventory owner and CLI
fallback. Harness choice no longer changes which configured servers exist.

## Configuration

Point the cascade config at the canonical source:

```yaml
mcp_inventory: /path/to/config/mcporter.json
```

The inventory must set top-level `imports: []`. Mcporter otherwise imports the
native harness registries back into its merged view, creating a reverse
configuration loop that lets stale native-only servers reappear.

## Projection contract

Bare `acompose` runs the projector after doctrine and native-skill convergence.
`agent-compose mcp --inventory <path>` exposes the same implementation to host
configuration and warded startup.

The projector:

* copies the source to `~/.mcporter/mcporter.json` for CLI fallback
* replaces Claude Code's user-scope `mcpServers` map while preserving unrelated
  Claude state
* replaces Codex's bracketed managed MCP block while preserving unrelated
  settings and explicit built-in disables
* reports drift without writing when `--check` is present

The hard server-set projection is intentional. Removing an entry from the
canonical inventory removes it from both native registries on the next
convergence.

## See also

* [cascade.md](cascade.md) - host doctrine, skills, and MCP convergence.
* [projection.md](projection.md) - immutable bundle load-point projection.
