# Role skills

Each person profile owns one ordinary role skill. KDL keeps compact identity data and leaves long-form doctrine in the skill body. It may also own separate [role methods](role-methods.md) whose procedure is determined by the package's cross-role policy.

## Profile layout

A role fragment binds a stable skill id:

```kdl
role "strats" {
    display-name "Portfolio Strategist"
    purpose "Decide where Kai should invest attention across her real portfolio."
    skill "role-strats"
    personality "curious" "skeptical" "grounded" "decisive"
}
```

The body lives at `roles/strats/SKILL.md` with ordinary skill frontmatter. It
requires `role-<slug>`, matching metadata, and at least three paragraphs.

V1.x packages may retain an inline `briefing`. The compatibility adapter
projects it as an in-memory `role-<slug>` skill. A role cannot declare both
forms, and the adapter never writes a second mutable source tree.

## Progressive disclosure

An assigned native bundle materializes the selected role skill, its role
methods, and every personality skill in its ordered meld. Startup instructions carry the fixed
role bootstrap and a compact identity card. The card retains purpose, seats,
pronouns, personality emblems, glyphs, motifs, colors, one-sentence cues, the
melded favorite color, and exact skill ids. Compact fields use ` // `.

The native roster installs role and personality skills for discovery without
loading their bodies globally. After role selection, the agent reads that role
skill and its complete meld before acting.

Compiled delivery has no native skill loader. It appends the selected role
skill first, then the active personality skills, role methods, and capability skills, so its
behavioral content stays equivalent without emitting unusable pointers.

## Authority boundary

Role skills define identity and the feedback loop a role owns. Capability
providers define task methods. Ward and guarded runtime policy define
executable authority.

Content exclusively owns recommendations about human communication, including
wording, tone, framing, timing, channel, reply strategy, and editorial fitness.
Other roles stop at a factual handoff. Community owns community state,
moderation context, durable feedback, and authorized interaction. Sending or
publishing an approved artifact still requires separate authorization.

The Engineer role owns repository-proven reusable software. Ops owns controlled
running-system change, live verification, and rollback. QA remains read-only
unless the runtime explicitly grants an enforced disposable fixture mode. No
role skill grants commands, credentials, mounts, network access, deployment,
model selection, or permission.

Designer (`design`) owns bounded, effect-tested experience definition and
page-level work in existing graphical web apps. Static pages, focused routes,
navigation, presentation, metadata, accessibility, and tests qualify when
business rules and data flow stay unchanged. User-supplied labels and list
items may be applied directly with native semantic HTML. Agent-authored meaning,
content strategy, behavior, data flow, stateful workflows, validation,
permissions, analytics, system routing, runtime data, terminal or procedural
experiences, infrastructure, release, and live operations do not. Content owns
wording and strategy when copy is not supplied. Designer owns hierarchy,
accessibility, and integration.

Content Manager (`content`) may land content-only repository
changes, including human-facing literals embedded in code. The exception
requires unchanged control flow, state, schemas, structured contracts, and
executable behavior. Mixed content and behavior returns to Engineer. Content
also accepts factual handoffs from every other role and returns the recommended
human-facing artifact or strategy without inferring authority to expose it
externally.

## See also

* [native-adaptation.md](native-adaptation.md) - inference and switching rules.
