---
title: "Issue: Qwen3 Blank Response in Continue.dev"
slug: issue-qwen3-blank-response-in-continue-dev
type: pitfall
stack: [qwen3, continue-dev, llama-cpp]
tags: [llm, issue]
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

# Issue: Qwen3 Blank Response in Continue.dev

## Symptom
Continue.dev chat returns blank responses when using Qwen3-4B via llama.cpp.

## Root Cause
Qwen3 thinking mode is enabled by default. It outputs the chain-of-thought to `reasoning_content`
and leaves `content` empty. Continue.dev reads only `content`.

## Fix
Add `--reasoning off` to llama-server.sh startup command.

## Affected Files
- `~/ai-assistant/llama-server.sh`

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/concept/thinking-mode-qwen3-reasoning-content-issue]]
- [[notes/decision/decision-disable-qwen3-thinking-mode-permanently]]
- [[notes/concept/llama-server-sh-llama-cpp-startup-script]]
