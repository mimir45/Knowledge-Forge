---
title: "Decision: Rewrite Tests to Assert toBeDisabled() Not Click"
slug: decision-rewrite-tests-to-assert-tobedisabled-not-click
type: decision
stack: [react, mui]
tags: [testing, leprecoin]
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

# Decision: Rewrite Tests to Assert toBeDisabled() Not Click

## Decision

When a submit button is legitimately disabled for invalid form states, tests must assert
`expect(button).toBeDisabled()` instead of attempting to click it.

## Context

Adding `isDisabled={isFormInvalid}` to the `ResetPasswordForm` submit button caused 6 existing
tests to fail: they tried to click the button with invalid data, triggering a
`pointer-events: none` error from `@testing-library/user-event`.

## Rationale

- The tests were asserting that submitting with invalid data shows an error — but the correct
  UX is that the button is unreachable; the test should reflect actual behavior
- One test ("should clear root error on password focus") was removed entirely: the root error is
  only settable via `onSubmit`, which is blocked when the button is disabled — unreachable state

## Sources

- [[sources/daily/2026-04-21-react-resetpassword-refactor]]

## Related

- [[notes/concept/mui-disabled-state-html-disabled-vs-aria-disabled]]
- [[notes/concept/leprecoin-react-next-js-frontend-project]]
