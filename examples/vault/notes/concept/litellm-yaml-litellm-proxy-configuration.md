---
title: "litellm.yaml — LiteLLM Proxy Configuration"
slug: litellm-yaml-litellm-proxy-configuration
type: concept
stack: [litellm]
tags: [proxy, api, local-ai]
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

# litellm.yaml — LiteLLM Proxy Configuration

Located at `~/ai-assistant/litellm.yaml`. Configures the LiteLLM proxy that sits between clients
and llama.cpp. Runs on port **4000**.

Note: Continue.dev was eventually pointed directly at llama.cpp `:8079` to eliminate LiteLLM as a
variable during debugging. LiteLLM remains part of the `ai-start` stack for other clients.

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/concept/llama-server-sh-llama-cpp-startup-script]]
- [[notes/concept/ai-start-ai-stop-shell-aliases]]
- [[notes/decision/decision-continue-dev-points-directly-at-llama-cpp-port-8079]]
- [[notes/concept/local-ai-stack-architecture-overview]]
