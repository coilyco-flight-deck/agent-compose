# Personality identity primitives

The embedded person source owns a renderer-neutral identity record for every
personality. Web, mobile, terminal, audio, and generated-art consumers project
the same semantics without making agent-compose own their presentation.

## Emblem

Each `emblem` supplies three equivalent fallbacks:

```kdl
emblem {
    name "compass"
    emoji "🧭"
    glyph "⌖"
}
```

`name` is safe in plain text, `emoji` is the rich mark, and `glyph` is the
compact monochrome mark. The loader rejects incomplete or duplicate emblems.

## Motif

`motif` is one lowercase semantic token such as `map-paper` or `moss`.
Renderers may interpret it as material, texture, pattern, or environmental
association. It is not a CSS class or an asset path.

## Procedural form

Each `form` declares a `silhouette`, `geometry`, and `motion` token. Together
they are the stable shape language from which a renderer generates an avatar,
sprite, overlay figure, or other representation. The form is the agent's
representation, not a separate pet or observability source.

Every renderer uses this fixed expression vocabulary:

* `available`
* `listening`
* `thinking`
* `acting`
* `waiting-for-human`
* `blocked`
* `completed`
* `failed`
* `offline`

Expressions communicate state supplied by the owning runtime. Agent-compose
defines the vocabulary but never infers live state.

## Sound mark

Each `sound-mark` declares `timbre`, `contour`, and `pulse` tokens. These form a
short semantic identity seed for notifications or conversation entry. A
renderer may synthesize or map the mark to an asset. Agent-compose ships no
audio files, playback behavior, volume policy, or event routing.

## Projection

The complete record and expression vocabulary ship in person snapshot schema
v2 and palette schema v2. Bundle role metadata stays compact and text-first.
Permissions, routing, model choice, and runtime authority remain outside the
identity contract.
