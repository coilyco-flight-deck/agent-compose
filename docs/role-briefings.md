# Role briefings

The embedded person policy gives every canonical role an unconditional
operating charter. The charter orients the agent before task-specific AOS
knowledge activates.

## Schema

`purpose` remains the concise role label. `briefing` carries the long-form
orientation:

```kdl
role "engineer" {
    purpose "Write code, merge code, stay focused on your goal."
    briefing """
        You are an engineer. You turn a defined goal into working code.

        You inspect the surrounding system and exercise the risky paths.

        You own validation and the resolved repository landing workflow.
        """
}
```

The loader trims outer whitespace while preserving internal paragraph breaks.
It rejects a missing, empty, or duplicate briefing. All twelve embedded roles
carry at least three substantial paragraphs.

## Authoring shape

The first paragraph establishes mission and default work. The second defines
the role's operating method and evidence discipline. The third establishes
completion, handoff, and escalation boundaries. The engineer briefing
explicitly makes that role exclusively unattended. Other briefings define the
interactive or autonomous posture their work requires.

A briefing may shape attention, operating loop, and completion posture. It
does not grant permissions, select a model, choose a harness, or weaken an
applicable safety boundary.

## Bundle delivery

Every role-specific bundle begins with the caller-selected role assignment,
heading, purpose, and complete briefing. The
[role-selection contract](role-selection.md) makes that assignment authoritative
for the session. Native-skill delivery projects the document to the harness
instruction load point. Compiled delivery keeps the same role section first.

The bundle remains harness-blind. It carries the selected role charter but no
harness identity, model choice, reasoning effort, or runtime authority. The
projection layout decides whether the document lands at `AGENTS.md`,
`CLAUDE.md`, or another harness-native instruction path.

## Roster delivery

The roster renders each named-seat role in this order:

* role identity, concise purpose, and long-form briefing
* harness-specific names and pronouns
* component personalities, colors, and aligned inspiration metadata
* melded favorite color

The host cascade carries all twelve named roles into global context. A bundle
places selected identity fit, achievement, impact, and citation metadata before
its briefing. The appearance catalogue remains snapshot-only.

## Ownership boundaries

Agent-compose owns these public-safe role charters. AOS owns reusable ordinary
and role-composed task doctrine. Ward owns execution permissions, guardfiles,
models, reasoning effort, and runtime authority.

Briefings therefore state stable role behavior and refer to the repository's
resolved workflow. They do not copy mutable repository commands or live
authority policy.

## See also

* [role-selection.md](role-selection.md) - assignment precedence and locks.
* [../internal/person/roles](../internal/person/roles) - canonical role text.
