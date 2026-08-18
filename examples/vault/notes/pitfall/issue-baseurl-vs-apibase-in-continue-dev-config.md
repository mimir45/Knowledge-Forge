---
title: "Issue: baseUrl vs apiBase in Continue.dev Config"
slug: issue-baseurl-vs-apibase-in-continue-dev-config
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

# Issue: baseUrl vs apiBase in Continue.dev Config

## Symptom
Continue.dev shows "Failed to load config" with error: `Use apiBase or requestOptions.baseUrl, not baseUrl`

## Root Cause
The config file used `baseUrl` which is not a valid field in Continue.dev. The correct field name
is `apiBase`.

## Fix
Replace `baseUrl: "http://localhost:8079/v1"` with `apiBase: "http://localhost:8079/v1"`.

## Affected Files
- `~/.continue/config.json` (or whichever config file is active)

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/concept/config-precedence-continue-dev-0-9-x]]
- [[notes/concept/continue-config-json-continue-dev-configuration]]
