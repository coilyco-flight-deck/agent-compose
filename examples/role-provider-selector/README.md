# Role-provider selector example

This fragment admits one exact ordinary skill plus every `machine-*` skill
from a role-only provider. `path` remains relative to `projects_root` in the
host's complete `agent-compose.yaml`.

Agent Compose first validates the provider's complete catalogue. It then
requires every pattern to match at least one skill and rejects overlaps. Omit
`skills` when the role should receive the whole provider.

See [Role-scoped skill providers](../../docs/role-scoped-providers.md) for the
full configuration and trace contract.
