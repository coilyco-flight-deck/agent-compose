# Bundle manifest schema

`manifest.json` names what was composed and where the entry points are:

```json
{
  "format": "agent-compose.bundle",
  "role": "engineer",
  "model_class": "frontier",
  "personalities": ["curious", "grounded", "meticulous"],
  "color": "#90a66a",
  "sources": ["person:kai", "aos-public"],
  "delivery": {
    "mode": "native-skills",
    "instructions": "content/instructions.md",
    "skills_root": "content/skills"
  }
}
```

A compiled bundle replaces `skills_root` with `compiled_context` pointing at
`delivery/compiled.md` while the canonical skill trees stay in the tree for
inspection. Every named entry point must exist inside the bundle.

That is the whole schema. There is no subject block, digest list, or
schema-version ceremony. `format` is a plain marker, not a trust or
cryptographic boundary. The built-in verifier checks structural integrity:
safe relative entry points, regular files and directories only, complete
delivery data, and one identity tree per trace-selected skill. A consumer that
needs content authentication still hashes or signs the tree itself.

`sources` records which places content came from, by stable id, so the trace
and a human reader can refer to them. Locators and absolute paths never
appear. `personalities` preserves the role's declaration order. `color` is
their melded favorite, derived from every component color.
[person-contract.md](person-contract.md) owns the legibility and blend rules.
`model_class` records the caller fact used for per-skill context pruning. It
does not identify or route a runtime model.

## See also

* [bundle-protocol.md](bundle-protocol.md) - tree layout and atomicity.
* [decision-trace.md](decision-trace.md) - retained decision evidence.
* [contract-review.md](contract-review.md) - review decisions of record.
