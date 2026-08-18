---
title: "llama-server.sh — llama.cpp Startup Script"
slug: llama-server-sh-llama-cpp-startup-script
type: concept
stack: [llama-cpp, qwen3, shell]
tags: [local-ai]
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

# llama-server.sh — llama.cpp Startup Script

Shell script that starts the llama.cpp inference server serving Qwen3-4B-Q4_K_M.gguf on port 8079.

Key flag: `--reasoning off` — disables Qwen3 thinking mode so responses land in `content` (not `reasoning_content`). Without this flag, Continue.dev receives empty content.

Port: **8079**
Model: `Qwen3-4B-Q4_K_M.gguf`

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/concept/ai-start-ai-stop-shell-aliases]]
- [[notes/decision/decision-disable-qwen3-thinking-mode-permanently]]
- [[notes/concept/thinking-mode-qwen3-reasoning-content-issue]]
- [[notes/concept/local-ai-stack-architecture-overview]]
