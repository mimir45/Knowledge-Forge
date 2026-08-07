---
title: "RAG server notes (superseded)"
tags: [rag, archive]
source: raw/daily/2026-04-13.md
date: 2026-04-13
status: archived
---

# RAG server notes (superseded)

Superseded by the local-AI stack write-up. Kept for history only.

The RAG server listened on :5001 and took `fullInput` rather than `query` on `POST /`.
Indexing used `BAAI/bge-small-en-v1.5`; the `embeddings.position_ids` warning at load
was harmless and could be ignored.
