# Role selection

Agent-compose distinguishes an unassigned native session from a role-bound
session. Task shape may help an unassigned agent select a role, but task shape
never overrides an assignment that a caller already made.

## Unassigned native sessions

The host cascade exposes the complete named-seat roster. When no launch context
assigns a role, an unassigned native agent uses the initial substantive request
as a soft signal and selects the closest role. The agent makes that selection
once, keeps it for the session, and activates only the selected role's briefing
and personality definitions.

## Role-bound sessions

A compose request names exactly one role. The resulting bundle declares that
role authoritative and fixed for the session before presenting its briefing.
The agent does not change roles because a task resembles another role. The
agent does not activate, blend, or adopt another role's briefing or personality
set. The caller launches a new bundle to assign a different role.

## Enforcement boundary

The resolver excludes inactive personality skills and inactive role-composed
skills from the bundle. Container-home projection receives that selected bundle
and does not run the host roster cascade. These structural limits back the
instruction contract without turning personality into an authority boundary.

## See also

* [role-briefings.md](role-briefings.md) - role charter and roster rendering.
* [integration.md](integration.md) - host cascade and container projection.
* [kdl-contracts.md](kdl-contracts.md) - caller-selected compose inputs.
