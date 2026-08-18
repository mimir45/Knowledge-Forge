---
title: "Decision: Disable Qwen3 Thinking Mode Permanently"
slug: decision-disable-qwen3-thinking-mode-permanently
type: decision
stack: [llama-cpp, qwen3, continue-dev]
tags: [decision]
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

# Decision: Disable Qwen3 Thinking Mode Permanently

**Decision:** Add `--reasoning off` flag permanently to `llama-server.sh`.

**Rationale:** Qwen3 thinking mode is on by default. It outputs reasoning to `reasoning_content`
and leaves `content` empty. Continue.dev reads only `content` and returns blank responses.
Any client that doesn't explicitly read `reasoning_content` is broken by thinking mode.

**Trade-off:** Chain-of-thought reasoning is disabled. For coding assistance tasks where
Continue reads the response, this is acceptable.

## Sources
- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related
- [[notes/concept/thinking-mode-qwen3-reasoning-content-issue]]
- [[notes/pitfall/issue-qwen3-blank-response-in-continue-dev]]
- [[notes/concept/llama-server-sh-llama-cpp-startup-script]]
