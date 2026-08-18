---
title: "Issue: Continue.dev IntelliJ Dropdown Shows Only Cloud Models"
slug: issue-continue-dev-intellij-dropdown-shows-only-cloud-models
type: pitfall
stack: [continue-dev, intellij]
tags: [local-ai, unresolved]
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

# Issue: Continue.dev IntelliJ Dropdown Shows Only Cloud Models

## Status

**Unresolved**

## Symptom

The "Qwen3 4B (local)" label appears in the top-left corner of the Continue chat panel, but
clicking the model dropdown shows only Claude and Gemini cloud models. The custom local model
is not selectable from the dropdown.

## Root Cause Hypothesis

Continue.dev 0.9.264 introduced a hub-based model selector. Models from `config.json` may not
surface in the UI dropdown even when the config is otherwise valid.

## Investigation Steps

- Verify Continue.dev version; check if newer versions fixed hub model selector sync
- Check if the model appears after restarting IntelliJ with the plugin freshly loaded

## Sources

- [[sources/daily/2026-04-13-local-ai-continue-rag-spring]]

## Related

- [[notes/pitfall/issue-continue-dev-sends-zero-requests-to-llama-cpp]]
- [[notes/concept/continue-config-json-continue-dev-configuration]]
- [[notes/concept/open-questions-unresolved-topics-and-open-threads]]
