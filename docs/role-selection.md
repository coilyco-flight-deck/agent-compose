# Role selection

Agent-compose distinguishes an inferred native role from a caller-assigned
role. Task shape may help an unassigned agent select its initial role, but task
shape never overrides an assignment that a caller already made.

## Inferred native roles

The host cascade exposes the complete named-seat roster. When no launch context
assigns a role, an unassigned native agent uses the initial substantive request
as a soft signal and selects the closest role. The agent records that role as
inferred and loads its role skill plus complete ordered personality meld. Its
role methods remain lazy until a matching task triggers them.

In an eligible directly steered session, an explicit user request naming a
valid rendered role slug switches immediately without a second confirmation.
The agent loads the target charter and meld, announces the role, and stops
acting from the prior charter. The switched role remains inferred and persists
until another explicit switch or session end. This permits later switches and
a return to an earlier role.

An agent-proposed switch requires a separate confirmation. An unknown target
fails with the available rendered role slugs. The complete eligibility and
confirmation rules live in [native adaptation](native-adaptation.md).

## Caller-assigned roles

A compose request names exactly one role. The resulting bundle declares that
role authoritative and fixed for the session before presenting its identity card.
The agent does not change roles because a task resembles another role. The
agent does not activate, blend, or adopt another role's briefing or personality
set. When the user asks to switch, the agent rejects the request and directs
the caller to launch a new bundle with the different role.

## Enforcement boundary

The resolver excludes inactive role, personality, role-method, and composed capability
skills from the bundle. Container-home projection receives that selected bundle
and does not run the host roster cascade. These structural limits back the
instruction contract without turning personality into an authority boundary.

Native switching changes only the active charter and meld carried by the
already loaded roster. It does not change the harness, model, tools,
permissions, credentials, or executable authority.

## See also

* [role-briefings.md](role-briefings.md) - role charter and roster rendering.
* [role-methods.md](role-methods.md) - role-bound task methods.
* [native-adaptation.md](native-adaptation.md) - native switch eligibility.
* [integration.md](integration.md) - host cascade and container projection.
* [kdl-contracts.md](kdl-contracts.md) - caller-selected compose inputs.
