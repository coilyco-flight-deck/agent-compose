# KDL contracts

Agent-compose uses KDL for human-authored requests and policy. Every root node
declares `schema-version=1`. Parsers fail on unknown nodes, duplicate scalar
facts, unsupported versions, missing required sources, or an empty selection.

## Compose request

A request contains one nested cli-guard agent claim, one personality, one mode,
at least one privacy scope, ordered target repositories, and source locators:

```kdl
compose schema-version=1 {
    agent-claim schema-version=1 {
        agent "codex"
        role "engineer" domain="context"
        model "gpt-5.6-terra"
        model-class "frontier"
        harness "codex"
        reasoning-effort "high"
    }
    personality "curious"
    mode "interactive"
    privacy-scope "public"
    target-repo "coilyco-flight-deck/agent-compose" declaration="repo-agent-compose.kdl"
    source "aos-public" kind="aos" declaration="source-public.kdl" required=#true
}
```

The request requires one context role. An authority role fails validation.
Personality, privacy scopes, targets, and source locators remain
agent-compose-owned fields outside the shared claim. Permitted external source
kinds are `aos` and `overlay`. Request order is significant within each source
kind, but absolute declaration paths are locator data rather than identity. A
locator's slug, source id, and kind must match its declaration root.

## Repository declaration

A product repo declares identity without provider paths or skill names:

```kdl
repo "coilyco-flight-deck/agent-compose" schema-version=1 {
    language "go" rank=1
    product "cli" rank=1
}
```

A repo may declare at most two languages and two products. Ranks are unique
positive integers where lower wins. Explicit declarations outrank inference.

## AOS source declaration

An AOS source owns reusable content and capability mappings:

```kdl
source "aos-public" schema-version=1 kind="aos" {
    privacy-scope "public"
    instruction "foundation" path="content/foundation.md" tier="foundational"
    capability "product:cli" {
        skill "fixture-review" path="skills/fixture-review"
    }
}
```

Each external source declares exactly one privacy scope and is eligible only
when the request names that scope. Paths are relative to the declaration. The
resolver hashes referenced bytes before selection and records normalized
provenance. A required missing source fails. An optional missing source produces
a warning and trace exclusion.
Source declarations may use only relative, clean paths beneath their source
root. Directory references contain recursively sorted regular files. Symlinks
and paths that escape the source root fail validation.

## See also

* [architecture.md](architecture.md) - fact and policy ownership.
* [bundle-protocol.md](bundle-protocol.md) - machine-readable output.
* [person-contract.md](person-contract.md) - embedded personal policy.
* [contract-review.md](contract-review.md) - cross-product decisions and gates.
