---
title: "Decision: Skip Storybook Stories for Page-Level Wrapper Components"
slug: decision-skip-storybook-stories-for-page-level-wrapper-components
type: decision
stack: [storybook, react]
tags: [coverage, leprecoin]
depth: 3
confidence: high
created: 2026-04-22
updated: 2026-04-22
verified: 2026-04-22
freshness_days: 730
sources:
  - url: sources/daily/2026-04-22-storybook-coverage-audit.md
    accessed: 2026-04-22
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Decision: Skip Storybook Stories for Page-Level Wrapper Components

## Decision

Full-page wrapper components (`LoginPage`, `SignUpPage`, `WelcomePage`, etc.) are intentionally
excluded from Storybook coverage. They count as uncovered in coverage metrics but are not
treated as gaps.

## Context

During the 2026-04-22 coverage audit, 7 page-level wrappers were identified without stories.
Coverage goal was 76.7% (33/43) — these 7 were left out.

## Rationale

- Page wrappers only compose already-storied child components
- They contain no isolated visual logic of their own
- Storying them would duplicate the children's stories with no additional visual signal
- Maintenance cost doubles with no coverage benefit

## Sources

- [[sources/daily/2026-04-22-storybook-coverage-audit]]

## Related

- [[notes/concept/storybook-decorator-pattern-provider-wrappers-for-redux-and-context]]
- [[notes/concept/leprecoin-react-next-js-frontend-project]]
