# Role-provider selector example

This fragment admits one exact ordinary skill plus every `machine-*` skill
from a role-only skill-provider repository. `path` remains relative to
`projects_root` in the trusted root's complete `.agents/roles.kdl`.

Agent Compose first validates the provider's complete catalogue. It then
requires every pattern to match at least one skill and rejects overlaps. Use
`skill "*"` when the role should receive the whole ordinary catalogue.

See [Role-scoped skill providers](../../docs/role-selection.md) for the
full configuration and trace contract.
