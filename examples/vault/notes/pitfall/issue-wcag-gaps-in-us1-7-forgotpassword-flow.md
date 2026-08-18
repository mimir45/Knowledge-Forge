---
title: "Issue: WCAG Gaps in US1.7 ForgotPassword Flow"
slug: issue-wcag-gaps-in-us1-7-forgotpassword-flow
type: pitfall
stack: [react]
tags: [accessibility, wcag, leprecoin, frontend]
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

# Issue: WCAG Gaps in US1.7 ForgotPassword Flow

## Status

**Not implemented**

## Gaps Identified

1. **`aria-live` on password strength indicator** — the strength meter updates visually but is
   not announced to screen readers; needs a debounced `aria-live="polite"` region
2. **Per-step focus management** — when the multi-step form advances to a new step, focus is
   not moved to the new step's first interactive element
3. **`aria-disabled` on cooldown resend button** — the resend button during cooldown uses only
   CSS styling; needs `aria-disabled="true"` to signal disabled state to assistive technology

## Sources

- [[sources/daily/2026-04-17-storybook-llm-wiki]]

## Related

- [[notes/concept/leprecoin-react-next-js-frontend-project]]
- [[notes/concept/open-questions-unresolved-topics-and-open-threads]]
