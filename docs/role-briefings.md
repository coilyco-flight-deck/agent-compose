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

Designer (`design`) owns experience definition everywhere and may implement and land only
effect-tested visual presentation changes in an existing graphical web
application. Behavior, semantics, data, generated systems, terminal
experiences, games, infrastructure, deployment, and live verification remain
outside that exception.

## See also

* [native-adaptation.md](native-adaptation.md) - inference and switching rules.
* [personality-libraries.md](personality-libraries.md) - profile and library boundaries.
* [person-snapshot.md](person-snapshot.md) - role skill provenance projections.
* [role-skill-context-budget.md](role-skill-context-budget.md) - measured startup reduction.
