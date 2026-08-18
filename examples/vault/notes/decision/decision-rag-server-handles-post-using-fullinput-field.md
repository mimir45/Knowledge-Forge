---
title: "Decision: RAG Server Handles POST / Using fullInput Field"
slug: decision-rag-server-handles-post-using-fullinput-field
type: decision
stack: [continue-dev]
tags: [rag, protocol, decision]
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

# Decision: RAG Server Handles POST / Using fullInput Field

**Decision:** RAG server accepts any POST path (root `/`) and reads `fullInput` from the request body.

**Rationale:** Continue.dev `@rag` context provider POSTs to `/` (not `/retrieve`). The `query`
field is always empty — the actual text is in `fullInput`.

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/concept/rag-server-port-5001]]
- [[notes/concept/rag-provider-continue-dev-rag-context-provider-protocol]]
