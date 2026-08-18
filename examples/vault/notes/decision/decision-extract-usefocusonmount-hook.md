---
title: "Decision: Extract useFocusOnMount Hook"
slug: decision-extract-usefocusonmount-hook
type: decision
stack: [react]
tags: [hooks, refactor, leprecoin]
depth: 3
confidence: high
created: 2026-04-21
updated: 2026-04-21
verified: 2026-04-21
freshness_days: 730
sources:
  - url: sources/daily/2026-04-21-react-resetpassword-refactor.md
    accessed: 2026-04-21
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Decision: Extract useFocusOnMount Hook

## Decision

Extract the repeated `useRef + useEffect` focus-on-mount pattern into a shared hook at
`src/shared/lib/hooks/useFocusOnMount.ts`.

## Context

The same two-line focus pattern appeared in `CheckEmail`, `ForgotPasswordForm`, and
`ResetPasswordForm`. The `/simplify` skill pass identified this as copy-paste boilerplate.

## Rationale

- Eliminates three identical code blocks
- Centralises the `React.RefObject<T | null>` typing quirk in one place
- Keeps components focused on form logic, not focus management

## Sources

- [[sources/daily/2026-04-21-react-resetpassword-refactor]]

## Related

- [[notes/concept/usefocusonmount-shared-react-focus-hook]]
- [[notes/concept/leprecoin-react-next-js-frontend-project]]
