---
title: "Config Precedence — Continue.dev 0.9.x"
slug: config-precedence-continue-dev-0-9-x
type: concept
stack: [continue-dev, intellij]
tags: [config]
depth: 3
confidence: low
created: 2026-04-13
updated: 2026-04-13
verified: 2026-04-13
freshness_days: 365
sources:
  - url: sources/daily/2026-04-13-local-ai-continue-rag-spring.md
    accessed: 2026-04-13
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Config Precedence — Continue.dev 0.9.x

In Continue.dev version 0.9.264 (IntelliJ), configuration files are loaded in this order:

```
config.ts  >  config.json  >  config.yaml
```

The first file found takes precedence. If `config.ts` exists but fails to compile (e.g., because
`@continuedev/config-types` is not installed in `~/.continue/`), Continue silently falls back to a
**default cloud-only config** — no error is shown to the user.

## Consequences
- `config.ts` failing to compile = no local models, no requests sent to local server
- `config.json` with `baseUrl` (wrong) = "Failed to load config" error
- `config.json` with `apiBase` (correct) = config loads

## Resolution
Rename `config.ts` to `config.ts.bak` and use only `config.json` with `apiBase`.

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/decision/decision-use-config-json-over-config-ts-in-continue-dev]]
- [[notes/pitfall/issue-config-ts-silently-overrides-config-json-in-continue-dev]]
- [[notes/pitfall/issue-baseurl-vs-apibase-in-continue-dev-config]]
- [[notes/concept/continue-config-json-continue-dev-configuration]]
- [[notes/concept/dev-tools-continue-dev-and-intellij-integration-summary]]
