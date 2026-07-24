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

        You own validation and the resolved repository landing workflow.
        """
}
```

The loader trims outer whitespace while preserving internal paragraph breaks.
It rejects a missing, empty, or duplicate briefing. All ten embedded roles
carry exactly two paragraphs.

## Authoring shape

The first paragraph establishes mission and default work. The second
establishes completion, handoff, and escalation boundaries. The engineer
briefing explicitly makes that role exclusively unattended. Other briefings
define the interactive or autonomous posture their work requires.

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
* component personality names, skills, and individual colors
* melded favorite color plus concise inspiration and appearance metadata

The host cascade carries all ten named roles into global context. A bundle
places the selected role's compact metadata before its briefing. Long catalogue
prose and citations stay in the documentation surface.

## Ownership boundaries

Agent-compose owns these public-safe role charters. AOS owns reusable ordinary
and role-composed task doctrine. Ward owns execution permissions, guardfiles,
models, reasoning effort, and runtime authority.

Briefings therefore state stable role behavior and refer to the repository's
resolved workflow. They do not copy mutable repository commands or live
authority policy.

## See also

* [person-contract.md](person-contract.md) - complete embedded person schema.
* [integration.md](integration.md) - host delivery and self-selection.
* [role-selection.md](role-selection.md) - assignment precedence and locks.
* [../internal/person/person.kdl](../internal/person/person.kdl) - canonical role text.
