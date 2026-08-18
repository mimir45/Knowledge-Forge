---
title: "LLM-Wiki Pattern — LLM-Maintained Knowledge Archive"
slug: llm-wiki-pattern-llm-maintained-knowledge-archive
type: concept
stack: [obsidian]
tags: [llm-wiki, knowledge-management, architecture]
depth: 3
confidence: low
created: 2026-04-17
updated: 2026-04-17
verified: 2026-04-17
freshness_days: 365
sources:
  - url: sources/daily/2026-04-17-storybook-llm-wiki.md
    accessed: 2026-04-17
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# LLM-Wiki Pattern — LLM-Maintained Knowledge Archive

A persistent, LLM-maintained wiki that prevents knowledge from dying in chat history.

## Three-Layer Architecture
1. **raw/** — immutable original sources (symlinks only)
2. **wiki pages** — processed, cross-referenced pages (sources, entities, concepts, decisions, issues, syntheses)
3. **schema** — index.md, log.md, CLAUDE.md

## Three Operations
- **INGEST** — read raw source → extract entities/decisions/issues/concepts → write pages → update log + index
- **QUERY** — read index → navigate to relevant pages → synthesize → cite sources → file back
- **LINT** — scan for contradictions, orphan pages, one-way links, concept gaps

## Key Differentiator vs RAG
The wiki is a **compiled, always-maintained artifact** — cross-references and contradictions are
pre-resolved at ingest time, not at query time. Filed-back answers are what prevent knowledge
from dying in chat history.

## Critical File
`CLAUDE.md` in the vault root transforms the agent from a general chatbot into a disciplined
wiki maintainer. Must include prohibitions: never write to `raw/`, no unsourced claims, no
deletion (archive instead), contradictions marked explicitly.

## Sources
- [[sources/daily/2026-04-17-storybook-llm-wiki]]

## Related
- [[notes/concept/open-questions-unresolved-topics-and-open-threads]]
