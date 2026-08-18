---
title: "Issue: 11 Components Missing Storybook Stories"
slug: issue-11-components-missing-storybook-stories
type: pitfall
stack: [storybook, react]
tags: [coverage, leprecoin]
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

# Issue: 11 Components Missing Storybook Stories

## Status

**Partially resolved** — coverage raised to 76.7% (33/43) on 2026-04-22. Remaining gap is
the `widgets/` layer at 50% and any components added since the audit.

## Context

As of 2026-04-17, 11 components listed in the status doc had no Storybook stories.
The 2026-04-22 audit added 9 stories (55.8% → 76.7%). The remaining ~10 uncovered components
are page-level wrappers excluded by [[notes/decision/decision-skip-storybook-stories-for-page-level-wrapper-components]] or `widgets/`
components with no isolated visual logic yet.

## Sources

- [[sources/daily/2026-04-17-storybook-llm-wiki]]
- [[sources/daily/2026-04-22-storybook-coverage-audit]]

## Related

- [[notes/concept/leprecoin-react-next-js-frontend-project]]
- [[notes/decision/decision-skip-storybook-stories-for-page-level-wrapper-components]]
- [[notes/concept/open-questions-unresolved-topics-and-open-threads]]
