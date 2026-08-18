---
title: "~/.continue/config.json — Continue.dev Configuration"
slug: continue-config-json-continue-dev-configuration
type: concept
stack: [continue-dev, intellij]
tags: [local-ai, config]
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

# ~/.continue/config.json — Continue.dev Configuration

The active Continue.dev configuration file (JSON format). Key fields:
- `provider`: `"openai"` (correct for 0.9.x with local server)
- `apiBase`: `"http://localhost:8079/v1"` — points directly at llama.cpp (not LiteLLM)
- `model`: `"Qwen3-4B-Q4_K_M.gguf"` — must match exactly what llama.cpp reports

## Config Precedence (Continue 0.9.264)
`config.ts` > `config.json` > `config.yaml`

`config.ts` was renamed to `.bak` to ensure `config.json` is loaded. See [[notes/concept/config-precedence-continue-dev-0-9-x]].

## Known Issues
- `baseUrl` is rejected; must be `apiBase` — see [[notes/pitfall/issue-baseurl-vs-apibase-in-continue-dev-config]]
- `config.ts` overrode this file silently — see [[notes/pitfall/issue-config-ts-silently-overrides-config-json-in-continue-dev]]

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/concept/config-precedence-continue-dev-0-9-x]]
- [[notes/decision/decision-use-config-json-over-config-ts-in-continue-dev]]
- [[notes/decision/decision-continue-dev-points-directly-at-llama-cpp-port-8079]]
- [[notes/concept/dev-tools-continue-dev-and-intellij-integration-summary]]
