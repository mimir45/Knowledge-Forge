---
title: "Issue: config.ts Silently Overrides config.json in Continue.dev"
slug: issue-config-ts-silently-overrides-config-json-in-continue-dev
type: pitfall
stack: [continue-dev]
tags: [config, issue]
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

# Issue: config.ts Silently Overrides config.json in Continue.dev

## Symptom
Continue.dev sends no requests to the local server even after config.json is corrected.
Config error is gone but model dropdown shows only cloud models.

## Root Cause
Continue.dev 0.9.x loads `config.ts` before `config.json`. If `config.ts` exists but fails to
compile (missing `@continuedev/config-types` package), Continue silently falls back to default
cloud-only config. No error is surfaced to the user.

## Fix
Rename `config.ts` to `config.ts.bak` to force Continue to use `config.json`.

## Affected Files
- `~/.continue/config.ts` → `~/.continue/config.ts.bak`

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/concept/config-precedence-continue-dev-0-9-x]]
- [[notes/decision/decision-use-config-json-over-config-ts-in-continue-dev]]
- [[notes/concept/continue-config-json-continue-dev-configuration]]
