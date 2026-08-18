---
title: "Thinking Mode — Qwen3 reasoning_content Issue"
slug: thinking-mode-qwen3-reasoning-content-issue
type: concept
stack: [qwen3, llama-cpp, continue-dev]
tags: [llm, reasoning]
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

# Thinking Mode — Qwen3 reasoning_content Issue

Qwen3 models have thinking mode **enabled by default**. When thinking mode is active, the model
outputs its chain-of-thought into the `reasoning_content` field of the response, leaving the
`content` field **empty**.

## Problem
Any client that reads only `content` (including Continue.dev) receives an empty response.
No error is raised — the response is technically valid JSON with an empty content string.

## Fix
Pass `--reasoning off` to llama-server at startup. This forces all output into `content`.

## Verification
Must verify the flag is active in the **running process** — a manual server restart can bypass
the startup script and launch without the flag.

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/decision/decision-disable-qwen3-thinking-mode-permanently]]
- [[notes/pitfall/issue-qwen3-blank-response-in-continue-dev]]
- [[notes/concept/llama-server-sh-llama-cpp-startup-script]]
