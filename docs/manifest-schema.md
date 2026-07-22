# Bundle manifest schema

`manifest.json` names what was composed and where the entry points are:

```json
{
  "format": "agent-compose.bundle",
  "role": "engineer",
  "personality": "curious",
  "density": "full",
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

That is the whole schema. There is no subject block, no digest list, and no
schema-version ceremony - this is a personality engine, not a security
boundary. `format` is a plain marker so a consumer knows what it is reading,
not the anchor of a compatibility regime. A consumer that wants integrity
checking can hash the tree itself.

`sources` records which places content came from, by stable id, so the trace
and a human reader can refer to them. Locators and absolute paths never
appear.

## See also

* [bundle-protocol.md](bundle-protocol.md) - tree layout and atomicity.
* [decision-trace.md](decision-trace.md) - retained decision evidence.
* [contract-review.md](contract-review.md) - review decisions of record.
