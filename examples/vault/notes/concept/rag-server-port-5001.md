---
title: "RAG Server — Port 5001"
slug: rag-server-port-5001
type: concept
stack: [continue-dev, bge]
tags: [rag, embeddings, local-ai]
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

# RAG Server — Port 5001

Local RAG (Retrieval-Augmented Generation) server wired into Continue.dev as a context provider.

- Port: **5001**
- Route: `POST /` (Continue sends to root, not `/retrieve`)
- Embedding model: `BAAI/bge-small-en-v1.5` (BGE small)
- Model load: eager at startup (not lazy on first request)
- Request field: `fullInput` must be used — `query` field is always empty in Continue's POST body

## Continue Protocol
Continue's `@rag` context provider sends:
```json
{ "query": "", "fullInput": "<actual query text>" }
```
Server must read `fullInput` and strip trailing whitespace.

## Notes
- `UNEXPECTED: embeddings.position_ids` warning on startup is harmless
- ChromaDB semaphore warning on shutdown is harmless
- Error logging added to `retrieve()` to produce tracebacks instead of silent drops

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/decision/decision-rag-server-handles-post-using-fullinput-field]]
- [[notes/decision/decision-rag-embedding-model-loaded-eagerly-at-startup]]
- [[notes/concept/rag-provider-continue-dev-rag-context-provider-protocol]]
- [[notes/concept/ai-start-ai-stop-shell-aliases]]
- [[notes/concept/local-ai-stack-architecture-overview]]
