---
title: "Dev Tools — Continue.dev and IntelliJ Integration Summary"
slug: dev-tools-continue-dev-and-intellij-integration-summary
type: concept
stack: [continue-dev, intellij]
tags: [config, ide, local-ai]
depth: 4
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

# Dev Tools — Continue.dev and IntelliJ Integration Summary

## Continue.dev Version
Continue.dev **0.9.264** (IntelliJ plugin) — NOT 1.0.x.

## Active Configuration
File: `~/.continue/config.json`

```json
{
  "models": [{
    "provider": "openai",
    "apiBase": "http://localhost:8079/v1",
    "model": "Qwen3-4B-Q4_K_M.gguf"
  }],
  "contextProviders": [{
    "name": "rag",
    "params": { "apiBase": "http://localhost:5001" }
  }]
}
```

## Config Precedence (0.9.x)
```
config.ts  >  config.json  >  config.yaml
```
`config.ts` was renamed to `.bak` to avoid silent override. See [[notes/concept/config-precedence-continue-dev-0-9-x]].

## Critical Field Names
| Wrong | Correct |
|-------|---------|
| `baseUrl` | `apiBase` |
| `requestOptions.baseUrl` | `apiBase` |

## Qwen3 Thinking Mode
Qwen3 thinking mode must be disabled. Continue reads only `content`, not `reasoning_content`.
- Fix: `--reasoning off` in `llama-server.sh`
- See [[notes/concept/thinking-mode-qwen3-reasoning-content-issue]], [[notes/pitfall/issue-qwen3-blank-response-in-continue-dev]]

## Open Issue: Requests Not Sent
Even with valid config and reachable llama.cpp, Continue may send zero requests in IntelliJ.
Root cause unknown — possible hub/auth UI issue. Investigate:
- Continue plugin logs in IntelliJ (Help → Show Log)
- Whether hub authentication intercepts requests
- Test with VS Code Continue to isolate if IntelliJ-specific

## Build-Helper Maven Plugin
`build-helper-maven-plugin` is required to register generated sources so IntelliJ can statically
discover them during Maven import. The generated path is TWO levels deep:
`target/generated-sources/openapi/src/main/java`

Always run `./mvnw generate-sources` before IntelliJ Maven reload.

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/concept/continue-config-json-continue-dev-configuration]]
- [[notes/concept/llama-server-sh-llama-cpp-startup-script]]
- [[notes/concept/config-precedence-continue-dev-0-9-x]]
- [[notes/concept/thinking-mode-qwen3-reasoning-content-issue]]
- [[notes/decision/decision-use-config-json-over-config-ts-in-continue-dev]]
- [[notes/decision/decision-continue-dev-points-directly-at-llama-cpp-port-8079]]
- [[notes/pitfall/issue-baseurl-vs-apibase-in-continue-dev-config]]
- [[notes/pitfall/issue-config-ts-silently-overrides-config-json-in-continue-dev]]
- [[notes/pitfall/issue-qwen3-blank-response-in-continue-dev]]
