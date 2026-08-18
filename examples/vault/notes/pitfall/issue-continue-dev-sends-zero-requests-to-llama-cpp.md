---
title: "Issue: Continue.dev Sends Zero Requests to llama.cpp"
slug: issue-continue-dev-sends-zero-requests-to-llama-cpp
type: pitfall
stack: [continue-dev, intellij, llama-cpp]
tags: [debugging, unresolved]
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

# Issue: Continue.dev Sends Zero Requests to llama.cpp

## Status

**Unresolved** — no requests reach llama.cpp despite valid configuration.

## Symptom

With a correct `config.json`, valid `apiBase: "http://localhost:8079/v1"`, and llama.cpp
reachable at port 8079, Continue.dev in IntelliJ sends zero HTTP requests when triggered.
Config loads without error; no network traffic observed.

## Root Cause Hypotheses

1. **Hub/auth UI interception** — Continue.dev 0.9.x may route requests through the hub even
   for local models if a user session is active
2. **IntelliJ-specific initialization bug** — may work correctly in VS Code Continue but fail
   in the JetBrains plugin variant
3. **Model not selected from dropdown** — the dropdown shows only cloud models (see
   [[notes/pitfall/issue-continue-dev-intellij-dropdown-shows-only-cloud-models]]); local model may need explicit selection

## Investigation Steps

- Check Continue plugin logs: IntelliJ → Help → Show Log in Finder
- Test VS Code Continue with same `config.json` to isolate IntelliJ vs. general issue
- Check if Continue hub authentication intercepts local model requests when signed in

## Sources

- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related

- [[notes/pitfall/issue-continue-dev-intellij-dropdown-shows-only-cloud-models]]
- [[notes/concept/continue-config-json-continue-dev-configuration]]
- [[notes/concept/llama-server-sh-llama-cpp-startup-script]]
- [[notes/concept/open-questions-unresolved-topics-and-open-threads]]
