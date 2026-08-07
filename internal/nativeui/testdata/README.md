# Vendored Claude Code harness data

Three facts this package is coupled to, extracted from the shipped Claude Code
binary because none of them is a published contract.

* `harness-version.txt` - the version the current data came from.
* `harness-default-verbs.txt` - the default spinner verbs. Under `replace` mode a
  role verb that repeats a default reads as the default, doing no identity work,
  so `TestVerbsDoNotRepeatTheHarnessDefaults` holds that line.
* `harness-theme-tokens.txt` - every theme token name the harness knows, the
  union across all six base themes. An unknown token is dropped silently rather
  than rejected, so `TestThemeOverridesAreAcceptedByTheHarness` turns a rename
  into a failure instead of a blank role.

## Why the union rather than the dark base alone

The six bases (`dark`, `light`, `dark-ansi`, `light-ansi`, and both daltonized
variants) are keyed identically, and the mapping from base name to its object is
not recoverable from a plain read of the bundle. The union is what can be
extracted honestly. It still catches the failure this guards, a token the harness
no longer knows, because a renamed token leaves every base at once.

## Refreshing

```sh
ward exec harness-refresh
```

That runs the extraction against the Claude Code binary on `PATH` and rewrites
all three files. Review the diff: a changed verb list is ordinary upstream
churn, a changed token list may mean a role stopped looking like itself.

Set `AGENT_COMPOSE_CLAUDE_BINARY` to extract from a specific binary instead.

## How drift is noticed

`TestVendoredHarnessDataMatchesTheInstalledBinary` re-extracts and compares
content on every run, and skips when `claude` is not on `PATH`. It deliberately
does not assert the version, so a Claude Code upgrade that changes neither list
stays quiet rather than failing every checkout the moment the harness
self-updates.

The extraction is anchored on the first default verb and on a token name that
appears once per base theme. If a future bundler layout breaks those anchors the
test fails loudly rather than reporting an empty list as agreement.
