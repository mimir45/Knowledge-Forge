---
title: "Decision: RAG Embedding Model Loaded Eagerly at Startup"
slug: decision-rag-embedding-model-loaded-eagerly-at-startup
type: decision
tags: [rag, performance, decision]
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

# Decision: RAG Embedding Model Loaded Eagerly at Startup

**Decision:** BGE-small embedding model is loaded when the RAG server starts, not on the first
request.

**Rationale:** Lazy loading causes high latency on the first `@rag` query in Continue. Eager
loading ensures instant responses from the first query in an IntelliJ session.

**Trade-off:** Slightly longer startup time for the RAG server.

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/concept/rag-server-port-5001]]
- [[notes/concept/ai-start-ai-stop-shell-aliases]]
