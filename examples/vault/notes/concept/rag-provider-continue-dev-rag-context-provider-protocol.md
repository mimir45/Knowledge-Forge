---
title: "RAG Provider — Continue.dev @rag Context Provider Protocol"
slug: rag-provider-continue-dev-rag-context-provider-protocol
type: concept
stack: [continue-dev]
tags: [rag, protocol, embeddings]
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

# RAG Provider — Continue.dev @rag Context Provider Protocol

Continue.dev's `@rag` context provider sends a POST request to the configured RAG server URL.

## Protocol Details
- Method: `POST`
- Path: `/` (root — NOT `/retrieve` or any other path)
- Body:
```json
{
  "query": "",
  "fullInput": "<actual query text>"
}
```

The `query` field is **always empty**. The actual query text is in `fullInput`.
Server must read `fullInput` and strip trailing whitespace before embedding.

## Server Requirements
- Accept any POST path (not strict route matching)
- Return an array of context items
- Load embedding model eagerly at startup for low latency

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/concept/rag-server-port-5001]]
- [[notes/decision/decision-rag-server-handles-post-using-fullinput-field]]
- [[notes/decision/decision-rag-embedding-model-loaded-eagerly-at-startup]]
- [[notes/concept/local-ai-stack-architecture-overview]]
