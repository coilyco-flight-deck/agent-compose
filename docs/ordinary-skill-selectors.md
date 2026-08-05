# Ordinary-skill selectors

A provider declaration in `.agents/roles.kdl` may bound its ordinary
`.agents/skills` catalogue with `skill` children:

```kdl
providers {
    provider hardware path="example/hardware-knowledge" {
        skill "compute-stack"
        skill "machine-*"
    }
}
roles {
    role engineer {
        use-provider hardware required=#true
    }
}
```

Patterns use Go path-match syntax. A literal is exact and `*`, `?`, or bracket
forms provide glob matching within one skill ID. Omitting `skills` admits the
whole provider.

## Fail-closed validation

The KDL and generated-JSON loaders reject an empty pattern or malformed glob.
The transitional YAML loader also rejects an explicit empty selector. At
composition time, every pattern must
match at least one ordinary skill. No skill may match two configured patterns.
Unmatched or overlapping patterns fail without producing a bundle.

Agent Compose loads and validates the provider's complete ordinary and
composed catalogues before filtering. A selector therefore cannot hide a
malformed source. Selection retains the catalogue's lexical order instead of
pattern order.

## Evidence and budget

The provider report records the patterns and admitted catalogue fraction.
Selected skill decisions carry that selector outcome. Skills outside the
slice remain explicit excluded decisions with a selector reason, so
`agent-compose describe --why skill:<id>` explains why they did not enter the
bundle.

Context bytes and approximate tokens measure only selected trees. Native and
staged projection consume the same immutable bundle and therefore retain the
same selector evidence and budget across Claude, Codex, Goose, and OpenCode.

See the [role-provider example](../examples/role-provider-selector/README.md)
for a minimal configuration fragment.
