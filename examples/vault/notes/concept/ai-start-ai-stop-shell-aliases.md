---
title: "ai-start / ai-stop — Shell Aliases"
slug: ai-start-ai-stop-shell-aliases
type: concept
stack: [shell, llama-cpp, litellm]
tags: [local-ai, rag]
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

# ai-start / ai-stop — Shell Aliases

Defined in `~/.zshrc`. Start and stop the full local AI stack:

- `ai-start` — launches llama.cpp server (port 8079) + LiteLLM proxy (port 4000) + RAG server (port 5001)
- `ai-stop` — terminates all three services

## Stack Ports
| Service | Port |
|---------|------|
| llama.cpp (llama-server) | 8079 |
| LiteLLM proxy | 4000 |
| RAG server | 5001 |

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/concept/llama-server-sh-llama-cpp-startup-script]]
- [[notes/concept/litellm-yaml-litellm-proxy-configuration]]
- [[notes/concept/rag-server-port-5001]]
- [[notes/concept/local-ai-stack-architecture-overview]]
