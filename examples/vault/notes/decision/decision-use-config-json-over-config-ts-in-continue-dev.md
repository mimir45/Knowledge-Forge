---
title: "Decision: Use config.json Over config.ts in Continue.dev"
slug: decision-use-config-json-over-config-ts-in-continue-dev
type: decision
stack: [continue-dev]
tags: [config, decision]
depth: 3
confidence: high
created: 2026-04-13
updated: 2026-04-13
verified: 2026-04-13
freshness_days: 730
sources:
  - url: sources/daily/2026-04-13-local-ai-continue-rag-spring.md
    accessed: 2026-04-13
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Decision: Use config.json Over config.ts in Continue.dev

**Decision:** Rename `config.ts` to `config.ts.bak`; use only `config.json` for Continue.dev
configuration.

**Rationale:** `config.ts` has higher precedence than `config.json`. If `config.ts` fails to
compile (due to missing `@continuedev/config-types` package), Continue silently falls back to
cloud-only defaults. `config.json` is plain JSON — no compilation step, no silent failures.

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/concept/config-precedence-continue-dev-0-9-x]]
- [[notes/pitfall/issue-config-ts-silently-overrides-config-json-in-continue-dev]]
- [[notes/concept/continue-config-json-continue-dev-configuration]]
