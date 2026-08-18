---
title: "Decision: Continue.dev Points Directly at llama.cpp (Port 8079)"
slug: decision-continue-dev-points-directly-at-llama-cpp-port-8079
type: decision
stack: [continue-dev, llama-cpp, litellm]
tags: [decision]
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

# Decision: Continue.dev Points Directly at llama.cpp (Port 8079)

**Decision:** `apiBase` in Continue config set to `http://localhost:8079/v1` (llama.cpp directly),
bypassing the LiteLLM proxy on port 4000.

**Rationale:** During debugging, reducing moving parts helped isolate issues. Direct connection
eliminates LiteLLM as a potential failure point for Continue specifically.

**Trade-off:** Other clients may still route through LiteLLM. The `ai-start` stack still runs
all three services.

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/concept/continue-config-json-continue-dev-configuration]]
- [[notes/concept/litellm-yaml-litellm-proxy-configuration]]
- [[notes/concept/llama-server-sh-llama-cpp-startup-script]]
- [[notes/concept/dev-tools-continue-dev-and-intellij-integration-summary]]
