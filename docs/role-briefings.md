# Role skills

The selected person profile owns one ordinary role skill for every role.
Structured KDL chooses the skill and retains compact identity data. The skill
body is the sole canonical long-form mission, operating loop, ownership,
completion, handoff, and escalation doctrine.

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

The body lives at `roles/strats/SKILL.md` with ordinary skill frontmatter.
The loader requires `role-<slug>`, matching frontmatter, a nonempty
description, and at least three substantive body paragraphs. Missing,
malformed, or mismatched skills fail before composition.

V1.x packages may retain an inline `briefing`. The compatibility adapter
projects it as an in-memory `role-<slug>` skill. A role cannot declare both
forms, and the adapter never writes a second mutable source tree.

## Progressive disclosure

An assigned native bundle materializes the selected role skill and every
personality skill in its ordered meld. Startup instructions carry the fixed
role bootstrap and a compact identity card. The card retains purpose, seats,
pronouns, personality emblems, glyphs, motifs, colors, one-sentence cues, the
melded favorite color, and exact skill ids. Compact fields use ` // `.

The native roster installs all role and personality skills for discovery but
does not import their full bodies into global startup context. After inference
or an allowed role switch, the agent reads the selected role skill and its
complete meld before acting.

Compiled delivery has no native skill loader. It appends the selected role
skill first, then the active personality skills and capability skills, so its
behavioral content stays equivalent without emitting unusable pointers.

## Authority boundary

Role skills define identity and the feedback loop a role owns. Capability
providers define task methods. Ward and guarded runtime policy define
executable authority.

The Engineer role owns repository-proven reusable software. Ops owns controlled
running-system change, live verification, and rollback. QA remains read-only
unless the runtime explicitly grants an enforced disposable fixture mode. No
role skill grants commands, credentials, mounts, network access, deployment,
model selection, or permission.

Designer (`design`) owns experience definition everywhere and may implement and
land an effect-tested, bounded page-level experience in an existing graphical
web application. Static or content-driven pages, focused routes, navigation
placement, page copy, static public display data, presentation, metadata,
ordinary accessibility, and focused tests qualify when business rules and
application data flow stay unchanged. Routing-system architecture, runtime
data, stateful workflows, terminal experiences, procedural games,
infrastructure, deployment, release, and live operations remain outside that
boundary. Definition separates sourced facts from unresolved decisions, keeps mechanics pending, and distinguishes implemented, verified, and delivered work. Content Manager owns verified copy, while Designer owns hierarchy, accessibility, and integration.

Content Manager (`content`) may implement and land content-only repository
changes, including human-facing literals embedded in code. The exception
requires unchanged control flow, state, schemas, structured contracts, and
executable behavior. Mixed content and behavior returns to Engineer.

## See also

* [native-adaptation.md](native-adaptation.md) - inference and switching rules.
* [personality-libraries.md](personality-libraries.md) - profile and library boundaries.
* [person-snapshot.md](person-snapshot.md) - role skill provenance projections.
* [role-skill-context-budget.md](role-skill-context-budget.md) - measured startup reduction.
