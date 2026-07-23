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

Every role-specific bundle begins its canonical instruction document with the
selected role's heading, purpose, and complete briefing. Native-skill delivery
projects that document directly to the harness instruction load point.
Compiled delivery keeps the same role section at the beginning of its combined
context document.

The bundle remains harness-blind. It carries the selected role charter but no
harness identity, model choice, reasoning effort, or runtime authority. The
projection layout decides whether the document lands at `AGENTS.md`,
`CLAUDE.md`, or another harness-native instruction path.

## Roster delivery

The roster renders each named-seat role in this order:

* concise heading and purpose
* long-form briefing
* harness-specific names and pronouns
* melded personalities, definitions, and favorite color

The host cascade carries the complete named-seat roster into global harness
context. Each agent self-selects using its harness and role, then adopts only
that role's briefing. Canonical roles without named seats remain in the person
model but do not render into the dispatch table.

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
* [../internal/person/person.kdl](../internal/person/person.kdl) - canonical role text.
