---
title: "Local AI Stack — Architecture Overview"
slug: local-ai-stack-architecture-overview
type: concept
stack: [llama-cpp, litellm, continue-dev]
tags: [rag, local-ai, architecture]
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

# Local AI Stack — Architecture Overview

Complete local AI coding assistant stack: llama.cpp + LiteLLM + RAG server + Continue.dev.

## Service Map

| Service | Port | Description |
|---------|------|-------------|
| llama.cpp (llama-server) | 8079 | Inference server; serves Qwen3-4B-Q4_K_M.gguf |
| LiteLLM proxy | 4000 | OpenAI-compatible proxy layer for multiple clients |
| RAG server | 5001 | Retrieval-augmented generation; BGE-small embeddings |

## Startup / Shutdown
```bash
ai-start   # starts all three services (defined in ~/.zshrc)
ai-stop    # stops all three services
```

## llama.cpp Configuration
- Model: `Qwen3-4B-Q4_K_M.gguf`
- Critical flag: `--reasoning off` — disables thinking mode (Qwen3 default)
  Without this, `content` is empty and all clients receive blank responses
- Script: `~/ai-assistant/llama-server.sh`

## LiteLLM
- Config: `~/ai-assistant/litellm.yaml`
- Port: 4000
- Continue.dev is configured to bypass LiteLLM and connect directly to llama.cpp at `:8079`

## RAG Server
- Port: 5001
- Route: `POST /` (Continue sends to root path)
- Request field: `fullInput` (not `query` — `query` is always empty)
- Model: `BAAI/bge-small-en-v1.5`, eager loaded at startup
- Commands: `rag-index` (index codebase), `rag-query` (standalone query)

## Continue.dev Integration
See [[notes/concept/dev-tools-continue-dev-and-intellij-integration-summary]] for full Continue.dev configuration details.

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/concept/llama-server-sh-llama-cpp-startup-script]]
- [[notes/concept/litellm-yaml-litellm-proxy-configuration]]
- [[notes/concept/rag-server-port-5001]]
- [[notes/concept/ai-start-ai-stop-shell-aliases]]
- [[notes/concept/continue-config-json-continue-dev-configuration]]
- [[notes/concept/thinking-mode-qwen3-reasoning-content-issue]]
- [[notes/concept/rag-provider-continue-dev-rag-context-provider-protocol]]
- [[notes/decision/decision-disable-qwen3-thinking-mode-permanently]]
- [[notes/decision/decision-continue-dev-points-directly-at-llama-cpp-port-8079]]
- [[notes/decision/decision-rag-server-handles-post-using-fullinput-field]]
- [[notes/decision/decision-rag-embedding-model-loaded-eagerly-at-startup]]
