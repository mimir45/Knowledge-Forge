---
title: "leprecoin — React/Next.js Frontend Project"
slug: leprecoin-react-next-js-frontend-project
type: concept
stack: [react, nextjs, typescript, storybook, mui]
tags: [fsd]
depth: 3
confidence: low
created: 2026-04-21
updated: 2026-04-21
verified: 2026-04-21
freshness_days: 365
sources:
  - url: sources/daily/2026-04-21-react-resetpassword-refactor.md
    accessed: 2026-04-21
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# leprecoin — React/Next.js Frontend Project

Next.js/React TypeScript frontend following Feature-Sliced Design (FSD) architecture.
Uses MUI, Redux, react-hook-form, and Storybook.

## Key Conventions

- FSD layers: `app`, `pages`, `widgets`, `features`, `entities`, `shared`
- Shared hooks live in `src/shared/lib/hooks/`
- Story decorators: `Provider + ToastProvider` for Redux forms; `SignUpContextProvider` for OTP flows
- `play` functions use `userEvent` for interaction demos in stories

## Storybook Coverage

| Date | Coverage | Notes |
|------|----------|-------|
| 2026-04-22 | 76.7% (33/43) | +9 stories from audit |
| Pre-audit | 55.8% (24/43) | Baseline |

## Testing Notes

- MUI `Button` with `disabled={true}` sets actual HTML `<button disabled>` — tests must use
  `toBeDisabled()` instead of attempting to click
- `passwordMinLength = 9` — verify test passwords against actual constants
- `npx tsc --noEmit` for TypeScript check (not in `package.json` scripts)

## Sources

- [[sources/daily/2026-04-21-react-resetpassword-refactor]]
- [[sources/daily/2026-04-22-storybook-coverage-audit]]

## Related

- [[notes/concept/usefocusonmount-shared-react-focus-hook]]
- [[notes/concept/storybook-decorator-pattern-provider-wrappers-for-redux-and-context]]
- [[notes/concept/mui-disabled-state-html-disabled-vs-aria-disabled]]
- [[notes/decision/decision-extract-usefocusonmount-hook]]
- [[notes/decision/decision-rewrite-tests-to-assert-tobedisabled-not-click]]
- [[notes/decision/decision-skip-storybook-stories-for-page-level-wrapper-components]]
